package port

import (
	"context"
	"fmt"
	"strings"

	mtfctx "github.com/smallinsky/mtf/framework/context"
	"github.com/smallinsky/mtf/match"
)

type PortImpl interface {
	Send(ctx context.Context, msg interface{}) error
	Receive(ctx context.Context) (interface{}, error)
}

type Port struct {
	impl PortImpl
}

type sendOptions struct {
	ctx context.Context
}

type SendOption func(*sendOptions)

func WithCtx(ctx context.Context) SendOption {
	return func(o *sendOptions) {
		o.ctx = ctx
	}
}

func (p *Port) Send(i interface{}, opts ...SendOption) error {
	t := getT()
	defOpts := &sendOptions{
		ctx: context.Background(),
	}

	for _, o := range opts {
		o(defOpts)
	}

	name := getPortName(p.impl)
	if err := p.impl.Send(defOpts.ctx, i); err != nil {
		t.Fatalf("failed to send %T from %s, err: %v", i, name, err)
	}

	if mtfc := mtfctx.Get(t); mtfc != nil {
		mtfc.LogSend(name, i)
	}
	return nil
}

func getPortName(i interface{}) string {
	name := fmt.Sprintf("%T", i)
	return fmt.Sprintf("%s", strings.ToLower(name))
}

func (p *Port) Receive(i interface{}) (interface{}, error) {
	t := getT()
	ctx := context.Background()
	m, err := p.impl.Receive(ctx)

	name := getPortName(p.impl)

	if matcher, ok := i.(*match.GRPCErrType); ok {
		if mtfc := mtfctx.Get(t); mtfc != nil {
			mtfc.LogReceive(name, err)
		}
		if err := matcher.Match(err); err != nil {
			t.Fatalf("Failed to receive GRPC error: %v", err)
		}
		return i, nil
	}

	if mtfc := mtfctx.Get(t); mtfc != nil {
		mtfc.LogReceive(name, m)
	}

	if err != nil {
		t.Fatalf("failed to receive %T from %s: %v", i, name, err)
	}

	switch t := i.(type) {
	case *match.FnType:
		err = t.Match(err, m)
	case *match.TypeT:
		err = t.Match(m)
	case *match.DeepEqualType:
		err = t.Match(m)
	case *match.PayloadMatcher:
		err = t.Match(err, m)
	case *match.ProtoEqualType:
		err = t.Match(m)
	case *match.DiffType:
		err = t.Match(m)
	default:
		err = match.DeepEqual(i).Match(m)
	}

	if err != nil {
		t.Fatalf("Failed to receive %T:\n %v", i, err)
	}

	return m, nil
}
