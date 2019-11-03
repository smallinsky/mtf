package pubsub

import (
	"github.com/smallinsky/mtf/pkg/docker"
)

type Config struct {
	ProjectID          string
	TopicSubscriptions []TopicSubscriptions
}

type TopicSubscriptions struct {
	Topic         string
	Subscriptions []string
}

func BuildContainerConfig() (*docker.ContainerConfig, error) {
	var (
		image   = "smallinsky/pubsub_emulator"
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
