package port

import (
	"log"
	"net/http"
	"runtime"
	"sync"

	"github.com/gorilla/mux"
	"github.com/smallinsky/mtf/pkg/cert"
)

var (
	ht   *httpserver
	once sync.Once
)

type httpserver struct {
	router *mux.Router
	wg     sync.WaitGroup

	httpPort *HTTPPort
	gcs      *GCStorage
}

func startHTTP() {
	return
	once.Do(func() {
		ht = &httpserver{
			router:   mux.NewRouter(),
			httpPort: newHTTPPort(),
			gcs:      NewGCStoragePort(),
		}
		ht.httpPort.Register(ht.router)
		if ht.gcs != nil {
			ht.gcs.registerRouter(ht.router)
		}
		if err := ht.run(); err != nil {
			panic(err)

		}
	})
}

func (ht *httpserver) run() error {
	if err := ht.servHTTP(); err != nil {
		return err
	}
	if err := ht.serveHTTPS(); err != nil {
		return err
	}
	ht.wg.Wait()
	return nil
}

func (ht *httpserver) serveHTTPS() error {
	ht.wg.Add(1)
	go func() {
		ht.wg.Done()
		if err := http.ListenAndServeTLS(":8443", cert.ServerCertFile, cert.ServerKeyFile, ht.router); err != nil {
			log.Fatalf("faield to start tls server: %v", err)
		}
	}()
	runtime.Gosched()
	return nil
}

func (ht *httpserver) servHTTP() error {
	ht.wg.Add(1)
	go func() {
		ht.wg.Done()
		if err := http.ListenAndServe(":8080", ht.router); err != nil {
			log.Fatalf("faield to start tls server: %v", err)
		}
	}()
	runtime.Gosched()
	return nil
}
