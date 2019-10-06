// +build docker

package docker

import (
	"testing"
)

func TestBuildImage(t *testing.T) {
	cli, err := NewClient()
	if err != nil {
		t.Fatalf("failed to get cli client: %v", err)
	}
	err = cli.BuildImage("../../docker/run_sut", "sut_test_build")

	if err != nil {
		t.Fatalf("failed to build image: %v", err)
	}

}
