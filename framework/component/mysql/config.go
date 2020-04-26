package mysql

import (
	"bytes"
	"fmt"
	"github.com/smallinsky/mtf/pkg/docker"
)

type MySQLConfig struct {
	//obsolete
	Database  string
	Databases []string
	Password  string
	Hostname  string
	Network   string

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
		EntryPoint: []string{
			`/bin/bash`, `-c`, createDBCommand(config.Databases),
		},
		Cmd: []string{
			"--default-authentication-plugin=mysql_native_password",
		},
		AttachIfExist: config.AttachIfExist,
		WaitPolicy:    &docker.WaitForCommand{Command: cmd},
	}, nil
}

func createDBCommand(databases []string) string {
	var buff bytes.Buffer
	_, _ = fmt.Fprint(&buff, `echo '`)
	for _, db := range databases {
		_, _ = fmt.Fprintf(&buff, `CREATE DATABASE IF NOT EXISTS %s; `, db)
	}
	_, _ = fmt.Fprintf(&buff, `' > /docker-entrypoint-initdb.d/init.sql; `)
	_, _ = fmt.Fprintf(&buff, `/usr/local/bin/docker-entrypoint.sh --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci`)
	return buff.String()
}
