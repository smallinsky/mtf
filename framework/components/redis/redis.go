package redis

import (
	"fmt"
	"log"
	"time"

	"github.com/docker/docker/client"

	"github.com/smallinsky/mtf/pkg/docker"
	"github.com/smallinsky/mtf/pkg/exec"
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

	c *docker.Container
}

func (c *Redis) Start() error {
	var (
		image = "bitnami/redis:4.0"
	)
	c.start = time.Now()
	defer close(c.ready)
	if containerIsRunning("redis_mtf") {
		log.Printf("[INFO] Redis component is already running\n")
		return nil
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	c1, err := docker.NewContainer(cli, docker.Config{
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
	c.c = c1
	if err != nil {
		return err
	}

	return err
}

func (c *Redis) Stop() error {
	cmd := []string{
		"docker", "kill", fmt.Sprintf("%s_mtf", "redis"),
	}
	return exec.Run(cmd)
}

func (c *Redis) Ready() error {
	<-c.ready
	fmt.Printf("%T start time %v\n", c, time.Now().Sub(c.start))
	return nil
}
