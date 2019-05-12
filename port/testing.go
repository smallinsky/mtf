package port

import (
	"testing"

	"github.com/smallinsky/mtf/match"
)

func (port *ClientPort) SendT(t *testing.T, msg interface{}) {
	if err := port.Send(msg); err != nil {
		t.Fatalf("failed to send %T, error: %v", msg, err)
	}
}

func (port *ClientPort) ReceiveT(t *testing.T, msg interface{}, opts ...PortOpt) {
	if err := port.Receive(msg, opts...); err != nil {
		t.Fatalf("failed to receive %T, error: %v", msg, err)
	}
}

func (port *ClientPort) ReceiveTM(t *testing.T, m match.Matcher, opts ...PortOpt) {
	if err := port.ReceiveM(m, opts...); err != nil {
		t.Fatalf("failed to receive, error: %v", err)
	}
}

func (p *PortIn) ReceiveT(t *testing.T, i interface{}, opts ...Opt) {
	if err := p.Receive(i, opts...); err != nil {
		t.Fatalf("failed to receive, error: %v", err)
	}
}

func (p *PortIn) ReceiveTM(t *testing.T, m match.Matcher, opts ...Opt) {
	if err := p.ReceiveM(m, opts...); err != nil {
		t.Fatalf("failed to receive, error: %v", err)
	}
}

func (p *PortIn) SendT(t *testing.T, msg interface{}, opts ...PortOpt) {
	if err := p.Send(msg, opts...); err != nil {
		t.Fatalf("failed to send %T, error: %v", msg, err)
	}
}

func (p *HTTPPort) SendT(t *testing.T, resp *HTTPResponse, opts ...Opt) {
	if err := p.Send(resp, opts...); err != nil {
		t.Fatalf("faield to send, error %v", err)
	}
}

func (p *HTTPPort) ReceiveT(t *testing.T, req *HTTPRequest, opts ...Opt) {
	if err := p.Receive(req, opts...); err != nil {
		t.Fatalf("faield to receive, error %v", err)
	}
}

func (p *HTTPPort) ReceiveTM(t *testing.T, m match.Matcher, opts ...Opt) {
	if err := p.ReceiveM(m, opts...); err != nil {
		t.Fatalf("faield to receive, error %v", err)
	}
}
