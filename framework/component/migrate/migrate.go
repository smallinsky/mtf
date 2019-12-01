package migrate

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/smallinsky/mtf/pkg/docker"
)

const (
	fetchStatusDelayStep = time.Millisecond * 200
)

type Component struct {
	config    MigrateConfig
	Container docker.Container
}

func New(cli *docker.Docker, config MigrateConfig) (*Component, error) {
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
	if err := c.Container.Start(ctx); err != nil {
		return err
	}

	state, err := c.Container.GetState(ctx)
	if err != nil {
		return err
	}

	for {
		if !state.Running {
			break
		}
		time.Sleep(fetchStatusDelayStep)
		state, err = c.Container.GetState(ctx)
		if err != nil {
			return err
		}
	}

	if state.ExitCode != 0 {
		return c.handleExecutionError(ctx)
	}

	return nil
}

func (c *Component) Stop(ctx context.Context) error {
	return c.Container.Stop(ctx)
}

func (c *Component) handleExecutionError(ctx context.Context) error {
	r, err := c.Container.Logs(ctx)
	if err != nil {
		return err
	}
	buff, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return fmt.Errorf("migarton command filed:\n%s", string(buff))
}
