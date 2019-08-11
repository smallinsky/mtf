package port

import (
	"testing"

	"github.com/smallinsky/mtf/match"
)

type Kind int

const (
	KIND_SERVER Kind = 1
	KIND_CLIENT      = iota
)

type PortImpl interface {
	Send(interface{}) error
	Receive() (interface{}, error)
	Kind() Kind
	Name() string
}

type Port struct {
	impl PortImpl
}

func (p *Port) Send(t *testing.T, i interface{}) error {
	if err := p.impl.Send(i); err != nil {
		t.Fatalf("failed to send %T from %s, err: %v", i, p.impl.Name(), err)
	}
	//t.Logf("[%v: %T] --> [SUT]\n", p.impl.Name(), i)
	return nil
}

func (p *Port) Receive(t *testing.T, i interface{}) error {
	m, err := p.impl.Receive()

	switch t := i.(type) {
	case match.FnMatcher:
		t.Match(err, m)
	case match.Any:
	default:
	}

	//t.Logf("[SUT: %T] --> [%s]\n", i, p.impl.Name())

	return nil
}