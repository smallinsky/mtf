package redis

import (
	"fmt"
	"time"

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
	start    time.Time

	contianer *docker.Container
}

func (c *Redis) Start() error {
	var (
		image = "bitnami/redis:4.0"
	)
	c.start = time.Now()
	defer close(c.ready)
	//if containerIsRunning("redis_mtf") {
	//	log.Printf("[INFO] Redis component is already running\n")
	//	return nil
	//}

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
	fmt.Printf("%T start time %v\n", c, time.Now().Sub(c.start))
	return nil
}
