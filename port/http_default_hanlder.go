package port

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func NewHTTP2Port() *Port {
	startHTTP()
	return &Port{
		impl: ht.httpPort2,
	}
}

func NewGCSPort() (*Port, error) {
	startHTTP()
	return &Port{
		impl: ht.gcs,
	}, nil
}
func newHTTPPort2() *HTTPPort2 {
	return &HTTPPort2{
		reqC:  make(chan *HTTPRequest),
		respC: make(chan *HTTPResponse),
	}
}

type HTTPPort2 struct {
	reqC  chan *HTTPRequest
	respC chan *HTTPResponse
	sync  chan struct{}
}

type HTTPRequest struct {
	Body []byte
	//URL    *url.URL
	Method string
	Host   string
	URL    string
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
		//URL:    r.URL,
		Body: buff,
		Host: r.Host,
		URL:  r.URL.RequestURI(),
	}

	return out
}

func (resp *HTTPResponse) setDefaults() {
	if resp.Status == 0 {
		resp.Status = http.StatusOK
	}
}

func (p *HTTPPort2) Register(router *mux.Router) {
	router.NotFoundHandler = p
}

func (p *HTTPPort2) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p.reqC <- convHTTPRequest(req)

	resp := <-p.respC
	w.WriteHeader(resp.Status)
	w.Write([]byte(resp.Body))
}

func (p *HTTPPort2) receive(opts ...Opt) (*HTTPRequest, error) {
	options := defaultPortOpts
	for _, o := range opts {
		o(&options)
	}

	select {
	case req := <-p.reqC:
		return req, nil
	case <-time.Tick(options.timeout):
		return nil, errors.Errorf("failed to receive  message, deadline exeeded")
	}
}

func (p *HTTPPort2) send(msg *HTTPResponse, opts ...Opt) error {
	options := defaultPortOpts
	for _, opt := range opts {
		opt(&options)
	}

	msg.setDefaults()
	go func() {
		p.respC <- msg
	}()
	return nil
}

func (p *HTTPPort2) Send(ctx context.Context, i interface{}) error {
	resp, ok := i.(*HTTPResponse)
	if !ok {
		return errors.Errorf("invalid type %T", i)
	}
	return p.send(resp)
}

func (p *HTTPPort2) Receive(ctx context.Context) (interface{}, error) {
	return p.receive()
}
