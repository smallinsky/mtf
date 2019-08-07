package components

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
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

func (c *MySQL) Start() error {
	c.start = time.Now()
	defer close(c.ready)

	if containerIsRunning("mysql_mtf") {
		log.Printf("[INFO] MySQL component is already running")
		return nil
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

	fmt.Println("Run ", join(cmd))
	return runCmd(cmd)
}

func (c *MySQL) Stop() error {
	cmd := []string{
		"docker", "kill", fmt.Sprintf("%s_mtf", "mysql"),
	}
	return runCmd(cmd)
}

func (c *MySQL) Ready() error {
	waitForOpenPort("localhost", "3306")
	<-c.ready
	migrate := &MigrateDB{}
	migrate.Start()
	fmt.Printf("%T start time %v\n", c, time.Now().Sub(c.start))
	return nil
}

type option func(*exec.Cmd)

func WithEnv(env ...string) option {
	return func(cmd *exec.Cmd) {
		cmd.Env = append(cmd.Env, env...)
	}
}

func runCmd(arg []string, opts ...option) error {
	cmd := exec.Command(arg[0], arg[1:]...)
	cmd.Env = os.Environ()
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stdout
	for _, opt := range opts {
		opt(cmd)
	}

	if buff, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to run: '%v' cmd\nerror: %v\noutput: %v\n", arg, err, string(buff))
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
