package docker

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func New() (*Docker, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return &Docker{
		cli: cli,
	}, nil
}

type Docker struct {
	cli *client.Client
}

// ImagePull fetch image from docker.io registry.
func (c *Docker) ImagePull(ctx context.Context, image string) error {
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
func (c Docker) ImageExists(ctx context.Context, image string) bool {
	_, _, err := c.cli.ImageInspectWithRaw(ctx, image)
	return !client.IsErrNotFound(err)
}

// ImageExists return if image is already present.
func (c Docker) ContainerExists(ctx context.Context, containerID string) bool {
	_, err := c.cli.ContainerInspect(ctx, containerID)
	return !client.IsErrContainerNotFound(err)
}

// PullImageIfNotExist fetch image from remote repository if not exits.
func (c *Docker) PullImageIfNotExist(ctx context.Context, image string) error {
	if c.ImageExists(ctx, image) {
		return nil
	}
	return c.ImagePull(ctx, image)
}

func (c *Docker) RemoveContainer(ctx context.Context, id string) error {
	return c.cli.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{
		Force: true,
	})
}

type ContainerConfig struct {
	Image           string
	Env             []string
	Name            string
	PortMap         PortMap
	PublishAllPorts bool
	EntryPoint      []string
	Cmd             []string
	Hostname        string
	CapAdd          []string
	Labels          map[string]string
	Mounts          Mounts
	NetworkName     string
	Healtcheck      *HealthCheckConfig
	AttachIfExist   bool
	AutoRemove      bool
	Privileged      bool
	WaitPolicy      WaitPolicy
}

type HealthCheckConfig struct {
	Test     []string
	Interval time.Duration
	Timeout  time.Duration
	Retries  int
}

func (c *Docker) NewContainer(config ContainerConfig) (*ContainerType, error) {
	if err := c.PullImageIfNotExist(context.Background(), config.Image); err != nil {
		return nil, fmt.Errorf("failed to pull image: %v", err)
	}

	res, err := c.cli.ContainerInspect(context.Background(), config.Name)
	if err == nil {
		if res.State.Running && config.AttachIfExist {
			return &ContainerType{
				ID:     res.ID,
				cli:    c.cli,
				config: config,
			}, nil
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

	if config.WaitPolicy != nil {
		config.Healtcheck = config.WaitPolicy.getHealthCheck()
		hc = &container.HealthConfig{
			Test:     config.Healtcheck.Test,
			Interval: config.Healtcheck.Interval,
			Timeout:  config.Healtcheck.Timeout,
			Retries:  config.Healtcheck.Retries,
		}
	}

	config.Env = append(config.Env, "DOCKER_HOST_ADDR="+hostAddr)

	if config.Hostname == "" {
		config.Hostname = config.Name
	}

	createConf := &container.Config{
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
	}

	if config.Privileged {
		config.CapAdd = append(config.CapAdd, []string{"NET_RAW", "NET_ADMIN"}...)
	}

	hostConf := &container.HostConfig{
		PortBindings:    config.PortMap.toNatPortMap(),
		Mounts:          config.Mounts.toDockerType(),
		CapAdd:          config.CapAdd,
		AutoRemove:      config.AutoRemove,
		PublishAllPorts: config.PublishAllPorts,
	}

	netConf := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			config.NetworkName: {
				Aliases: []string{"networkalias"},
			},
		},
	}

	result, err := c.cli.ContainerCreate(context.Background(), createConf, hostConf, netConf, config.Name)
	if err != nil {
		return nil, err
	}

	return &ContainerType{
		ID:         result.ID,
		cli:        c.cli,
		config:     config,
		WaitPolicy: config.WaitPolicy,
	}, nil
}

func (c *Docker) CreateNetwork(name string) (*Network, error) {
	result, err := c.cli.NetworkInspect(context.Background(), name)
	if err == nil {
		return &Network{
			ID:  result.ID,
			cli: c.cli,
		}, nil
	} else if !client.IsErrNotFound(err) {
		return nil, err
	}

	net, err := c.cli.NetworkCreate(context.Background(), name, types.NetworkCreate{
		CheckDuplicate: true,
	})
	if err != nil {
		return nil, err
	}
	return &Network{
		ID:  net.ID,
		cli: c.cli,
	}, nil
}

func (n *Network) Remove() error {
	//	return n.cli.NetworkRemove(context.Background(), n.ID)
	return nil
}
