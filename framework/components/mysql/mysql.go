package mysql

import (
	"fmt"

	"github.com/smallinsky/mtf/pkg/docker"
)

type MySQLConfig struct {
	Database string
	Password string
	Hostname string
	Network  string

	Labels        map[string]string
	AttachIfExist bool
}

func NewMySQL(cli *docker.Docker, config MySQLConfig) *MySQL {
	return &MySQL{
		cli:    cli,
		config: config,
	}
}

type MySQL struct {
	container *docker.ContainerType
	cli       *docker.Docker

	config MySQLConfig
}

func (c *MySQL) Start() error {
	var (
		image    = "library/mysql"
		name     = "mysql_mtf"
		hostname = "mysql_mtf"
		network  = "mtf_net"
	)

	cmd := fmt.Sprintf("mysqladmin -h localhost status --password=%s", c.config.Password)

	result, err := c.cli.NewContainer(docker.ContainerConfig{
		Image:       image,
		Name:        name,
		Hostname:    hostname,
		NetworkName: network,
		PortMap: docker.PortMap{
			3306: 3306,
		},
		Env: []string{
			fmt.Sprintf("MYSQL_DATABASE=%s", c.config.Database),
			fmt.Sprintf("MYSQL_ROOT_PASSWORD=%s", c.config.Password),
		},
		Cmd: []string{
			"--default-authentication-plugin=mysql_native_password",
		},
		AttachIfExist: c.config.AttachIfExist,
		WaitPolicy:    &docker.WaitForCommand{Command: cmd},
	})
	if err != nil {
		return err
	}
	c.container = result

	return c.container.Start()
}

func (c *MySQL) Stop() error {
	return c.container.Stop()
}

func (m *MySQL) StartPriority() int {
	return 1
}
