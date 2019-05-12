package port

import (
	"context"
	"log"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/smallinsky/mtf/match"
)

type EndpointRespTypePair struct {
	RespType reflect.Type
	Endpoint string
}

type MsgTypeMap map[reflect.Type]EndpointRespTypePair

func NewGRPCClient(i interface{}, target string, opts ...PortOpt) (*ClientPort, error) {
	options := defaultPortOpts
	for _, o := range opts {
		o(&options)
	}
	port := &ClientPort{
		emd:         make(map[reflect.Type]EndpointRespTypePair),
		callResultC: make(chan callResult, 1),
	}

	d, err := getGrpcDetails(i)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get grpc details")
	}
	for _, m := range d.methodsDesc {
		port.emd[m.InType] = EndpointRespTypePair{
			RespType: m.OutType,
			Endpoint: d.Name + "/" + m.Name,
		}
		log.Printf("Endpoint url: %s\n", port.emd[m.InType].Endpoint)
	}
	if err := port.connect(target, options.clientCertPath); err != nil {
		return nil, errors.Wrapf(err, "failed to connect")
	}
	return port, nil
}

type connection interface {
	Invoke(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error
	Close() error
}

type ClientPort struct {
	conn connection

	emd         MsgTypeMap
	sendMtx     sync.Mutex
	callResultC chan callResult
}

type callResult struct {
	resp interface{}
	err  error
}

func (p *ClientPort) connect(addr, certfile string) error {
	options := []grpc.DialOption{grpc.WithInsecure()}
	if certfile != "" {
		// TODO: set dynamic authority header file.
		creds, err := credentials.NewClientTLSFromFile(certfile, strings.Split(addr, ":")[0])
		if err != nil {
			return errors.Wrapf(err, "failed to load cert from file %v", certfile)
		}
		options[0] = grpc.WithTransportCredentials(creds)
	}
	var err error
	c, err := grpc.Dial(addr, options...)
	if err != nil {
		return errors.Wrapf(err, "failed to dial %s", addr)
	}
	p.conn = c
	return nil
}

func (p *ClientPort) Close() {
	p.conn.Close()
}

func (p *ClientPort) Send(msg interface{}) error {
	startSync.Wait()
	errC := make(chan error)

	go func() {
		v, ok := p.emd[reflect.TypeOf(msg)]
		if !ok {
			errC <- errors.Errorf("port dosn't support message type %T", msg)
			return
		}

		out := reflect.New(v.RespType.Elem()).Interface()
		if err := p.conn.Invoke(context.Background(), v.Endpoint, msg, out); err != nil {
			go func() {
				p.callResultC <- callResult{
					err:  err,
					resp: nil,
				}
			}()
			errC <- errors.Wrapf(err, "failed to send %T message", msg)
			return
		}

		var resp interface{}
		rv := reflect.ValueOf(&resp)
		rv.Elem().Set(reflect.New(v.RespType))
		rv.Elem().Set(reflect.ValueOf(out))
		go func() {
			p.callResultC <- callResult{
				err:  nil,
				resp: resp,
			}
		}()
		errC <- nil
	}()

	select {
	case <-time.Tick(time.Second * 5):
		return errors.Errorf("failed to send %T message, deadline exeeded", msg)
	case err := <-errC:
		return err
	}
}

func (p *ClientPort) ReceiveM(m match.Matcher, opts ...PortOpt) error {
	options := defaultPortOpts
	for _, o := range opts {
		o(&options)
	}

	if err := m.Validate(); err != nil {
		return errors.Wrapf(err, "invalid marcher argument")
	}

	select {
	case result := <-p.callResultC:
		if err := m.Match(result.err, result.resp); err != nil {
			return errors.Wrapf(err, "%T message match failed", m)
		}
	case <-time.Tick(options.timeout):
		return errors.Errorf("failed to receive  message, deadline exeeded")
	}
	return nil
}

func (p *ClientPort) Receive(msg interface{}, opts ...PortOpt) error {
	options := defaultPortOpts
	for _, o := range opts {
		o(&options)
	}

	deadlineC := time.Tick(options.timeout)

	select {
	case <-deadlineC:
		return errors.Errorf("failed to receive  message, deadline exeeded")
	case result := <-p.callResultC:
		if result.err != nil {
			return errors.Wrapf(result.err, "Got unexpected error during receive, err: %v", result.err)
		}
		if diff := deep.Equal(msg, result.resp); diff != nil {
			return errors.Errorf("struct not eq:\n diff '%s'\n", diff)
		}
	}
	return nil
}

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
