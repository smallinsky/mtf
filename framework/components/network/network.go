package network

import (
	"github.com/docker/docker/client"
	"github.com/smallinsky/mtf/pkg/docker"
)

type Network struct {
	startC chan struct{}
	name   string

	dockerNet *docker.Netowrk
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

	net, err := docker.NewNetwork(cli, docker.NetworkSettings{
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
