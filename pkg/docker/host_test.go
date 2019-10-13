package docker

import (
	"testing"
)

func TestHostAddr(t *testing.T) {
	addr, err := HostIP()
	if err != nil {
		t.Fatalf("HostIP call failed: %v", err)
	}
	t.Log("Got addr: ", addr)
}
