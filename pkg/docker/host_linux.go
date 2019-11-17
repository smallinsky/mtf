package docker

import (
	"os/exec"
	"strings"
)

// HostIP returns host ip that allows to reach host directly inside docker container.
func HostIP() (string, error) {
	cmd := strings.Join([]string{
		`ip addr show scope global dev docker0`,
		`grep inet`,
		`tr -s " "`,
		`cut -d " " -f 3`,
		`cut -d / -f1`,
		`tr -d '\n'`,
	}, "|")

	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return "", err
	}
	return string(out), err
}
