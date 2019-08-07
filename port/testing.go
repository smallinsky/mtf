package port

import (
	"github.com/smallinsky/mtf/match"
	"testing"
)

type Kind int

const (
	KIND_SERVER Kind = 1
	KIND_CLIENT      = iota
)

type SenderReceiverKinder interface {
	Send(interface{}) error
	Receive() (interface{}, error)
	Kind() Kind
}

type Port struct {
	sck     SenderReceiverKinder
	message chan interface{}
}

func (p *Port) Send(msg interface{}) error {
	p.sck.Send(msg)
	return nil
}

func (p *Port) Receive(msg interface{}) error {
	m, err := p.sck.Receive()
	if err != nil {
		return err
	}
	m = m
	return nil
}

func (port *ClientPort) SendT(t *testing.T, msg interface{}) {
	port.send(msg)
}

func (port *ClientPort) ReceiveT(t *testing.T, msg interface{}, opts ...PortOpt) {
	port.receive(opts...)
	//if err := port.Receive(msg, opts...); err != nil {
	//	t.Fatalf("failed to receive %T, error: %v", msg, err)
	//}
}

func (port *ClientPort) ReceiveTM(t *testing.T, m match.Matcher, opts ...PortOpt) {
	//	if err := port.ReceiveM(m, opts...); err != nil {
	//		t.Fatalf("failed to receive, error: %v", err)
	//	}
}

func (p *PortIn) ReceiveT(t *testing.T, i interface{}, opts ...Opt) {
	p.receive(opts...)
}

func (p *PortIn) ReceiveTM(t *testing.T, m match.Matcher, opts ...Opt) {
	p.receive(opts...)
	//if err := p.ReceiveM(m, opts...); err != nil {
	//	t.Fatalf("failed to receive, error: %v", err)
	//}
}

func (p *PortIn) SendT(t *testing.T, msg interface{}, opts ...PortOpt) {
	p.send(msg, opts...)
	//	if err := p.Send(msg, opts...); err != nil {
	//		t.Fatalf("failed to send %T, error: %v", msg, err)
	//	}
}

func (p *HTTPPort) SendT(t *testing.T, resp *HTTPResponse, opts ...Opt) {
	p.send(resp, opts...)
	//	if err := p.Send(resp, append([]Opt{WithT(t)}, opts...)...); err != nil {
	//		t.Fatalf("faield to send, error %v", err)
	//	}
}

func (p *HTTPPort) ReceiveT(t *testing.T, req *HTTPRequest, opts ...Opt) {
	p.receive(opts...)
	//	if err := p.Receive(req, append([]Opt{WithT(t)}, opts...)...); err != nil {
	//		t.Fatalf("faield to receive, error %v", err)
	//	}
}

func (p *HTTPPort) ReceiveTM(t *testing.T, m match.Matcher, opts ...Opt) {
	p.receive(opts...)
	//if err := p.ReceiveM(m, opts...); err != nil {
	//	t.Fatalf("faield to receive, error %v", err)
	//}
}
