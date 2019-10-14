package pubsub

import (
	"time"

	"github.com/smallinsky/mtf/pkg/docker"
)

func NewPubsub(cli *docker.Docker) *Pubsub {
	return &Pubsub{
		cli:   cli,
		ready: make(chan struct{}),
	}
}

type Pubsub struct {
	ready     chan struct{}
	cli       *docker.Docker
	container *docker.ContainerType
}

func (c *Pubsub) Start() error {
	defer close(c.ready)

	var (
		image    = "adilsoncarvalho/gcloud-pubsub-emulator"
		name     = "pubsub_mtf"
		hostname = "pubsub_mtf"
		network  = "mtf_net"
	)

	healtcheck := &docker.HealthCheckConfig{
		Test:     []string{"nc -z localhost:8085"},
		Interval: time.Millisecond * 100,
		Timeout:  time.Second * 3,
	}

	result, err := c.cli.NewContainer(docker.ContainerConfig{
		Image:    image,
		Name:     name,
		Hostname: hostname,
		PortMap: docker.PortMap{
			8085: 8085,
		},
		NetworkName:   network,
		Healtcheck:    healtcheck,
		AttachIfExist: false,
		WaitPolicy:    &docker.WaitForPort{Port: 8085},
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
	return nil
}

func (m *Pubsub) StartPriority() int {
	return 1
}
