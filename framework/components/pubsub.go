package components

import "fmt"

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
}

func (c *Pubsub) Start() {
	defer close(c.ready)
	if containerIsRunning("redis_mtf") {
		fmt.Printf("pubsub_mtf is already running")
		return
	}
	cmd := `docker run --rm -d --network=mtf_net --name pubsub_mtf --hostname=pubsub_mtf -p 8085:8085 adilsoncarvalho/gcloud-pubsub-emulator`
	run(cmd)
}

func (c *Pubsub) Stop() {
	run("docker kill pubsub_mtf")
}

func (c *Pubsub) Ready() {
	return
	<-c.ready
	waitForPortOpen("localhost", "8001")
}
