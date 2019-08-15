package sut

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/client"
	"github.com/pkg/errors"

	"github.com/smallinsky/mtf/framework/core"
	"github.com/smallinsky/mtf/pkg/docker"
	"github.com/smallinsky/mtf/pkg/exec"
)

type Config struct {
	Dir string
	Env []string
}

func NewSUT(path string, env ...string) *SUT {
	return &SUT{
		Path: path,
		Env:  env,
	}
}

type SUT struct {
	Config Config

	Path      string
	Env       []string
	start     time.Time
	container *docker.Container
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
		// TODO Get binary base on the path and repo name or if binary deosn't exist build it.
		// Add ability to run sut from existing image.
		binary = bin
		path   = c.Path
	)

	exec.Run([]string{
		"mkdir", "-p", "/tmp/mtf/cert",
	})

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	container, err := docker.NewContainer(cli, docker.Config{
		Name:     "sut_mtf",
		Image:    "run_sut",
		Hostname: "sut_mtf",
		CapAdd:   []string{"NET_RAW", "NET_ADMIN"},
		Labels: map[string]string{
			"mtf": "mtf",
		},
		Env: append([]string{
			fmt.Sprintf("SUT_BINARY_NAME=%s", binary),
			"ORACLE_ADDR=host.docker.internal:8002",
		}, c.Env...),
		PortMap: docker.PortMap{
			8001: 8001,
		},
		Mounts: docker.Mounts{
			docker.Mount{
				Source: path,
				Target: "/component",
			},
			docker.Mount{
				Source: "/tmp/mtf/cert",
				Target: "/usr/local/share/ca-certificates",
			},
		},
		NetworkName: "mtf_net",
		Healtcheck: &docker.Healtcheck{
			Test:     []string{"CMD", "pgrep", binary, "||", "exit", "1"},
			Interval: time.Millisecond * 100,
			Timeout:  time.Second * 1,
		},
	})
	if err != nil {
		return err
	}

	if err := container.Start(); err != nil {
		return err
	}

	c.container = container
	return nil
}

func join(args []string) string {
	return strings.Join(args, " ")
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

	if err := exec.Run(cmd, exec.WithEnv("GOOS=linux", "GOARCH=amd64")); err != nil {
		return errors.Wrapf(err, "failed to run cmd")
	}
	return nil
}

func (c *SUT) Ready() (err error) {
	defer func() {
		if err != nil {
			_ = c.container.Stop()
		}
	}()
	state, err := c.container.WaitForReady()
	if err != nil {
		return err
	}
	if state.ExitCode != 0 {
		logs, _ := c.container.Logs()
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to start container:\nExitCode: %v\nContainer logs: %s", state.ExitCode, logs)
	}
	return nil
}

func (c *SUT) Stop() error {
	return c.container.Stop()
}

func (c *SUT) StartPriority() int {
	return 5
}
