package port

import (
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"time"
)

//TODO Add https support
func NewHTTP() HTTPPort {
	p := HTTPPort{
		req:   make(chan HttpRequest),
		resp:  make(chan HttpResponse),
		sync:  make(chan struct{}),
		syncR: make(chan struct{}),
	}
	p.serve()
	return p
}

type HttpRequest struct {
	Body   string
	URL    string
	Method string
}

func Match(l *HttpRequest, r *http.Request) {
	//TODO add patter matching
	if l.URL != "*" && l.URL != r.URL.String() {
		log.Fatalf("URL matcher failed, expected: %s, got: %s\n", l.URL, r.URL.String())
	}

	if l.Method != "*" && l.Method != r.Method {
		log.Fatalf("Method matcher failed, expected: %s, got: %s\n", l.Method, r.Method)
	}
}

type HttpResponse struct {
	Body string
}

type HTTPPort struct {
	req  chan HttpRequest
	resp chan HttpResponse
	sync chan struct{}

	rcvSync map[HttpRequest]chan struct{}
	syncR   chan struct{}

	svr *httptest.Server
}

func (m *HTTPPort) serve() {
	var err error

	m.svr = httptest.NewUnstartedServer(http.HandlerFunc(m.Handle))
	m.svr.Listener, err = net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Filed to start net listener: %v", err)
	}
	m.svr.Start()
}

func (m *HTTPPort) Stop() {
	m.svr.Close()
}

func (m *HTTPPort) Handle(w http.ResponseWriter, r *http.Request) {
	log.Printf("Got http request URL: %s \n", r.URL.String())
	mr := <-m.req
	mr = mr
	Match(&mr, r)
	msgS := <-m.resp
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(msgS.Body))
	m.sync <- struct{}{}
}

func (m *HTTPPort) Receive(r HttpRequest, opts ...Opt) {
	log.Println("Start reciving")
	options := defaultPortOpts
	for _, o := range opts {
		o(&options)
	}

	select {
	case m.req <- r:
		return
	case <-time.Tick(options.timeout):
		log.Fatalf("Timeout, expected message %T not received\n", r)
	}
}

func (m *HTTPPort) Send(r HttpResponse, opts ...Opt) {
	go func() {
		m.resp <- r
	}()
	<-m.sync
	time.Sleep(time.Millisecond * 100)
}
