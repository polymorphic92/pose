package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
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

//
// openshift config aka ~/.kube/config
//
type openshiftConfig struct {
	Users []struct {
		Name string
		User map[string]string
	}
}

//###############################################
//###############################################

func main() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)

	var currentProject = readConfigFile()
	fields := reflect.TypeOf(readConfigFile())
	for i := 0; i < fields.NumField(); i++ {
		switch fields.Field(i).Name {
		case "Inline":
			addInLineMapping(currentProject.Inline)
		case "Openshift":
			addOpenshiftMapping(currentProject.Openshift)
		}
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
	var osConfig = getOpenshiftConfig()

	for _, openshiftObj := range openshiftArr {
		var token = getOpenshiftToken(osConfig, strings.Split(openshiftObj.Endpoint, ".")[0])
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

	requestURL.WriteString("https://" + openshiftObj.Endpoint + ":8443")
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

func getOpenshiftToken(config openshiftConfig, endpoint string) string {
	for _, item := range config.Users {
		if strings.Contains(item.Name, endpoint) {
			if token, ok := item.User["token"]; ok {
				return token
			}
		}
	}
	return ""
}

func getOpenshiftConfig() openshiftConfig {
	args := [5]string{"oc", "config", "view", "-o", "json"}
	out, err := exec.Command(args[0], args[1:5]...).Output()
	if err != nil {
		log.Fatalf("failed  getting openshift configwith %s\n", err)
	}

	osConfig := openshiftConfig{}

	jsonErr := json.Unmarshal(out, &osConfig)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return osConfig
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
