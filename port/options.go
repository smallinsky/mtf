package port

import (
	"testing"
	"time"

	"github.com/smallinsky/mtf/pkg/cert"
)

type Opt func(*portOpts)

func WithError(err error) Opt {
	return func(o *portOpts) {
		o.err = err
	}
}

func WithT(t *testing.T) Opt {
	return func(o *portOpts) {
		o.t = t
	}
}

func WithTimeout(timout time.Duration) Opt {
	return func(o *portOpts) {
		o.timeout = timout
	}
}

type portOpts struct {
	clientCertPath string

	serverCertPath string
	serverKeyPath  string

	pkgName string
	err     error
	timeout time.Duration

	t *testing.T
}

type PortOpt func(*portOpts)

func WithTLS() PortOpt {
	return func(o *portOpts) {
		o.clientCertPath = cert.ServerCertFile
		o.serverCertPath = cert.ServerCertFile
		o.serverKeyPath = cert.ServerKeyFile
	}
}

var defaultPortOpts = portOpts{
	timeout: time.Second * 15,
}
