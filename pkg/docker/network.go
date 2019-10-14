package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type NetworkConfig struct {
	Name          string
	Labels        map[string]string
	AttachIfExist bool
}

type Network struct {
	ID string
	NetworkConfig

	cli *client.Client
}

func NewNetwork(client *Docker, config NetworkConfig) (*Network, error) {
	if result, err := client.cli.NetworkInspect(context.Background(), config.Name); err == nil && config.AttachIfExist {
		return &Network{
			ID:            result.ID,
			NetworkConfig: config,
			cli:           client.cli,
		}, nil
	}

	result, err := client.cli.NetworkCreate(context.Background(), config.Name, types.NetworkCreate{
		CheckDuplicate: true,
		Labels:         config.Labels,
	})
	if err != nil {
		return nil, err
	}

	return &Network{
		NetworkConfig: config,
		ID:            result.ID,
		cli:           client.cli,
	}, nil
}

func (n *Network) Close() error {
	return n.cli.NetworkRemove(context.Background(), n.ID)
}
