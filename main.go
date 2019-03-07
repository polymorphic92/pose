package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
)

var envMap = make(map[string]string)

//
// strucs for yaml config file
//

type openshiftBackend struct {
	Endpoint      string
	Namespace     string
	Fieldselector map[string]string
	Mapping       map[string]map[string]string
}

type workProject struct {
	Inline    map[string]string
	Openshift []openshiftBackend
}

type poseConfig struct {
	Projects map[string]workProject
}

//
// struct for openshift json response
//

type openshiftSecert struct {
	Items []struct {
		Metadata struct {
			Name string
		}
		Data map[string]string
	}
}

//###############################################
//###############################################

func main() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)

	var currentProject = readConfigFile()

	if inlineMap := currentProject.Inline; inlineMap != nil {
		addInLineMapping(inlineMap)
	}

	if openshiftBackends := currentProject.Openshift; openshiftBackends != nil {
		addOpenshiftMapping(openshiftBackends)
	}

	setEnvs(envMap)
	runDockerCompose()

}

func readConfigFile() workProject {

	dat, err := ioutil.ReadFile(os.Getenv("HOME") + "/pose-config.yml")
	if err != nil {
		panic(err) // instead of error create config in $USER home dir
	}

	projectPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	projecBasetPath := filepath.Base(projectPath)

	var config poseConfig

	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		panic(err)
	}
	return config.Projects[projecBasetPath]

}

func addInLineMapping(inlineMap map[string]string) {
	for key, value := range inlineMap {
		envMap[key] = value
	}
}

func addOpenshiftMapping(openshiftArr []openshiftBackend) {
	var token = getOpenshiftToken()
	for _, openshiftObj := range openshiftArr {

		var req = buildOpenshiftRequest(openshiftObj, token)
		var secertsObj = getOpenshiftSecert(req)

		for _, item := range secertsObj.Items {
			var secretMap = openshiftObj.Mapping[item.Metadata.Name]
			if secretMap != nil {
				for envName, secretPart := range secretMap {
					envMap[envName] = base64Decode(item.Data[string(secretPart)])
				}
			}
		}

	}
}

func getOpenshiftSecert(req *http.Request) openshiftSecert {

	openshiftClient := http.Client{}

	res, getErr := openshiftClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	secertsObj := openshiftSecert{}

	jsonErr := json.Unmarshal(body, &secertsObj)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return secertsObj
}

func buildOpenshiftRequest(openshiftObj openshiftBackend, token string) *http.Request {

	var requestURL strings.Builder

	requestURL.WriteString("https://" + openshiftObj.Endpoint)
	requestURL.WriteString("/api/v1/namespaces/" + openshiftObj.Namespace + "/secrets")

	var fSelectors strings.Builder

	for field, fieldValue := range openshiftObj.Fieldselector {
		fSelectors.WriteString(field + "=" + fieldValue)
	}

	if len(fSelectors.String()) > 0 {
		requestURL.WriteString("?fieldSelector=" + fSelectors.String())
	}

	req, err := http.NewRequest(http.MethodGet, requestURL.String(), nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func setEnvs(m map[string]string) {
	for key, value := range m {
		os.Setenv(key, value)
	}

}

func runDockerCompose() {
	compose := "docker-compose"

	if cmdExists(compose) {
		cmd := exec.Command(compose, os.Args[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
}

func getOpenshiftToken() string {
	args := [3]string{"oc", "whoami", "-t"}
	out, err := exec.Command(args[0], args[1:3]...).Output()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err) // need to  show real error message
	}
	return strings.TrimSuffix(string(out), "\n")
}

func cmdExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	if err == nil {
		return true
	}
	return false
}

func base64Decode(str string) string {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return ""
	}
	return string(data)
}
