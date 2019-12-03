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

func NewHTTPPort() *Port {
	startHTTP()
	return &Port{
		impl: ht.httpPort,
	}
}

func NewGCSPort() (*Port, error) {
	startHTTP()
	return &Port{
		impl: ht.gcs,
	}, nil
}
func newHTTPPort() *HTTPPort {
	return &HTTPPort{
		reqC:  make(chan *HTTPRequest),
		respC: make(chan *HTTPResponse),
	}
}

type HTTPPort struct {
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
		Body:   buff,
		Host:   r.Host,
		URL:    r.URL.RequestURI(),
	}

	if len(out.Body) == 0 {
		out.Body = nil
	}

	return out
}

func (resp *HTTPResponse) setDefaults() {
	if resp.Status == 0 {
		resp.Status = http.StatusOK
	}
}

func (p *HTTPPort) Register(router *mux.Router) {
	router.NotFoundHandler = p
}

func (p *HTTPPort) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p.reqC <- convHTTPRequest(req)

	resp := <-p.respC
	w.WriteHeader(resp.Status)
	w.Write([]byte(resp.Body))
}

func (p *HTTPPort) receive(opts ...Opt) (*HTTPRequest, error) {
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

func (p *HTTPPort) send(msg *HTTPResponse, opts ...Opt) error {
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

func (p *HTTPPort) Send(ctx context.Context, i interface{}) error {
	resp, ok := i.(*HTTPResponse)
	if !ok {
		return errors.Errorf("invalid type %T", i)
	}
	return p.send(resp)
}

func (p *HTTPPort) Receive(ctx context.Context) (interface{}, error) {
	return p.receive()
}
