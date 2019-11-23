package port

import (
	"net"
	"sync"
)

// Custom proxy around net.Listen interface that allow to
// observe when client connection is enstablished.
// Current GRPC transport impementation does retry with exponential backoff
// and when connection is not in READY state whole GRPC call fails with error.

var _ net.Listener = &listener{}

var startSync sync.WaitGroup

func WaitForGRPCConn() {
	startSync.Wait()
}

func listen(network, address string) (net.Listener, error) {
	defer startSync.Add(1)
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
	l  net.Listener
	wg sync.WaitGroup
}

func (l *listener) Accept() (net.Conn, error) {
	defer startSync.Done()
	return l.l.Accept()
}

func (l *listener) Close() error {
	return l.l.Close()
}

func (l *listener) Addr() net.Addr {
	return l.l.Addr()
}
