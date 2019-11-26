// +build docker

package docker

import (
	"testing"
)

func TestContainer(t *testing.T) {
	cli, err := New()
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
	t.Logf("Container created [ID:%s]", container.ID)
}
