package pubsub

import (
	"github.com/smallinsky/mtf/pkg/docker"
)

func NewPubsub(cli *docker.Docker) *Pubsub {
	return &Pubsub{
		cli: cli,
	}
}

type Pubsub struct {
	cli       *docker.Docker
	container *docker.ContainerType
}

func (c *Pubsub) Start() error {

	var (
		image    = "adilsoncarvalho/gcloud-pubsub-emulator"
		name     = "pubsub_mtf"
		hostname = "pubsub_mtf"
		network  = "mtf_net"
	)

	result, err := c.cli.NewContainer(docker.ContainerConfig{
		Image:    image,
		Name:     name,
		Hostname: hostname,
		PortMap: docker.PortMap{
			8085: 8085,
		},
		NetworkName:   network,
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

func (m *Pubsub) StartPriority() int {
	return 1
}
