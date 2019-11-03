package docker

import (
	"context"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Container interface {
	Start() error
	Stop() error
	//Terminate() error
	//	Logs(context.Context) (string, error)
	GetStateV2(context.Context) (*types.ContainerState, error)
}

type WaitPolicy interface {
	WaitForIt(context.Context, *ContainerType) error
	getHealthCheck() *HealthCheckConfig
}

type ContainerType struct {
	ID         string
	WaitPolicy WaitPolicy

	cli    *client.Client
	config ContainerConfig
}

func (c *ContainerType) Name() string {
	return c.config.Name
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

func (c *ContainerType) Start() error {
	if err := c.cli.ContainerStart(context.Background(), c.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	if c.WaitPolicy != nil {
		if err := c.WaitPolicy.WaitForIt(context.Background(), c); err != nil {
			return err
		}
	}
	return nil
}

func (c *ContainerType) GetState() (*types.ContainerState, error) {
	result, err := c.cli.ContainerInspect(context.Background(), c.ID)
	if err != nil {
		return nil, err
	}
	if result.State == nil {
		return nil, fmt.Errorf("state is nil")
	}
	return result.State, nil
}

func (c *ContainerType) WaitForReady() (state *types.ContainerState, err error) {
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

func (c *ContainerType) GetStateV2(ctx context.Context) (*types.ContainerState, error) {
	result, err := c.cli.ContainerInspect(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	if result.State == nil {
		return nil, fmt.Errorf("got nil state")
	}
	return result.State, nil
}

func (c *ContainerType) WaitForStatusHealthly() (state *types.ContainerState, err error) {
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

func (c *ContainerType) Stop() error {
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

func (c *ContainerType) Logs() (string, error) {
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
