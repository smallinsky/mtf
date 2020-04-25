package netw

import (
	"net"
	"sync"
)

// Custom proxy around net.Listen interface that allow to
// observe when client connection is established.
// Current GRPC transport implementation does retry with exponential backoff
// and when connection is not in READY state whole GRPC call fails with error.

var _ net.Listener = &listener{}

var startSync sync.WaitGroup

func WaitForGRPCConn() {
	startSync.Wait()
}

func Listen(network, address string) (net.Listener, error) {
	l, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	defer startSync.Add(1)
	lProxy := &listener{
		l: l,
	}
	return lProxy, err
}

type listener struct {
	l  net.Listener
	wg sync.WaitGroup
}

func (l *listener) Accept() (net.Conn, error) {
	conn, err := l.l.Accept()
	if err != nil {
		return nil, err
	}
	defer startSync.Done()
	return conn, nil
}

func (l *listener) Close() error {
	return l.l.Close()
}

func (l *listener) Addr() net.Addr {
	return l.l.Addr()
}
