// +build docker

package docker

import (
	"context"
	"testing"
)

func TestContainer(t *testing.T) {
	cli, err := NewClient()
	if err != nil {
		t.Fatalf("failed to create docker client: %v", err)
	}

	container, err := cli.NewContainer(Config{
		Name:  "nginx",
		Image: "nginx",
	})
	if err != nil {
		t.Fatalf("fialed to create container: %v", err)
	}

	container = container
}

func TestFoo(t *testing.T) {
	cli, err := NewClient()
	if err != nil {
		t.Fatalf("failed to create docker client: %v", err)
	}
	_ = cli.ContainerExists(context.Background(), "mysql_mtf")
	_ = cli.ContainerExists(context.Background(), "keen_wescoff")
}
