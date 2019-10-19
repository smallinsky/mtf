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

	AttachIfExist bool
}

func BuildContainerConfig(config MySQLConfig) (*docker.ContainerConfig, error) {
	var (
		image   = "library/mysql"
		name    = "mysql_mtf"
		network = "mtf_net"
	)

	cmd := fmt.Sprintf("mysqladmin -h localhost status --password=%s", config.Password)

	return &docker.ContainerConfig{
		Image:       image,
		Name:        name,
		NetworkName: network,
		PortMap: docker.PortMap{
			3306: 3306,
		},
		Env: []string{
			fmt.Sprintf("MYSQL_DATABASE=%s", config.Database),
			fmt.Sprintf("MYSQL_ROOT_PASSWORD=%s", config.Password),
		},
		Cmd: []string{
			"--default-authentication-plugin=mysql_native_password",
		},
		AttachIfExist: config.AttachIfExist,
		WaitPolicy:    &docker.WaitForCommand{Command: cmd},
	}, nil
}
