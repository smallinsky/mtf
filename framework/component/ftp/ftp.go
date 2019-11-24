package ftp

import (
	"context"

	"github.com/smallinsky/mtf/pkg/docker"
)

type Component struct {
	config    FTPConfig
	Container docker.Container
}

func New(cli *docker.Docker, config FTPConfig) (*Component, error) {
	containerConf, err := BuildContainerConfig(config)
	if err != nil {
		return nil, err
	}

	container, err := cli.NewContainer(*containerConf)
	if err != nil {
		return nil, err
	}

	return &Component{
		config:    config,
		Container: container,
	}, nil
}

func (c *Component) Start(ctx context.Context) error {
	err := c.Container.Start(ctx)
	return err
}

func (c *Component) Stop(ctx context.Context) error {
	return c.Container.Stop(ctx)
}
