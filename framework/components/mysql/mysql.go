package mysql

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/docker/docker/client"

	"github.com/smallinsky/mtf/framework/components/migrate"
	"github.com/smallinsky/mtf/pkg/docker"
	"github.com/smallinsky/mtf/pkg/exec"
)

func NewMySQL() *MySQL {
	return &MySQL{
		ready: make(chan struct{}),
	}
}

type MySQL struct {
	Pass     string
	Port     string
	DB       []string
	Hostname string
	Network  string
	ready    chan struct{}
	start    time.Time

	c *docker.Container
}

func (c *MySQL) Start() error {
	c.start = time.Now()
	defer close(c.ready)

	if containerIsRunning("mysql_mtf") {
		log.Printf("[INFO] MySQL component is already running")
		return nil
	}

	var (
		database = "test_db"
		password = "test"
	)

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	c1, err := docker.NewContainer(cli, docker.Config{
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
			fmt.Sprintf("MYSQL_DATABASE=%s", database),
			fmt.Sprintf("MYSQL_ROOT_PASSWORD=%s", password),
		},
		Cmd: []string{
			"--default-authentication-plugin=mysql_native_password",
		},
	})
	c.c = c1

	if err != nil {
		return err
	}
	return nil
}

func (c *MySQL) Stop() error {
	cmd := []string{
		"docker", "kill", fmt.Sprintf("%s_mtf", "mysql"),
	}
	return exec.Run(cmd)
}

func (c *MySQL) Ready() error {
	waitForOpenPort("localhost", "3306")
	<-c.ready
	migrate := &migrate.MigrateDB{}
	migrate.Start()
	fmt.Printf("%T start time %v\n", c, time.Now().Sub(c.start))
	return nil
}

func waitForOpenPort(host, port string) {
	firstRun := true
	for {
		if firstRun {
			firstRun = false
		} else {
			time.Sleep(time.Millisecond * 50)
		}
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), time.Millisecond*500)
		if err != nil {
			continue
		}
		buff := make([]byte, 100)
		if _, err = conn.Read(buff); err != nil {
			conn.Close()
			continue
		}
		conn.Close()
		return
	}
}
