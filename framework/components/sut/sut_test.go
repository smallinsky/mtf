package sut

import (
	"testing"
)

func TestSUT(t *testing.T) {
	sut := NewSUT(
		"/Users/Marek/Go/src/github.com/smallinsky/mtf/e2e/service/echo/",
	)
	if err := sut.Start(); err != nil {
		t.Fatalf("Failes to start: %v", err)
	}

	if err := sut.Ready(); err != nil {
		t.Fatalf("ready error: %v", err)
	}

	logs, err := sut.container.Logs()
	if err != nil {
		t.Fatalf("logs error: %v", err)
	}

	t.Logf("Got logs from container: '%s'", logs)

	if err := sut.Stop(); err != nil {
		t.Fatalf("stop error: %v", err)
	}
}
