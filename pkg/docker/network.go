package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type NetworkConfig struct {
	Name   string
	Labels map[string]string
}

type Network struct {
	ID string
	NetworkConfig

	dockerClient *client.Client
}

func NewNetwork(dockerClient *client.Client, cfg NetworkConfig) (*Network, error) {
	result, err := dockerClient.NetworkCreate(context.Background(), cfg.Name, types.NetworkCreate{
		CheckDuplicate: true,
		Labels:         cfg.Labels,
	})
	if err != nil {
		return nil, err
	}

	net := &Network{
		NetworkConfig: cfg,
		ID:            result.ID,
		dockerClient:  dockerClient,
	}

	return net, nil
}

func (n *Network) Close() error {
	return n.dockerClient.NetworkRemove(context.Background(), n.ID)
}
