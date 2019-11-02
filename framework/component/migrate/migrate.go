package migrate

import (
	"github.com/smallinsky/mtf/pkg/docker"
)

type Component struct {
	config    MigrateConfig
	Container docker.Container
}

func New(cli docker.Docker, config MigrateConfig) (*Component, error) {
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

func (c *Component) Start() error {
	return c.Container.Start()
}

func (c *Component) Stop() error {
	return c.Container.Stop()
}
