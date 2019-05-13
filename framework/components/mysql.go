package components

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
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
}

func (c *MySQL) Start() {
	c.start = time.Now()
	defer close(c.ready)

	if containerIsRunning("mysql_mtf") {
		fmt.Printf("mysql_mtf is already running")
		return
	}

	var (
		name     = "mysql"
		port     = "3306"
		database = "test_db"
		password = "test"
		image    = "mysql"
		arg      = "--default-authentication-plugin=mysql_native_password"
	)

	cmd := []string{
		"docker", "run", "--rm", "-d",
		fmt.Sprintf("--name=%s_mtf", name),
		fmt.Sprintf("--hostname=%s_mtf", name),
		"--network=mtf_net",
		"-p", fmt.Sprintf("%s:%s", port, port),
		"-e", fmt.Sprintf("MYSQL_DATABASE=%v", database),
		"-e", fmt.Sprintf("MYSQL_ROOT_PASSWORD=%v", password),
		image, arg,
	}

	runCmd(cmd)
}

func (c *MySQL) Stop() {
	return
	cmd := []string{
		"docker", "kill", fmt.Sprintf("%s_mtf", "mysql"),
	}
	runCmd(cmd)
}

func (c *MySQL) Ready() {
	waitForOpenPort("localhost", "3306")
	<-c.ready
	migrate := &MigrateDB{}
	migrate.Start()
	fmt.Printf("%T start time %v\n", c, time.Now().Sub(c.start))
}

func run(s string) error {
	args := strings.Split(s, " ")
	cmd := exec.Command(args[0], args[1:len(args)]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	err := cmd.Start()
	if err != nil {
		log.Printf("[ERROR] cmd start: %v , \n", err)
		return err
	}
	err = cmd.Wait()
	if err != nil {
		log.Printf("[ERROR] cmd wait: %v , \n", err)
		return err
	}
	return nil
}

func runCmd(arg []string) error {
	cmd := exec.Command(arg[0], arg[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	err := cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
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
