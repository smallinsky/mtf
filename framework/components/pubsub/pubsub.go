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

func BuildContainerConfig() (*docker.ContainerConfig, error) {
	var (
		image   = "adilsoncarvalho/gcloud-pubsub-emulator"
		name    = "pubsub_mtf"
		network = "mtf_net"
	)

	return &docker.ContainerConfig{
		Image: image,
		Name:  name,
		PortMap: docker.PortMap{
			8085: 8085,
		},
		NetworkName:   network,
		AttachIfExist: false,
		WaitPolicy:    &docker.WaitForPort{Port: 8085},
	}, nil
}

func (m *Pubsub) StartPriority() int {
	return 1
}
