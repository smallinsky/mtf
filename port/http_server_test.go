package port

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestHTTPServer(t *testing.T) {

	var p *HTTPPort

	go func() {
		z, err := NewHTTP(WithTLSHost("*.google.com", "*.googleapis.com", "googleapis.com"))
		if err != nil {
			t.Fatalf("failed to create http port %v", err)
		}
		p = z
	}()
	time.Sleep(time.Second * 1)
	fmt.Println("receving printf")
	t.Logf("receiving")
	m, err := p.Receive(context.Background())
	if err != nil {
		t.Fatalf("failed to create http port %v", err)
	}

	t.Logf("%+v", m)
}
