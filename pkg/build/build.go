package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/smallinsky/mtf/pkg/exec"
)

func Build(path string) error {
	var err error
	if path, err = filepath.Abs(path); err != nil {
		return errors.Wrapf(err, "failed to get abs path")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.Wrapf(err, "dir doesn't exist")
	}

	pwd, err := os.Getwd()
	if err != nil {
		return errors.Wrapf(err, "failed to get current dir")
	}

	if err := os.Chdir(path); err != nil {
		return errors.Wrapf(err, "failed to change working dir")
	}

	b := strings.Split(path, `/`)
	bin := b[len(b)-1]

	cmd := []string{
		"go", "build", "-o", fmt.Sprintf("%s/%s", path, bin), path,
	}

	if err := exec.Run(cmd, exec.WithEnv("CGO_ENABLED=0", "GODEBUG=x509ignoreCN=1", "GOOS=linux", "GOARCH=amd64", "GO111MODULE=on")); err != nil {
		return errors.Wrapf(err, "failed to run cmd")
	}

	if err := os.Chdir(pwd); err != nil {
		return errors.Wrapf(err, "failed to restore working dir")
	}
	return nil
}
