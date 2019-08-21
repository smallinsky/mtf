package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// Wrapper around native docker client with containers and netowrks cache.
type Client struct {
	cli        *client.Client
	networks   []types.NetworkResource
	containers []types.Container
}

func NewClient() (*Client, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	networks, err := cli.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		return nil, err
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}

	return &Client{
		cli:        cli,
		networks:   networks,
		containers: containers,
	}, nil
}
