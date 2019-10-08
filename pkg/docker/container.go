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
	Name            string
	Image           string
	PortMap         PortMap
	PublishAllPorts bool
	EntryPoint      []string
	Cmd             []string
	Hostname        string
	CapAdd          []string
	Labels          map[string]string
	Env             []string
	Mounts          Mounts

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

// ImagePull fetch image from docker.io registry.
func (c *Client) ImagePull(ctx context.Context, image string) error {
	pull, err := c.cli.ImagePull(ctx, "docker.io/"+image, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull %s image: %v", image, err)
	}
	if _, err := ioutil.ReadAll(pull); err != nil {
		return err
	}
	return nil
}

// ImageExists return if image is already present.
func (c Client) ImageExists(ctx context.Context, image string) bool {
	_, _, err := c.cli.ImageInspectWithRaw(ctx, image)
	return !client.IsErrNotFound(err)
}

// ImageExists return if image is already present.
func (c Client) ContainerExists(ctx context.Context, containerID string) bool {
	result, err := c.cli.ContainerInspect(ctx, containerID)
	fmt.Printf("%s\n", toJson(result.ContainerJSONBase))
	return !client.IsErrContainerNotFound(err)
}

func toJson(i interface{}) string {
	buff, err := json.MarshalIndent(i, "", " ")
	if err != nil {
		panic(err)
	}
	return string(buff)
}

// PullImageIfNotExist fetch image from remote repository if not exits.
func (c *Client) PullImageIfNotExist(ctx context.Context, image string) error {
	if c.ImageExists(ctx, image) {
		return nil
	}
	return c.ImagePull(ctx, image)
}

func (c *Client) RemoveContainer(ctx context.Context, id string) error {
	return c.cli.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{
		Force: true,
	})
}

func (c *Client) NewContainer(config Config) (*Container, error) {
	if err := c.PullImageIfNotExist(context.Background(), config.Image); err != nil {
		return nil, fmt.Errorf("failed to pull image: %v", err)
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
			PortBindings:    config.PortMap.toNatPortMap(),
			Mounts:          config.Mounts.toDockerType(),
			CapAdd:          config.CapAdd,
			AutoRemove:      config.AutoRemove,
			PublishAllPorts: config.PublishAllPorts,
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
