package sut

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/smallinsky/mtf/framework/core"
	"github.com/smallinsky/mtf/pkg/build"
	"github.com/smallinsky/mtf/pkg/docker"
	"github.com/smallinsky/mtf/pkg/exec"
)

type SutConfig struct {
	Path         string
	Env          []string
	ExposedPorts []int
}

type SUT struct {
	cli       *docker.Docker
	container *docker.ContainerType
	config    SutConfig
}

func NewSUT(cli *docker.Docker, config SutConfig) *SUT {
	return &SUT{
		config: config,
		cli:    cli,
	}
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
		if err := build.Build(c.config.Path); err != nil {
			return fmt.Errorf("failed to build sut binary from %s, err %v", c.config.Path, err)
		}
	}

	var (
		binary = bin
		path   = c.config.Path
	)

	exec.Run([]string{
		"mkdir", "-p", "/tmp/mtf/cert",
	})

	ports := make(map[docker.ContainerPort]docker.HostPort)
	for _, v := range c.config.ExposedPorts {
		ports[docker.ContainerPort(v)] = docker.HostPort(v)
	}

	var (
		name     = fmt.Sprintf("sut_mtf-%v", time.Now().Unix())
		image    = "smallinsky/run_sut"
		hostname = "sut_mtf"
	)

	dockerConf := docker.ContainerConfig{
		Name:     name,
		Image:    image,
		Hostname: hostname,
		CapAdd:   []string{"NET_RAW", "NET_ADMIN"},
		Env: append([]string{
			fmt.Sprintf("SUT_BINARY_NAME=%s", binary),
		}, c.config.Env...),
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
		PortMap:     ports,
		NetworkName: "mtf_net",
		WaitPolicy:  &docker.WaitForProcess{Process: binary},
	}

	result, err := c.cli.NewContainer(dockerConf)
	if err != nil {
		return err
	}

	c.container = result

	return c.container.Start()
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
