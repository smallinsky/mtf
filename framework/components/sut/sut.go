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

func BuildContainerConfig(config SutConfig) (*docker.ContainerConfig, error) {
	var err error
	if config.Path, err = filepath.Abs(config.Path); err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %v path", config.Path)
	}

	if _, err := os.Stat(config.Path); os.IsNotExist(err) {
		return nil, fmt.Errorf("path '%v' doesn't exist", config.Path)
	}

	b := strings.Split(config.Path, `/`)
	bin := b[len(b)-1]

	if core.Settings.BuildBinary {
		if err := build.Build(config.Path); err != nil {
			return nil, fmt.Errorf("failed to build sut binary from %s, err %v", config.Path, err)
		}
	}

	var (
		binary = bin
		path   = config.Path
	)

	exec.Run([]string{
		"mkdir", "-p", "/tmp/mtf/cert",
	})

	ports := make(map[docker.ContainerPort]docker.HostPort)
	for _, v := range config.ExposedPorts {
		ports[docker.ContainerPort(v)] = docker.HostPort(v)
	}

	var (
		name     = fmt.Sprintf("sut_mtf-%v", time.Now().Unix())
		image    = "smallinsky/run_sut"
		hostname = "sut_mtf"
	)

	return &docker.ContainerConfig{
		Name:     name,
		Image:    image,
		Hostname: hostname,
		CapAdd:   []string{"NET_RAW", "NET_ADMIN"},
		Env: append([]string{
			fmt.Sprintf("SUT_BINARY_NAME=%s", binary),
		}, config.Env...),
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
	}, nil
}
