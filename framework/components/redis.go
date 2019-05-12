package components

import (
	"fmt"
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

func (c *Redis) Start() {
	c.start = time.Now()
	defer close(c.ready)
	if containerIsRunning("redis_mtf") {
		fmt.Printf("mysql_mtf is already running")
		return
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

	runCmd(cmd)
}

func (c *Redis) Stop() {
	return
	cmd := []string{
		"docker", "kill", fmt.Sprintf("%s_mtf", "redis"),
	}
	runCmd(cmd)
}

func (c *Redis) Ready() {
	<-c.ready
	fmt.Printf("%T start time %v\n", c, time.Now().Sub(c.start))
}
