package port

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/smallinsky/mtf/pkg/cert"
)

func TestHTTPServer(t *testing.T) {
	if _, err := cert.GenCert(nil); err != nil {
		t.Fatalf("failed to generate certs: %v", err)
	}
	startHTTP()
	port := ht.httpPort

	sync := make(chan struct{})
	go func() {
		close(sync)
		m, err := port.receive()
		if err != nil {
			t.Fatalf("failed to create http port %v", err)
		}

		if want, got := http.MethodGet, m.Method; want != got {
			t.Fatalf("Method want: %v, got %v", want, got)
		}
		if want, got := "/testpath", m.URL; want != got {
			t.Fatalf("URL want: %v, got %v", want, got)
		}
		resp := &HTTPResponse{
			Status: http.StatusAccepted,
		}
		if err := port.send(resp); err != nil {
			t.Fatalf("http port: failed to send response %v", err)
		}
	}()
	<-sync

	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/testpath", nil)
	if err != nil {
		t.Fatalf("failed to crate http request: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		t.Fatalf("http call failed: %v", err)
	}

	if want, got := http.StatusAccepted, resp.StatusCode; want != got {
		t.Fatalf("StatusCode want: %v, got %v", want, got)
	}

}
