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

// SutConfig is a configuration required to buld and run sut
// into docker container.
type SutConfig struct {
	// Path to dir binary source files that will be
	// build and executed into docker containers as a SUT.
	Path string
	// Env is list of environment variables will be passed to SUT.
	Env []string
	// ExposedPorts is a list of port that will be exposed and forwarded
	// to docker host.
	ExposedPorts []int
	// Mounts allows to pass list of directories which should be mountet
	// inside system under test container in form 'src:dst'.
	Mounts []string
	// RuntimeTypeCommand allows to distinguish between service and simple command binary.
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

	customMounts := []docker.Mount{}
	for _, m := range config.Mounts {
		src, dst, err := parseMount(m)
		if err != nil {
			return nil, err
		}
		customMounts = append(customMounts, docker.Mount{
			Source: src,
			Target: dst,
		})
	}

	var waitPolicy docker.WaitPolicy
	if !config.RuntimeTypeCommand {
		waitPolicy = &docker.WaitForProcess{Process: config.binaryName}
	}

	mounts := []docker.Mount{
		certMount,
		binaryMount,
	}
	mounts = append(mounts, customMounts...)
	return &docker.ContainerConfig{
		Name:        name,
		Image:       image,
		Env:         env,
		Mounts:      docker.Mounts(mounts),
		PortMap:     ports,
		NetworkName: network,
		Privileged:  true,
		WaitPolicy:  waitPolicy,
	}, nil
}

func parseMount(in string) (source, destination string, err error) {
	sd := strings.Split(in, ":")
	if len(sd) == 2 {
		return sd[0], sd[1], nil
	}
	return "", "", fmt.Errorf("invalid mount format: got %s, expected <src>:<dst>", in)
}
