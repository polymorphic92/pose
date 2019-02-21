package main

import (
	"io/ioutil"
	"os"
	"os/exec"

	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
)

var envMap = make(map[string]string)

func main() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)

	readConfigFile()
	setEnvs(envMap)
	runDockerCompose()

}

func readConfigFile() {

	type Config struct {
		Foo string
		Bar []string
		Baz struct {
			Test1 string
			Test2 string
			Test3 string
		}
	}

	dat, err := ioutil.ReadFile(os.Getenv("HOME") + "/pose-config.yml")
	if err != nil {
		panic(err) // instead of error create config in $USER home dir
	}

	var config map[string]interface{}

	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		panic(err)
	}
	for key, value := range config {
		envMap[key] = value.(string)
	}

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
