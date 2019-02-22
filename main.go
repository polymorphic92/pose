package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
)

var envMap = make(map[string]string)

type openshiftBackend struct {
	clusterURL       string
	project          string
	secretEnvMapping map[string]map[string]string
}

type workProject struct {
	Inline    map[string]string
	Openshift []openshiftBackend
}

type poseConfig struct {
	Projects map[string]workProject
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)

	readConfigFile()
	setEnvs(envMap)
	runDockerCompose()

}

func readConfigFile() {

	dat, err := ioutil.ReadFile(os.Getenv("HOME") + "/pose-config.yml")
	if err != nil {
		panic(err) // instead of error create config in $USER home dir
	}

	projectPath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	projecBasetPath := filepath.Base(projectPath)
	// fmt.Println(projecBasetPath)

	var config poseConfig

	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		panic(err)
	}

	// fmt.Println(config)
	for key, value := range config.Projects[projecBasetPath].Inline {
		envMap[key] = value
	}

}

func setEnvs(m map[string]string) {
	for key, value := range m {

		// fmt.Println("KEY: " + key + " VALUE: " + value)
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

func ocWhoAmI() string {
	args := [3]string{"oc", "whoami", "-t"}
	out, err := exec.Command(args[0], args[1:3]...).Output()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	return string(out)
}

func cmdExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	if err == nil {
		return true
	}
	return false
}
