package port

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"github.com/smallinsky/mtf/match"
)

//TODO Add https support
func NewHTTP() (*HTTPPort, error) {
	port := &HTTPPort{
		reqC:  make(chan *HTTPRequest),
		respC: make(chan *HTTPResponse),
		sync:  make(chan struct{}),
	}
	if err := port.serve(); err != nil {
		return nil, errors.Wrapf(err, "failed to serv")
	}
	return port, nil
}

type HTTPRequest struct {
	Body   []byte
	URL    *url.URL
	Method string
}

type HTTPResponse struct {
	Body   []byte
	Status int
}

func convHTTPRequest(r *http.Request) *HTTPRequest {
	if r == nil {
		return nil
	}

	defer r.Body.Close()
	buff, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("failed to read request body content, err: %v\n", err)
	}

	out := &HTTPRequest{
		Method: r.Method,
		URL:    r.URL,
		Body:   buff,
	}

	return out
}

func (resp *HTTPResponse) setDefaults() {
	if resp.Status == 0 {
		resp.Status = http.StatusOK
	}
}

type HTTPPort struct {
	reqC  chan *HTTPRequest
	respC chan *HTTPResponse
	sync  chan struct{}

	svr *httptest.Server
}

func (p *HTTPPort) serve() error {
	p.svr = httptest.NewUnstartedServer(http.HandlerFunc(p.Handle))

	var err error
	p.svr.Listener, err = net.Listen("tcp", ":8080")
	if err != nil {
		return errors.Wrapf(err, "faield to start net listener")
	}
	p.svr.Start()
	return nil
}

func (p *HTTPPort) Stop() {
	p.svr.Close()
}

func (p *HTTPPort) Handle(w http.ResponseWriter, req *http.Request) {
	p.reqC <- convHTTPRequest(req)

	resp := <-p.respC
	w.WriteHeader(resp.Status)
	w.Write([]byte(resp.Body))
	p.sync <- struct{}{}
}

func (p *HTTPPort) Receive(r *HTTPRequest, opts ...Opt) error {
	options := defaultPortOpts
	for _, o := range opts {
		o(&options)
	}

	select {
	case req := <-p.reqC:
		// Add matcher
		log.Printf("[DEBUG]: %T Received %v", p, req)
	case <-time.Tick(options.timeout):
		return errors.Errorf("failed to receive  message, deadline exeeded")
	}

	return nil
}

func (p *HTTPPort) ReceiveM(m match.Matcher, opts ...Opt) error {
	options := defaultPortOpts
	for _, o := range opts {
		o(&options)
	}
	if err := m.Validate(); err != nil {
		return errors.Wrapf(err, "invalid marcher argument")
	}

	select {
	case req := <-p.reqC:
		if err := m.Match(nil, req); err != nil {
			return errors.Wrapf(err, "%T message match failed", m)
		}
	case <-time.Tick(options.timeout):
		return errors.Errorf("failed to receive  message, deadline exeeded")
	}
	return nil
}

func (m *HTTPPort) Send(resp *HTTPResponse, opts ...Opt) error {
	resp.setDefaults()
	go func() {
		m.respC <- resp
	}()
	<-m.sync
	time.Sleep(time.Millisecond * 100)
	return nil
}
