package port

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

func NewHTTPPort(options ...PortOpt) (*Port, error) {
	p, err := NewHTTP(options...)
	if err != nil {
		return nil, err
	}
	return &Port{
		impl: p,
	}, nil
}

func (p *HTTPPort) Kind() Kind {
	return KIND_SERVER
}

func (p *HTTPPort) Name() string {
	return "http_server"
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

//TODO Add https support
func NewHTTP(options ...PortOpt) (*HTTPPort, error) {
	port := &HTTPPort{
		reqC:  make(chan *HTTPRequest),
		respC: make(chan *HTTPResponse),
		sync:  make(chan struct{}),
	}

	defaultOption := defaultPortOpts
	for _, option := range options {
		option(&defaultOption)
	}

	if err := port.serveHTTPS(defaultOption.TLSHosts); err != nil {
		return nil, errors.Wrapf(err, "failed to serv https")
	}

	if err := port.serveHTTP(); err != nil {
		return nil, errors.Wrapf(err, "failed to serv http")
	}
	return port, nil
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

func (p *HTTPPort) serveHTTP() error {
	p.svr = httptest.NewUnstartedServer(http.HandlerFunc(p.ServeHTTP))

	var err error
	p.svr.Listener, err = net.Listen("tcp", ":8080")
	if err != nil {
		return errors.Wrapf(err, "failed to start net listener")
	}
	p.svr.Start()

	return nil
}

func (p *HTTPPort) serveHTTPS(hosts []string) error {
	ck, err := genCertForHost(hosts)
	if err != nil {
		return err
	}
	_, err = tls.X509KeyPair(ck.Cert, ck.Key)
	if err != nil {
		return fmt.Errorf("Faield to verify keyPair err: %v\n", err)
	}

	if err := writeCert(ck); err != nil {
		return err
	}

	srv := &http.Server{
		Addr:    ":8443",
		Handler: p,
	}
	go func() {
		if err := srv.ListenAndServeTLS(serverCertFile, serverKeyFile); err != nil {
			log.Fatalf("faield to start tls server: %v", err)
		}
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
	<-p.sync
	time.Sleep(time.Millisecond * 50)
	return nil
}

var (
	serverCertFile = "/tmp/mtf/cert/server.crt"
	serverKeyFile  = "/tmp/mtf/cert/server.key"
)

func writeCert(ck *CertKey) error {
	dir := filepath.Dir(serverCertFile)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return err
	}
	if err := ioutil.WriteFile(serverCertFile, ck.Cert, 0665); err != nil {
		return err
	}
	if err := ioutil.WriteFile(serverKeyFile, ck.Key, 0665); err != nil {
		return err
	}
	return nil
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

type CertKey struct {
	Cert []byte
	Key  []byte
}

func genCertForHost(hosts []string) (*CertKey, error) {
	var priv interface{}
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %s", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %s", err)
	}

	now := time.Now()
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme"},
		},
		NotBefore: now,
		NotAfter:  now.Add(365 * 24 * time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts = append(hosts, []string{"localhost", "host.docker.internal"}...)
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	template.IsCA = true
	template.KeyUsage |= x509.KeyUsageCertSign

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		return nil, fmt.Errorf("Failed to create certificate: %s", err)
	}

	var certBuff bytes.Buffer
	if err := pem.Encode(&certBuff, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return nil, fmt.Errorf("failed to write data to cert.pem: %s", err)
	}

	var keyBuff bytes.Buffer
	if err := pem.Encode(&keyBuff, pemBlockForKey(priv)); err != nil {
		return nil, fmt.Errorf("failed to write data to key.pem: %s", err)
	}

	return &CertKey{
		Cert: certBuff.Bytes(),
		Key:  keyBuff.Bytes(),
	}, nil
}

func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}
