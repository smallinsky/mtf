package redis

import (
	"fmt"

	"github.com/smallinsky/mtf/pkg/docker"
)

type RedisConfig struct {
	Password string
	Port     string
}

func BuildContainerConfig(config RedisConfig) (*docker.ContainerConfig, error) {
	var (
		image   = "bitnami/redis:4.0"
		name    = "redis_mtf"
		network = "mtf_net"
	)

	return &docker.ContainerConfig{
		Name:  name,
		Image: image,
		PortMap: docker.PortMap{
			6379: 6379,
		},
		NetworkName: network,
		Env: []string{
			fmt.Sprintf("REDIS_PASSWORD=%s", config.Password),
		},
	}, nil
}
