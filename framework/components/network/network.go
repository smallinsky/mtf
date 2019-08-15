package network

import (
	"github.com/docker/docker/client"
	"github.com/smallinsky/mtf/pkg/docker"
)

type Network struct {
	startC chan struct{}
	net    *docker.Network
	cli    *client.Client
	config NetworkConfig
}

type NetworkConfig struct {
	Name string
}

func New(cli *client.Client, config NetworkConfig) *Network {
	return &Network{
		startC: make(chan struct{}),
		cli:    cli,
		config: config,
	}
}

func (n *Network) Start() error {
	net, err := docker.NewNetwork(n.cli, docker.NetworkConfig{
		Name: n.config.Name,
	})
	if err != nil {
		return err
	}
	n.net = net
	close(n.startC)
	return nil
}

func (n *Network) Stop() error {
	return n.net.Close()
}

func (n *Network) Ready() error {
	<-n.startC
	return nil
}

func (n *Network) StartPriority() int {
	return 0
}
