package components

import (
	"fmt"
	"os/exec"
	"strings"
)

func networkExists(name string) bool {
	return cmdExitStatus(fmt.Sprintf("docker inspect %s", name))
}

func containerIsRunning(name string) bool {
	return cmdExitStatus(fmt.Sprintf("docker top %s", name))
}

func cmdExitStatus(command string) bool {
	args := strings.Split(command, " ")
	cmd := exec.Command(args[0], args[1:len(args)]...)
	err := cmd.Start()
	if err != nil {
		return false
	}
	err = cmd.Wait()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return ee.ExitCode() == 0
		}
	}
	return true
}
