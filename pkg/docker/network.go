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

func NewNetwork(client *Client, config NetworkConfig) (*Network, error) {
	networks, err := client.cli.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		return nil, err
	}
	for _, v := range networks {
		if !config.AttachIfExist {
			break
		}
		if v.Name != config.Name {
			continue
		}

		return &Network{
			NetworkConfig: config,
			ID:            v.ID,
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

	net := &Network{
		NetworkConfig: config,
		ID:            result.ID,
		cli:           client.cli,
	}

	return net, nil
}

func (n *Network) Close() error {
	return n.cli.NetworkRemove(context.Background(), n.ID)
}
