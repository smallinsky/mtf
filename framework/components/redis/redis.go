package redis

import (
	"github.com/docker/docker/client"

	"github.com/smallinsky/mtf/pkg/docker"
)

func NewRedis() *Redis {
	return &Redis{
		ready: make(chan struct{}),
	}
}

type Redis struct {
	Pass     string
	Port     string
	DB       []string
	Hostname string
	Network  string
	ready    chan struct{}

	contianer *docker.Container
}

func (c *Redis) Start() error {
	var (
		image = "bitnami/redis:4.0"
	)
	defer close(c.ready)

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	container, err := docker.NewContainer(cli, docker.Config{
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
			"REDIS_PASSWORD=test",
		},
	})
	if err != nil {
		return err
	}
	if err := container.Start(); err != nil {
		return err
	}
	c.contianer = container

	return err
}

func (c *Redis) Stop() error {
	return c.contianer.Stop()
}

func (c *Redis) Ready() error {
	<-c.ready
	return nil
}

func (n *Redis) StartPriority() int {
	return 1
}
