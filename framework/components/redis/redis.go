package redis

import (
	"fmt"

	"github.com/smallinsky/mtf/pkg/docker"
)

type Redis struct {
	container *docker.ContainerType
	cli       *docker.Docker
	cfg       RedisConfig
}

type RedisConfig struct {
	Password string
	Port     string
	Labels   map[string]string
}

func BuildContainerConfig(config RedisConfig) (*docker.ContainerConfig, error) {
	var (
		image    = "bitnami/redis:4.0"
		name     = "redis_mtf"
		hostname = "redis_mtf"
		network  = "mtf_net"
	)

	return &docker.ContainerConfig{
		Name:     name,
		Image:    image,
		Hostname: hostname,
		PortMap: docker.PortMap{
			6379: 6379,
		},
		NetworkName: network,
		Env: []string{
			fmt.Sprintf("REDIS_PASSWORD=%s", config.Password),
		},
	}, nil

}

func NewRedis(cli *docker.Docker, config RedisConfig) *Redis {
	return &Redis{
		cfg: config,
		cli: cli,
	}
}

func (c *Redis) Start() error {
	var (
		image    = "bitnami/redis:4.0"
		name     = "redis_mtf"
		hostname = "redis_mtf"
		network  = "mtf_net"
	)

	dockerConf := docker.ContainerConfig{
		Name:     name,
		Image:    image,
		Hostname: hostname,
		PortMap: docker.PortMap{
			6379: 6379,
		},
		NetworkName: network,
		Env: []string{
			fmt.Sprintf("REDIS_PASSWORD=%s", c.cfg.Password),
		},
	}

	result, err := c.cli.NewContainer(dockerConf)
	if err != nil {
		return err
	}

	c.container = result

	return c.container.Start()
}

func (c *Redis) Stop() error {
	if c.container == nil {
		return fmt.Errorf("container is not running")
	}
	return c.container.Stop()
}

func (c *Redis) Ready() error {
	return nil
}

func (n *Redis) StartPriority() int {
	return 1
}
