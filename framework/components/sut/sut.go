package sut

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/smallinsky/mtf/framework/core"
	"github.com/smallinsky/mtf/pkg/docker"
	"github.com/smallinsky/mtf/pkg/exec"
)

type SutConfig struct {
	Path string
	Env  []string
}

type SUT struct {
	cli       *docker.Client
	container *docker.Container

	config SutConfig
}

func NewSUT(cli *docker.Client, config SutConfig) (*SUT, error) {
	return &SUT{
		config: config,
		cli:    cli,
	}, nil
}

func (c *SUT) Start() error {
	var err error
	if c.config.Path, err = filepath.Abs(c.config.Path); err != nil {
		return fmt.Errorf("failed to get absolute path for %v path", c.config.Path)
	}

	if _, err := os.Stat(c.config.Path); os.IsNotExist(err) {
		return fmt.Errorf("path '%v' doesn't exist", c.config.Path)
	}

	b := strings.Split(c.config.Path, `/`)
	bin := b[len(b)-1]

	if core.Settings.BuildBinary {
		if err := BuildGoBinary(c.config.Path); err != nil {
			return fmt.Errorf("failed to build sut binary from %s, err %v", c.config, err)
		}
	}

	var (
		// TODO Get binary base on the path and repo name or if binary deosn't exist build it.
		// Add ability to run sut from existing image.
		binary = bin
		path   = c.config.Path
	)

	exec.Run([]string{
		"mkdir", "-p", "/tmp/mtf/cert",
	})

	result, err := c.cli.NewContainer(docker.Config{
		Name:     fmt.Sprintf("sut_mtf-%v", time.Now().Unix()),
		Image:    "run_sut",
		Hostname: "sut_mtf",
		CapAdd:   []string{"NET_RAW", "NET_ADMIN"},
		Labels: map[string]string{
			"mtf": "mtf",
		},
		Env: append([]string{
			fmt.Sprintf("SUT_BINARY_NAME=%s", binary),
		}, c.config.Env...),
		PortMap: docker.PortMap{
			8001: 8001,
			8082: 8082,
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

	c.container = result

	return c.container.Start()
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

	if err := exec.Run(cmd, exec.WithEnv("GOOS=linux", "GOARCH=amd64", "GO111MODULE=on")); err != nil {
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

func (c *SUT) Logs() ([]byte, error) {
	logs, err := c.container.Logs()
	if err != nil {
		return nil, err
	}
	return []byte(logs), nil
}

func (c *SUT) Stop() error {
	return c.container.Stop()
}

func (c *SUT) StartPriority() int {
	return 5
}
