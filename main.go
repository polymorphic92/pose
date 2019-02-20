package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"gopkg.in/yaml.v2"

	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)

	readConfigFile()
	setEnvs()
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

	var config Config
	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		panic(err)
	}
	fmt.Printf(spew.Sdump(config))

}

func setEnvs() {
	os.Setenv("FOO", "TESTING_ENV")
}

func runDockerCompose() {
	compose := "docker-compose"

	if cmdExists(compose) {
		// setEnvs()

		cmd := exec.Command(compose, os.Args[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			log.WithFields(log.Fields{"Message": err}).Warn("Error while running docker-compose")
		}
	}
}

func cmdExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	if err == nil {
		return true
	}
	return false
}
