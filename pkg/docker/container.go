package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Container struct {
	ID string

	cli    *client.Client
	config Config
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

	Healtcheck *Healtcheck

	AttachIfExist bool
	AutoRemove    bool
}

type Healtcheck struct {
	Test     []string
	Interval time.Duration
	Timeout  time.Duration
	Retries  int
}

type State struct {
	ExitCode int
	Status   string
}

type ContainerPort int
type HostPort int
type PortMap map[ContainerPort]HostPort

type Mount struct {
	Source string
	Target string
}

type Mounts []Mount

type Options func(*options)

type options struct {
	ContainerID string
}

func (c *Client) NewContainer(config Config, opts ...Options) (*Container, error) {
	var options options
	for _, opt := range opts {
		opt(&options)
	}

	for _, v := range c.containers {
		if v.Names[0] != "/"+config.Name {
			continue
		}
		container := &Container{
			ID:     v.ID,
			cli:    c.cli,
			config: config,
		}

		// container already exists, check if it is healty and can be reused.
		if v.State == "running" && config.AttachIfExist {
			return container, nil
		}
		err := c.cli.ContainerRemove(context.Background(), config.Name, types.ContainerRemoveOptions{
			Force: true,
		})
		if err != nil {
			return nil, err
		}
	}

	exposedPorts := make(nat.PortSet)
	for k := range config.PortMap {
		exposedPorts[toNatPort(k)] = struct{}{}
	}

	var hc *container.HealthConfig
	if config.Healtcheck != nil {
		hc = &container.HealthConfig{
			Test:     config.Healtcheck.Test,
			Interval: config.Healtcheck.Interval,
			Timeout:  config.Healtcheck.Timeout,
			Retries:  config.Healtcheck.Retries,
		}
	}

	result, err := c.cli.ContainerCreate(
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
			Healthcheck:  hc,
		},
		&container.HostConfig{
			PortBindings: config.PortMap.toNatPortMap(),
			Mounts:       config.Mounts.toDockerType(),
			CapAdd:       config.CapAdd,
			AutoRemove:   config.AutoRemove,
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

	return &Container{
		ID:     result.ID,
		cli:    c.cli,
		config: config,
	}, nil
}

func (c *Container) Start() error {
	if err := c.cli.ContainerStart(context.Background(), c.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}
	return nil
}

func (c *Container) GetState() (*types.ContainerState, error) {
	result, err := c.cli.ContainerInspect(context.Background(), c.ID)
	if err != nil {
		return nil, err
	}
	if result.State == nil {
		return nil, fmt.Errorf("state is nil")
	}
	return result.State, nil
}

func (c *Container) WaitForReady() (state *types.ContainerState, err error) {
	if c.config.Healtcheck == nil {
		return nil, fmt.Errorf("heltcheck was not set")
	}
	for {
		state, err = c.GetState()
		if err != nil {
			return nil, err
		}
		if state.Health == nil {
			return nil, fmt.Errorf("failed to get health status")
		}

		if state.Health.Status == types.Starting {
			time.Sleep(time.Millisecond * 100)
			continue
		}
		break
	}
	return state, nil
}

func (c *Container) WaitForStatusHealthly() (state *types.ContainerState, err error) {
	if c.config.Healtcheck == nil {
		return nil, fmt.Errorf("heltcheck was not set")
	}
	for {
		state, err = c.GetState()
		if err != nil {
			return nil, err
		}
		if state.Health == nil {
			return nil, fmt.Errorf("failed to get health status")
		}

		if state.Health.Status != types.Healthy {
			continue
		}
		break
	}
	return state, nil
}

func dump(i interface{}) string {
	buff, _ := json.MarshalIndent(i, "", " ")
	return string(buff)
}

func (c *Container) Stop() error {
	if c == nil {
		return fmt.Errorf("container is nil")
	}
	if c.config.AttachIfExist {
		return nil
	}

	err := c.cli.ContainerRemove(context.Background(), c.ID, types.ContainerRemoveOptions{
		Force: true,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Container) Logs() (string, error) {
	if c.cli == nil {
		return "", fmt.Errorf("cli is nil")
	}
	r, err := c.cli.ContainerLogs(context.Background(), c.ID, types.ContainerLogsOptions{
		ShowStderr: true,
		ShowStdout: true,
	})
	if err != nil {
		return "", err
	}
	defer func() { r.Close() }()

	buff, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}

	return string(buff), nil
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
