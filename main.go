package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {

	program := "docker-compose"

	if cmdExists(program) {
		os.Setenv("FOO", "TESTING_ENV")

		cmd := exec.Command(program, os.Args[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Printf("%v\n", err)
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
