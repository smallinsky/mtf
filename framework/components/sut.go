package components

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
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
	if _, err := os.Stat(c.Path); os.IsNotExist(err) {
		log.Printf("[ERROR]: Migraitn path: %v doesn't exist\n", c.Path)
		return
	}

	b := strings.Split(c.Path, `/`)
	bin := b[len(b)-1]

	// TODO Add go test flag to rebuild sut binary.
	if false {
		if err := BuildGoBinary(c.Path); err != nil {
			log.Printf("[ERROR]: failed to build sut binary from %s, err %v", c.Path, err)
			return
		}
	}

	var (
		name  = "sut"
		port  = "8001"
		image = "run_sut"
		// TODO Get binary base on the path and repo name or if binary deosn't exist build it.
		// Add ability to run sut from existing image.
		binary = bin
		path   = c.Path
	)

	runCmd([]string{
		"mkdir", "-p", "/tmp/mtf/cert",
	})

	runCmd([]string{
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
		"-v", "/tmp/mtf/cert:/usr/local/share/ca-certificates",
		image,
	})
}

func BuildGoBinary(path string) error {
	var err error
	if path, err = filepath.Abs(path); err != nil {
		return errors.Wrapf(err, "failed to get abs path")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.Wrapf(err, "dir doesn't exist")
	}

	b := strings.Split(path, `/`)
	bin := b[len(b)-1]

	cmd := []string{
		"go", "build", "-o", fmt.Sprintf("%s/%s", path, bin), path,
	}

	if err := runCmd(cmd, WithEnv("GOOS=linux", "GOARCH=amd64")); err != nil {
		return errors.Wrapf(err, "failed to run cmd")
	}
	return nil
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
