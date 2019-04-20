package grpc

import (
	"net"
	"sync"
)

// Custom proxy around net.Listen interface that allow to
// observe when client connection is enstablished.
// Current GRPC transport impementation does retry with exponencial backoff
// and when connection is not in READY state whole GRPC call fails with error.

var _ net.Listener = &listener{}

var startSync sync.WaitGroup

func listen(network, address string) (net.Listener, error) {
	l, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}

	lProxy := &listener{
		l: l,
	}

	return lProxy, err
}

type listener struct {
	l net.Listener
}

func (l *listener) Accept() (net.Conn, error) {
	conn, err := l.l.Accept()
	if err != nil {
		return conn, err
	}

	startSync.Done()
	return conn, err
}

func (l *listener) Close() error {
	return l.l.Close()
}

func (l *listener) Addr() net.Addr {
	return l.l.Addr()
}
