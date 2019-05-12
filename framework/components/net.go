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

func (c *Net) Start() {
	c.start = time.Now()
	defer close(c.ready)
	if networkExists("mtf_net") {
		return
	}
	run("docker network create --driver bridge mtf_net")
}

func (c *Net) Stop() {
	return
	run("docker network rm mtf_net")
}

func (c *Net) Ready() {
	fmt.Printf("%T start time %v\n", c, time.Now().Sub(c.start))
	return
	<-c.ready
}
