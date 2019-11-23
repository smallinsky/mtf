package sut

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/smallinsky/mtf/framework/core"
	"github.com/smallinsky/mtf/pkg/build"
	"github.com/smallinsky/mtf/pkg/docker"
)

type SutConfig struct {
	Path               string
	Env                []string
	ExposedPorts       []int
	RuntimeTypeCommand bool

	absoltePath string
	binaryName  string
}

func (c *SutConfig) Build() error {
	stat, err := os.Stat(c.Path)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return fmt.Errorf("path is not directory")
	}
	c.absoltePath, err = filepath.Abs(c.Path)
	if err != nil {
		return err
	}
	path := strings.Split(c.absoltePath, `/`)
	c.binaryName = path[len(path)-1]
	return nil
}

func BuildContainerConfig(config SutConfig) (*docker.ContainerConfig, error) {
	if err := config.Build(); err != nil {
		return nil, err
	}
	if core.Settings.BuildBinary {
		if err := build.Build(config.absoltePath); err != nil {
			return nil, fmt.Errorf("failed to build sut binary from %s, err %v", config.absoltePath, err)
		}
	}

	ports := make(map[docker.ContainerPort]docker.HostPort)
	for _, v := range config.ExposedPorts {
		ports[docker.ContainerPort(v)] = docker.HostPort(v)
	}

	var (
		image   = "smallinsky/run_sut"
		name    = "sut_mtf"
		network = "mtf_net"
	)

	env := append(config.Env, fmt.Sprintf("SUT_BINARY_NAME=%s", config.binaryName))

	certMount := docker.Mount{
		Source: "/tmp/mtf/cert",
		Target: "/usr/local/share/ca-certificates",
	}

	binaryMount := docker.Mount{
		Source: config.absoltePath,
		Target: "/component",
	}
	var waitPolicy docker.WaitPolicy
	if !config.RuntimeTypeCommand {
		waitPolicy = &docker.WaitForProcess{Process: config.binaryName}
	}

	return &docker.ContainerConfig{
		Name:  name,
		Image: image,
		Env:   env,
		Mounts: docker.Mounts{
			certMount,
			binaryMount,
		},
		PortMap:     ports,
		NetworkName: network,
		Privileged:  true,
		WaitPolicy:  waitPolicy,
	}, nil
}
