package components

import (
	"fmt"
	"time"
)

func NewPubsub() *Pubsub {
	return &Pubsub{
		ready: make(chan struct{}),
	}
}

type Pubsub struct {
	Pass     string
	Port     string
	DB       []string
	Hostname string
	Network  string

	ready chan struct{}
	start time.Time
}

func (c *Pubsub) Start() {
	c.start = time.Now()
	defer close(c.ready)
	if containerIsRunning("redis_mtf") {
		fmt.Printf("pubsub_mtf is already running")
		return
	}
	cmd := `docker run --rm -d --network=mtf_net --name pubsub_mtf --hostname=pubsub_mtf -p 8085:8085 adilsoncarvalho/gcloud-pubsub-emulator`
	run(cmd)
}

func (c *Pubsub) Stop() {
	return
	run("docker kill pubsub_mtf")
}

func (c *Pubsub) Ready() {
	<-c.ready
	waitForPortOpen("localhost", "8001")
	fmt.Printf("%T start time %v\n", c, time.Now().Sub(c.start))
}
