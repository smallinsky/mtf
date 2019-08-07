package components

import (
	"fmt"
	"log"
	"time"
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
}

func (c *Redis) Start() error {
	c.start = time.Now()
	defer close(c.ready)
	if containerIsRunning("redis_mtf") {
		log.Printf("[INFO] Redis component is already running\n")
		return nil
	}

	var (
		name  = "redis"
		port  = "6379"
		image = "bitnami/redis:4.0"
	)

	cmd := []string{
		"docker", "run", "--rm", "-d",
		fmt.Sprintf("--name=%s_mtf", name),
		fmt.Sprintf("--hostname=%s_mtf", name),
		"--network=mtf_net",
		"--env", "REDIS_PASSWORD=test",
		"-p", fmt.Sprintf("%s:%s", port, port),
		image,
	}
	fmt.Println("Run ", join(cmd))
	return runCmd(cmd)
}

func (c *Redis) Stop() error {
	cmd := []string{
		"docker", "kill", fmt.Sprintf("%s_mtf", "redis"),
	}
	return runCmd(cmd)
}

func (c *Redis) Ready() error {
	<-c.ready
	fmt.Printf("%T start time %v\n", c, time.Now().Sub(c.start))
	return nil
}
