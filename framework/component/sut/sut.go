package sut

import (
	"context"
	"io"

	"github.com/smallinsky/mtf/pkg/docker"
)

type Component struct {
	config    SutConfig
	Container docker.Container
}

func New(cli *docker.Docker, config SutConfig) (*Component, error) {
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
	return c.Container.Start(ctx)
}

func (c *Component) Stop(ctx context.Context) error {
	return c.Container.Stop(ctx)
}

func (c *Component) Logs(ctx context.Context) (io.Reader, error) {
	return c.Container.Logs(ctx)
}

func (c *Component) Name() string {
	return c.Container.Name()
}
