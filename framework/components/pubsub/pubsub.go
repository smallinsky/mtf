package pubsub

import (
	"fmt"
	"time"

	"github.com/smallinsky/mtf/pkg/docker"
)

func NewPubsub(cli *docker.Client) *Pubsub {
	return &Pubsub{
		cli:   cli,
		ready: make(chan struct{}),
	}
}

type Pubsub struct {
	ready     chan struct{}
	cli       *docker.Client
	container *docker.Container
}

func (c *Pubsub) Start() error {
	defer close(c.ready)

	var (
		image = "adilsoncarvalho/gcloud-pubsub-emulator"
	)

	result, err := c.cli.NewContainer(docker.Config{
		Name:     "pubsub_mtf",
		Image:    image,
		Hostname: "pubsub_mtf",
		Labels: map[string]string{
			"mtf": "mtf",
		},
		PortMap: docker.PortMap{
			8085: 8085,
		},
		NetworkName: "mtf_net",
		Healtcheck: &docker.Healtcheck{
			Test:     []string{"nc -z localhost:8085"},
			Interval: time.Millisecond * 100,
			Timeout:  time.Second * 3,
		},
		AttachIfExist: false,
	})
	if err != nil {
		return err
	}

	c.container = result

	return c.container.Start()

}

func (c *Pubsub) Stop() error {
	return c.container.Stop()
}

func (c *Pubsub) Ready() error {
	state, err := c.container.GetState()
	if err != nil {
		return err
	}

	if state.Status != "running" {
		return fmt.Errorf("container is in wrong state %v", state.Status)
	}
	return nil
}

func (m *Pubsub) StartPriority() int {
	return 1
}
