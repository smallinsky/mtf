package port

import (
	"testing"
	"time"
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

	pkgName  string
	err      error
	timeout  time.Duration
	TLSHosts []string

	t *testing.T
}

type PortOpt func(*portOpts)

func WithTLSCon(path string) PortOpt {
	return func(o *portOpts) {
		o.clientCertPath = path
	}
}

func WithTLS(crtPath, keyPath string) PortOpt {
	return func(o *portOpts) {
		o.serverCertPath = crtPath
		o.serverKeyPath = keyPath
	}
}

func WithTLSHost(hosts ...string) PortOpt {
	return func(o *portOpts) {
		o.TLSHosts = hosts
	}
}

var defaultPortOpts = portOpts{
	timeout: time.Second * 120,
}
