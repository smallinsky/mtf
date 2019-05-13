package components

import (
	"fmt"
	"log"
	"net"
	"path/filepath"
	"time"
)

func NewSUT() *SUT {
	return &SUT{}
}

type SUT struct {
	Path  string
	start time.Time
}

func (c *SUT) Start() {
	c.start = time.Now()

	c.Path = "../e2e/service/echo/"

	var err error
	if c.Path, err = filepath.Abs(c.Path); err != nil {
		log.Printf("[ERROR]: Failed to get absolute path for %v path", c.Path)
	}

	var (
		name  = "sut"
		port  = "8001"
		image = "run_sut"
		// TODO Get binary base on the path and repo name or if binary deosn't exist build it.
		// Add ability to run sut from existing image.
		binary = "echo"
		path   = c.Path
	)

	arg := []string{
		"docker", "run", "--rm", "-d",
		fmt.Sprintf("--name=%s_mtf", name),
		fmt.Sprintf("--hostname=%s_mtf", name),
		"--network=mtf_net",
		"-p", fmt.Sprintf("%s:%s", port, port),
		"--cap-add=NET_ADMIN",
		"--cap-add=NET_RAW",
		"-e", fmt.Sprintf("SUT_BINARY_NAME=%v", binary),
		"-e", "ORACLE_ADDR=host.docker.internal:8002",
		"-v", fmt.Sprintf("%s:/component", path),
		image,
	}

	runCmd(arg)
}

func (c *SUT) Ready() {
	waitForPortOpen("localhost", "8001")
	// TODO sync sut start
	time.Sleep(time.Millisecond * 700)
	fmt.Printf("%T start time %v\n", c, time.Now().Sub(c.start))
}

func (c *SUT) Stop() {
	cmd := []string{
		"docker", "kill", fmt.Sprintf("%s_mtf", "sut"),
	}
	runCmd(cmd)
}

func waitForPortOpen(host, port string) {
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

		conn.Close()
		return
	}
}
