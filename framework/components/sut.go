package components

import (
	"fmt"
	"log"
	"net"
	"path/filepath"
	"time"
)

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
	cmd := fmt.Sprintf("docker run --rm -d --name=sut_mtf --hostname=sut_mtf --network=mtf_net -p 8001:8001 --cap-add=NET_ADMIN --cap-add=NET_RAW -v %s:/component run_sut", c.Path)
	run(cmd)
}

func (c *SUT) Ready() {
	waitForPortOpen("localhost", "8001")
	// TODO sync sut start
	time.Sleep(time.Millisecond * 700)
	fmt.Printf("%T start time %v\n", c, time.Now().Sub(c.start))
}

func (c *SUT) Stop() {
	run("docker kill sut_mtf")
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
