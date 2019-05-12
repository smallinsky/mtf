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
	if containerIsRunning("pubsub_mtf") {
		fmt.Printf("pubsub_mtf is already running")
		return
	}

	var (
		name  = "pubsub"
		port  = "8085"
		image = "adilsoncarvalho/gcloud-pubsub-emulator"
	)

	cmd := []string{
		"docker", "run", "--rm", "-d",
		fmt.Sprintf("--name=%s_mtf", name),
		fmt.Sprintf("--hostname=%s_mtf", name),
		"--network=mtf_net",
		"-p", fmt.Sprintf("%s:%s", port, port),
		image,
	}

	runCmd(cmd)
}

func (c *Pubsub) Stop() {
	return
	cmd := []string{
		"docker", "kill", fmt.Sprintf("%s_mtf", "pubsub"),
	}
	runCmd(cmd)
}

func (c *Pubsub) Ready() {
	<-c.ready
	waitForPortOpen("localhost", "8001")
	fmt.Printf("%T start time %v\n", c, time.Now().Sub(c.start))
}
