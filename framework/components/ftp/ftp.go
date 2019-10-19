package ftp

import (
	"github.com/smallinsky/mtf/pkg/docker"
)

type FTPConfig struct {
	User     string
	Password string
}

func BuildContainerConfig(cfg FTPConfig) (*docker.ContainerConfig, error) {
	var (
		image   = "smallinsky/ftpserver"
		name    = "ftp_mtf"
		network = "mtf_net"
	)

	return &docker.ContainerConfig{
		Image:       image,
		Name:        name,
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
	}, nil
}
