package components

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/smallinsky/mtf/framework/core"
)

func NewSUT(path string, env ...string) *SUT {
	return &SUT{
		Path: path,
		Env:  env,
	}
}

type SUT struct {
	Path  string
	Env   []string
	start time.Time
}

func (c *SUT) Start() error {
	c.start = time.Now()

	var err error
	if c.Path, err = filepath.Abs(c.Path); err != nil {
		return fmt.Errorf("failed to get absolute path for %v path", c.Path)
	}
	if _, err := os.Stat(c.Path); os.IsNotExist(err) {
		return fmt.Errorf("path '%v' doesn't exist", c.Path)
	}

	b := strings.Split(c.Path, `/`)
	bin := b[len(b)-1]

	if core.Settings.BuildBinary {
		if err := BuildGoBinary(c.Path); err != nil {
			return fmt.Errorf("failed to build sut binary from %s, err %v", c.Path, err)
		}
	}

	var (
		name  = "sut"
		port  = "8001"
		image = "run_sut"
		// TODO Get binary base on the path and repo name or if binary deosn't exist build it.
		// Add ability to run sut from existing image.
		binary = bin
		path   = c.Path
	)

	runCmd([]string{
		"mkdir", "-p", "/tmp/mtf/cert",
	})

	cmd := []string{
		"docker", "run", "--rm", "-d",
		fmt.Sprintf("--name=%s_mtf", name),
		fmt.Sprintf("--hostname=%s_mtf", name),
		"--network=mtf_net",
		"-p", fmt.Sprintf("%s:%s", port, port),
		"--cap-add=NET_ADMIN",
		"--cap-add=NET_RAW",
		"-e", fmt.Sprintf("SUT_BINARY_NAME=%v", binary),
		envPlaceholder,
		"-v", fmt.Sprintf("%s:/component", path),
		"-v", "/tmp/mtf/cert:/usr/local/share/ca-certificates",
		image,
	}
	cmd = appendEnv(cmd, c.Env)
	fmt.Println("Run ", join(cmd))
	return runCmd(cmd)
}

const (
	envPlaceholder = "ENV_PLACEHOLDER"
)

func join(args []string) string {
	return strings.Join(args, " ")
}

func appendEnv(cmd, env []string) []string {
	var penv []string
	for _, s := range env {
		penv = append(penv, []string{"-e", s}...)
	}
	var out []string
	for i := range cmd {
		if cmd[i] == envPlaceholder {
			out = append(out, cmd[0:i]...)
			out = append(out, penv...)
			out = append(out, cmd[i+1:]...)
			return out
		}
	}
	return cmd
}

func BuildGoBinary(path string) error {
	var err error
	if path, err = filepath.Abs(path); err != nil {
		return errors.Wrapf(err, "failed to get abs path")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.Wrapf(err, "dir doesn't exist")
	}

	b := strings.Split(path, `/`)
	bin := b[len(b)-1]

	cmd := []string{
		"go", "build", "-o", fmt.Sprintf("%s/%s", path, bin), path,
	}

	if err := runCmd(cmd, WithEnv("GOOS=linux", "GOARCH=amd64")); err != nil {
		return errors.Wrapf(err, "failed to run cmd")
	}
	return nil
}

func (c *SUT) Ready() error {
	waitForPortOpen("localhost", "8001")
	// TODO sync sut start
	time.Sleep(time.Millisecond * 700)
	fmt.Printf("%T start time %v\n", c, time.Now().Sub(c.start))
	return nil
}

func (c *SUT) Stop() error {
	cmd := []string{
		"docker", "kill", fmt.Sprintf("%s_mtf", "sut"),
	}
	return runCmd(cmd)
}

func waitForPortOpen(host, port string) {
	firstRun := true
	for {
		if firstRun {
			firstRun = false
		} else {
			time.Sleep(time.Millisecond * 50)
		}
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), time.Millisecond*500)
		if err != nil {
			continue
		}

		conn.Close()
		return
	}
}
