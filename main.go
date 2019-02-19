package main

import (
	_ "fmt"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func main() {
	// read config file
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)
	runDockerCompose()
}

func runDockerCompose() {
	compose := "docker-compose"

	if cmdExists(compose) {
		setEnvs()

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

func setEnvs() {
	os.Setenv("FOO", "TESTING_ENV")
}
