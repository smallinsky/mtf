package port

import (
	"time"
)

type Opt func(*portOpts)

func WithError(err error) Opt {
	return func(o *portOpts) {
		o.err = err
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

func WithPkgName(name string) PortOpt {
	return func(o *portOpts) {
		o.pkgName = name
	}
}

func WithTLSHost(hosts ...string) PortOpt {
	return func(o *portOpts) {
		o.TLSHosts = hosts
	}
}

var defaultPortOpts = portOpts{
	// TODO: build dynamically base on proto package name.
	pkgName: "",
	timeout: time.Second * 5,
}
