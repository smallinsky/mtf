// +build docker

package sut

import (
	"testing"
	"time"

	"github.com/smallinsky/mtf/pkg/docker"
)

func TestSUT(t *testing.T) {

	start := time.Now()

	cli, err := docker.NewClient()
	if err != nil {
		t.Fatalf("failed to create docker client %v", err)
	}

	t.Logf("Create docker client - %v\n", time.Since(start))
	start = time.Now()
	sut := NewSUT(cli, SutConfig{
		Path: "./test_service/",
	})
	t.Logf("NetSut - %v\n", time.Since(start))
	start = time.Now()
	if err := sut.Start(); err != nil {
		t.Fatalf("Failes to start: %v", err)
	}
	t.Logf("Start() - %v\n", time.Since(start))
	start = time.Now()

	if err := sut.Ready(); err != nil {
		t.Fatalf("ready error: %v", err)
	}

	t.Logf("Ready() - %v\n", time.Since(start))
	start = time.Now()
	if false {
		logs, err := sut.container.Logs()
		if err != nil {
			t.Fatalf("logs error: %v", err)
		}

		if len(logs) != 0 {
			t.Logf("Got logs from container: '%s'", logs)
		}
	}

	if err := sut.Stop(); err != nil {
		t.Fatalf("stop error: %v", err)
	}
	t.Logf("Stop() - %v\n", time.Since(start))
}
