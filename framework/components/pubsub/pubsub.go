package pubsub

import (
	"github.com/smallinsky/mtf/pkg/docker"
)

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
