package mysql

import (
	"fmt"
	"log"
	"time"

	"github.com/docker/docker/client"

	"github.com/smallinsky/mtf/framework/components/migrate"
	"github.com/smallinsky/mtf/pkg/docker"
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

	container *docker.Container
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
			fmt.Sprintf("MYSQL_DATABASE=%s", database),
			fmt.Sprintf("MYSQL_ROOT_PASSWORD=%s", password),
		},
		Cmd: []string{
			"--default-authentication-plugin=mysql_native_password",
		},

		Healtcheck: &docker.Healtcheck{
			Test: []string{"CMD", "mysqladmin", "-h", "localhost", "status", fmt.Sprintf("--password=%s", password)},

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
	time.Sleep(time.Millisecond * 200)
	migrate := &migrate.MigrateDB{}
	if err := migrate.Start(); err != nil {
		fmt.Printf("migrate start error: %v", err)
		return err
	}
	fmt.Printf("%T start time %v\n", c, time.Now().Sub(c.start))
	return nil
}
