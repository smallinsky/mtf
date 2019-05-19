package components

import (
	"fmt"
	"time"
)

func NewNet() *Net {
	return &Net{
		ready: make(chan struct{}),
	}
}

type Net struct {
	ready chan struct{}
	start time.Time
}

func (c *Net) Start() error {
	c.start = time.Now()
	defer close(c.ready)
	if networkExists("mtf_net") {
		return nil
	}

	var (
		name = "mtf_net"
	)

	cmd := []string{
		"docker", "network", "create",
		"--driver", "bridge", name,
	}

	return runCmd(cmd)
}

func (c *Net) Stop() error {
	return nil
	cmd := []string{
		"docker", "network", "rm", "mtf_net",
	}
	return runCmd(cmd)
}

func (c *Net) Ready() error {
	fmt.Printf("%T start time %v\n", c, time.Now().Sub(c.start))
	return nil
	<-c.ready
	return nil
}
