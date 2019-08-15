package network

import (
	"github.com/docker/docker/client"
	"github.com/smallinsky/mtf/pkg/docker"
)

type Network struct {
	startC chan struct{}
	name   string

	dockerNet *docker.Network
}

func New() *Network {
	return &Network{
		name:   "mtf_net",
		startC: make(chan struct{}),
	}
}

func (n *Network) Start() error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	net, err := docker.NewNetwork(cli, docker.NetworkConfig{
		Name: n.name,
	})
	if err != nil {
		return err
	}
	n.dockerNet = net
	close(n.startC)
	return nil
}

func (n *Network) Stop() error {
	return n.dockerNet.Close()
}

func (n *Network) Ready() error {
	<-n.startC
	return nil
}

func (n *Network) StartPriority() int {
	return 0
}
