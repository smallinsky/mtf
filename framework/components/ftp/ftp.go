package ftp

import (
	"fmt"
	"time"

	"github.com/smallinsky/mtf/pkg/docker"
)

func NewFTP(cli *docker.Client) *FTP {
	return &FTP{
		cli:   cli,
		ready: make(chan struct{}),
	}
}

type FTPConfig struct {
}

type FTP struct {
	ready     chan struct{}
	cli       *docker.Client
	container *docker.Container
}

func (c *FTP) Start() error {
	defer close(c.ready)

	var (
		image = "smallinsky/ftpserver"
		//image = "ftpwithwatcher"
	)

	result, err := c.cli.NewContainer(docker.Config{
		Name:     "ftp_mtf",
		Image:    image,
		Hostname: "ftp_mtf",
		Labels: map[string]string{
			"mtf": "mtf",
		},
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
		NetworkName: "mtf_net",
		Healtcheck: &docker.Healtcheck{
			Test:     []string{"nc -z localhost:21"},
			Interval: time.Millisecond * 100,
			Timeout:  time.Second * 3,
		},
		AttachIfExist: false,
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

func (c *FTP) Ready() error {
	state, err := c.container.GetState()
	if err != nil {
		return err
	}

	if state.Status != "running" {
		return fmt.Errorf("container is in wrong state %v", state.Status)
	}
	return nil
}

func (m *FTP) StartPriority() int {
	return 1
}
