package redis

import (
	"fmt"

	"github.com/smallinsky/mtf/pkg/docker"
)

type Redis struct {
	container *docker.Container
	cli       *docker.Client
	cfg       RedisConfig

	ready chan struct{}
}

type RedisConfig struct {
	Password string
	Labels   map[string]string
}

func NewRedis(cli *docker.Client, config RedisConfig) *Redis {
	return &Redis{
		cfg:   config,
		cli:   cli,
		ready: make(chan struct{}),
	}
}

func (c *Redis) Start() error {
	defer close(c.ready)
	var (
		image = "bitnami/redis:4.0"
	)

	result, err := c.cli.NewContainer(docker.Config{
		Name:     "redis_mtf",
		Image:    image,
		Hostname: "redis_mtf",
		Labels: map[string]string{
			"mtf": "mtf",
		},
		PortMap: docker.PortMap{
			6379: 6379,
		},
		NetworkName: "mtf_net",
		Env: []string{
			fmt.Sprintf("REDIS_PASSWORD=%s", c.cfg.Password),
		},
	})
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
	if c.container == nil {
		return fmt.Errorf("container is not running")
	}
	<-c.ready
	return nil
}

func (n *Redis) StartPriority() int {
	return 1
}
