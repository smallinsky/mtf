package ftp

import (
	"github.com/smallinsky/mtf/pkg/docker"
)

func NewFTP(cli *docker.Docker) *FTP {
	return &FTP{
		cli: cli,
	}
}

type FTPConfig struct {
}

type FTP struct {
	cli       *docker.Docker
	container *docker.ContainerType
}

func (c *FTP) Start() error {
	var (
		image    = "smallinsky/ftpserver"
		name     = "ftp_mtf"
		hostname = "ftp_mtf"
		network  = "mtf_net"
	)

	result, err := c.cli.NewContainer(docker.ContainerConfig{
		Image:       image,
		Name:        name,
		Hostname:    hostname,
		NetworkName: network,
		Env: []string{
			"FTP_USER=test",
			"FTP_PASS=test",
		},
		PortMap: docker.PortMap{
			20:    20,
			21:    21,
			21100: 21100,
			21101: 21101,
			21102: 21102,
			21103: 21103,
			21104: 21104,
			21105: 21105,
			21106: 21106,
			21107: 21107,
			21108: 21108,
			21109: 21109,
			21110: 21110,
		},
		AttachIfExist: false,
		WaitPolicy:    &docker.WaitForPort{Port: 21},
	})
	if err != nil {
		return err
	}

	c.container = result

	return c.container.Start()

}

func (c *FTP) Stop() error {
	return c.container.Stop()
}

func (m *FTP) StartPriority() int {
	return 1
}
