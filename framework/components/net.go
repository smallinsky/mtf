package components

import (
	"fmt"
	"time"

	"github.com/smallinsky/mtf/pkg/exec"
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

	return exec.Run(cmd)
}

func (c *Net) Stop() error {
	cmd := []string{
		"docker", "network", "rm", "mtf_net",
	}
	return exec.Run(cmd)
}

func (c *Net) Ready() error {
	fmt.Printf("%T start time %v\n", c, time.Now().Sub(c.start))
	return nil
	<-c.ready
	return nil
}
