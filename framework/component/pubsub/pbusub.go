package pubsub

import (
	"github.com/smallinsky/mtf/pkg/docker"
)

type Component struct {
	Config    Config
	Container docker.Container
}

func New(cli docker.Docker, config Config) (*Component, error) {
	containerConf, err := BuildContainerConfig()
	if err != nil {
		return nil, err
	}

	container, err := cli.NewContainer(*containerConf)
	if err != nil {
		return nil, err
	}

	return &Component{
		Config:    config,
		Container: container,
	}, nil
}

func (c *Component) Start() error {
	return c.Container.Start()
}

func (c *Component) Stop() error {
	return c.Container.Stop()
}
