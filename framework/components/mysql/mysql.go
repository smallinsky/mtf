package mysql

import (
	"fmt"
	"time"

	"github.com/docker/docker/client"

	"github.com/smallinsky/mtf/pkg/docker"
)

type MySQLConfig struct {
	Database string
	Password string
	Hostname string
	Network  string
}

func NewMySQL(cli *client.Client, config MySQLConfig) *MySQL {
	return &MySQL{
		cli:    cli,
		config: config,
		ready:  make(chan struct{}),
	}
}

type MySQL struct {
	ready     chan struct{}
	container *docker.Container
	cli       *client.Client

	config MySQLConfig
}

func (c *MySQL) Start() error {
	defer close(c.ready)

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	container, err := docker.NewContainer(cli, docker.Config{
		Name:     "mysql_mtf",
		Image:    "mysql",
		Hostname: "mysql_mtf",
		Labels: map[string]string{
			"mtf": "mtf",
		},
		PortMap: docker.PortMap{
			3306: 3306,
		},
		NetworkName: "mtf_net",
		Env: []string{
			"name=mysql_mtf",
			"hostname=mysql_mtf",
			"network=mtf_net",
			fmt.Sprintf("MYSQL_DATABASE=%s", c.config.Database),
			fmt.Sprintf("MYSQL_ROOT_PASSWORD=%s", c.config.Password),
		},
		Cmd: []string{
			"--default-authentication-plugin=mysql_native_password",
		},

		Healtcheck: &docker.Healtcheck{
			Test: []string{"CMD", "mysqladmin", "-h", "localhost", "status", fmt.Sprintf("--password=%s", c.config.Password)},

			Interval: time.Millisecond * 500,
			Timeout:  time.Second * 40,
		},
	})
	if err != nil {
		return err
	}
	if err := container.Start(); err != nil {
		return err
	}
	c.container = container
	return nil
}

func (c *MySQL) Stop() error {
	return c.container.Stop()
}

func (c *MySQL) Ready() error {
	state, err := c.container.WaitForStatusHealthly()
	if err != nil {
		return err
	}
	if state.Status != "running" {
		return fmt.Errorf("container is in wrong state %v", state.Status)
	}
	<-c.ready
	return nil
}

func (m *MySQL) StartPriority() int {
	return 2
}
