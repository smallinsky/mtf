package network

import (
	"github.com/smallinsky/mtf/pkg/docker"
)

type Network struct {
	startC chan struct{}
	net    *docker.Network
	cli    *docker.Client
	config NetworkConfig
}

type NetworkConfig struct {
	Name          string
	Labels        map[string]string
	AttachIfExist bool
}

func New(cli *docker.Client, config NetworkConfig) *Network {
	return &Network{
		startC: make(chan struct{}),
		cli:    cli,
		config: config,
	}
}

func (n *Network) Start() error {
	net, err := docker.NewNetwork(n.cli, docker.NetworkConfig{
		Name:          n.config.Name,
		AttachIfExist: n.config.AttachIfExist,
	})
	if err != nil {
		return err
	}
	n.net = net
	close(n.startC)
	return nil
}

func (n *Network) Stop() error {
	if n.config.AttachIfExist {
		return nil
	}
	return n.net.Close()
}

func (n *Network) Ready() error {
	<-n.startC
	return nil
}

func (n *Network) StartPriority() int {
	return 0
}
