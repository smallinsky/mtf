package docker

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
)

type Container struct {
	ID string

	cli *client.Client
}

type Config struct {
	Name       string
	Image      string
	PortMap    PortMap
	EntryPoint []string
	Cmd        []string
	Hostname   string
	CapAdd     []string
	Labels     map[string]string
	Env        []string
	Mounts     Mounts

	NetworkName string
}

type ContainerPort int
type HostPort int
type PortMap map[ContainerPort]HostPort

type Mount struct {
	Source string
	Target string
}

type Mounts []Mount

func NewContainer(cli *client.Client, config Config) (*Container, error) {
	exposedPorts := make(nat.PortSet)
	for k := range config.PortMap {
		exposedPorts[toNatPort(k)] = struct{}{}
	}

	result, err := cli.ContainerCreate(
		context.Background(),
		&container.Config{
			Hostname:     config.Hostname,
			AttachStdin:  true,
			AttachStdout: true,
			ExposedPorts: exposedPorts,
			Env:          config.Env,
			Image:        config.Image,
			Entrypoint:   config.EntryPoint,
			Labels:       config.Labels,
			Cmd:          config.Cmd,
		},
		&container.HostConfig{
			PortBindings: config.PortMap.toNatPortMap(),
			Mounts:       config.Mounts.toDockerType(),
			CapAdd:       config.CapAdd,
		},
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				config.NetworkName: {
					Aliases: []string{"networkalias"},
				},
			},
		},
		config.Name,
	)
	if err != nil {
		return nil, err
	}
	if err := cli.ContainerStart(context.Background(), result.ID, types.ContainerStartOptions{}); err != nil {
		_ = cli.Close()
		return nil, err
	}
	return &Container{
		ID:  result.ID,
		cli: cli,
	}, nil
}

func (c *Container) Stop() error {
	err := c.cli.ContainerRemove(context.Background(), c.ID, types.ContainerRemoveOptions{
		Force: true,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Container) Logs() (string, error) {
	r, err := c.cli.ContainerLogs(context.Background(), c.ID, types.ContainerLogsOptions{
		ShowStderr: true,
		ShowStdout: true,
	})
	if err != nil {
		return "", err
	}
	defer func() { r.Close() }()

	var buff bytes.Buffer
	if _, err = stdcopy.StdCopy(&buff, &buff, r); err != nil {
		return "", err
	}

	return buff.String(), nil
}

func (m Mounts) toDockerType() []mount.Mount {
	var out []mount.Mount
	for _, v := range m {
		out = append(out, mount.Mount{
			Type:   mount.TypeBind,
			Source: v.Source,
			Target: v.Target,
		})
	}
	return out
}

func (m PortMap) toNatPortMap() nat.PortMap {
	out := make(nat.PortMap)
	for k, v := range m {
		out[toNatPort(k)] = []nat.PortBinding{{HostPort: strconv.Itoa(int(v))}}
	}
	return out
}

func toNatPort(p ContainerPort) nat.Port {
	return nat.Port(fmt.Sprintf("%d/tcp", p))
}
