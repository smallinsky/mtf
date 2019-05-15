package port

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
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

type handler struct{}

func (p *HTTPPort) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p.reqC <- convHTTPRequest(req)

	resp := <-p.respC
	w.WriteHeader(resp.Status)
	w.Write([]byte(resp.Body))
	p.sync <- struct{}{}
}

func (p *HTTPPort) serve() error {

	p.svr = httptest.NewUnstartedServer(http.HandlerFunc(p.ServeHTTP))

	var err error
	p.svr.Listener, err = net.Listen("tcp", ":8080")
	if err != nil {
		return errors.Wrapf(err, "faield to start net listener")
	}
	p.svr.Start()

	srv := &http.Server{
		Addr:    ":8443",
		Handler: p,
	}
	go func() {
		log.Fatal(srv.ListenAndServeTLS(serverCertFile, serverKeyFile))
	}()
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

func (p *HTTPPort) Send(resp *HTTPResponse, opts ...Opt) error {
	resp.setDefaults()
	go func() {
		p.respC <- resp
	}()
	<-p.sync
	time.Sleep(time.Millisecond * 100)
	return nil
}

var (
	serverCertFile = "/tmp/mtf/server.crt"
	serverKeyFile  = "/tmp/mtf/server.key"
)

func init() {
	dir := filepath.Dir(serverCertFile)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll("/tmp/mtf", 0777); err != nil {
			panic(err)
		}
	}
	if err := ioutil.WriteFile(serverCertFile, serverCert, 0665); err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(serverKeyFile, serverKey, 0665); err != nil {
		panic(err)
	}
}

var serverCert = []byte(`-----BEGIN CERTIFICATE-----
MIIDLjCCAhYCCQD/jIo7GMKmYjANBgkqhkiG9w0BAQsFADBZMQswCQYDVQQGEwJQ
TDEQMA4GA1UECAwHV3JvY2xhdzEQMA4GA1UEBwwHV3JvY2xhdzEMMAoGA1UECgwD
TVRGMQwwCgYDVQQLDANNVEYxCjAIBgNVBAMMASowHhcNMTkwNTE1MjA1NjM0WhcN
MzkwNTEwMjA1NjM0WjBZMQswCQYDVQQGEwJQTDEQMA4GA1UECAwHV3JvY2xhdzEQ
MA4GA1UEBwwHV3JvY2xhdzEMMAoGA1UECgwDTVRGMQwwCgYDVQQLDANNVEYxCjAI
BgNVBAMMASowggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDZSuamylfv
vUuQw7jjjtc5sb0drwtidsuqMBMlKeznmajQjZb7TCRCDNo+d+JJ1SAVWaaVUzAQ
wj/MMKhm4tchS8ZYANZD65IKxFOZn7tp3sZzGL+X8VTdVgoyMYZd2H5O2HU8vkd8
SNOiYfzFsQtJbkFMbABwbDP5ZNSatIQ3x9QECBHg4We3o+UNqsvMw3PKCNjwPlrS
MhSzgbfJ1ZMH57w47PX3sGMyceJdeKhJqLNwH37nYFHkfXyo4to848SpAA79NouI
Gk+OhFXYJxRVpQQ/4D8LI/HTlZM9xFCaKDLcv9vRodpfUe/Jy9EzFAj9HfhI46zM
3ciS73BsVO9TAgMBAAEwDQYJKoZIhvcNAQELBQADggEBAF9IXL93AZWJV+krCbIn
pVY/PQXcNFADjbadjP7dEAy5rqfYSk3llcGr4ct2MynawtLQj84YNX9L1mQaUW0b
uAegiLP0+cDUCkCPCVdevbK/nGBH1caYRaiFJNqpihjuErGDm+zKeG9hRSCqRh5j
Oz+j225UxmZpL9/peJnOzNZY56zg9oswHJcEzN2paSeNgfmK68UDTrOr/Od/rn+A
6NpGSfXh1egl5GlHc2mHz3VyqrrXaQTkmHEqBGSnGjRFp0q4VQpRDPK0RaoOFddL
syGhC6L8SM/B793UEQR3HUcoYFYZX2fhHVcUc9BpD3VmegcJi5x65k7U/i3nXvyC
j3c=
-----END CERTIFICATE-----`)

var serverKey = []byte(`-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDZSuamylfvvUuQ
w7jjjtc5sb0drwtidsuqMBMlKeznmajQjZb7TCRCDNo+d+JJ1SAVWaaVUzAQwj/M
MKhm4tchS8ZYANZD65IKxFOZn7tp3sZzGL+X8VTdVgoyMYZd2H5O2HU8vkd8SNOi
YfzFsQtJbkFMbABwbDP5ZNSatIQ3x9QECBHg4We3o+UNqsvMw3PKCNjwPlrSMhSz
gbfJ1ZMH57w47PX3sGMyceJdeKhJqLNwH37nYFHkfXyo4to848SpAA79NouIGk+O
hFXYJxRVpQQ/4D8LI/HTlZM9xFCaKDLcv9vRodpfUe/Jy9EzFAj9HfhI46zM3ciS
73BsVO9TAgMBAAECggEABSfczypP6dVQ/K9YLLYP70ODXDfyCjUNYg1f9urGvzwL
IF+rrGzDE3ogl4jaqqvO5hLJfBOMOWmSf/LLnB1Xw2d73kyuyM/HGFBON3/tv3ZU
uRhmO2GzhMjs1wIL0SA45wAF0BonshA8TUcL61jnDqf6DqklXYWDujAlR0JvPK+K
4pfNwPLT3czAwb+t9OmIDZLJ7xo8q7uXD6E9lhHvBYJdy/BdYUeA7cPPFJEAFZ0K
3Kv3Zp9wnGlJJY/uIfT6IgViUkRswRn+dP2bq+8+A8YWX2/ifQ6swFkLlxnSiJqZ
PuNFzGmc7vYYbPcv1GuM3kRIdW0wocHazJm8VZVCmQKBgQD/y1GTnfn3D7SXRLz7
O90ec+K3OJobfkkNDKYmtP5H1cazvQzrH9U319Ngsu3W4RI+h2HWTfN+BAVCJlTy
+IC8yCSFzoDK5imnCPBupkJXypZzhVknOdjq9GjC+B8OqfA/3ez4Sucs9gR77yfz
ifoxlEhD9SBONJ1iCr3Ns7iqrwKBgQDZd6cgIjYp7SRjRNLdJSa1oMCizirzxHlx
zVTdFuVFsKpy8wp25jK352vTbPUP3fK8Z90XTGeFX2qjE0BprEAdR9S6yyaFR29U
5GTf3KEU8m2IaShxYsyfUM2q4H6hMGcO5wgdsUz6CNMRj02iIQ3GpbkanEIgCz98
WlawKxhenQKBgErwRfX5UkIPV9j5SmRQJXfGe6Ux7/QeC0jHa+XrIJPrDUubFy3L
Jaw2jrbFtOg/CBlJkGA4dh11EBVRJZIJO64S9KA+33yR8aH9/HJuQwF1WJ5/cp8L
U4GCGS8FghPJtZkAa2xShWemq6mjZxDyW1orFwDRz6UZxQH0I6cf//oBAoGALpuk
aBCtByNaLyRrBRaXS0oev0Xskr5DQQ6+53umu973SRep4H3J1Px2caPiifoJsjOY
gQvRDBa9JiJUJdHTE/N3Nmmf4eTDibBBpnEE3RZwP1I6ZsLEFEkfK0ZeHXHgRKNj
a+m6E8ScaCEMhHkNGMwf9gITcga3HpHGDo/N80kCgYEApUBr6+x90nhDUV8IntNB
JHr0d0Tp8D2C2EjleGfQ1AuNcheCD+BdD8Exn4Hot0wxZIrE9PF1AsQC1ATIJgfR
XC5VgDfPmzTG+u+pyqqMN1/pPCPQABz3j7HGeTObnrumnXmHgBLbFxvkkNEgZDEV
7aZxkbcqsPRzeI8UThNm9BU=
-----END PRIVATE KEY-----`)
