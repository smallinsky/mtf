package port

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
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
	for _, m := range d.MethodsDesc {
		port.emd[m.InType] = EndpointRespTypePair{
			RespType: m.OutType,
			Endpoint: d.Name + "/" + m.Name,
		}
	}
	if err := port.connect(target, options.clientCertPath); err != nil {
		return nil, errors.Wrapf(err, "failed to connect")
	}
	return port, nil
}

func NewGRPCClientPort(i interface{}, target string, opts ...PortOpt) (*Port, error) {
	c, err := NewGRPCClient(i, target, opts...)
	if err != nil {
		return nil, err
	}
	return &Port{
		impl: c,
	}, nil
}

type connection interface {
	Invoke(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error
	Close() error
	GetState() connectivity.State
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

type res struct {
	err error
	msg interface{}
}

func (p *ClientPort) Receive(ctx context.Context) (interface{}, error) {
	return p.receive()
}

func (p *ClientPort) Send(ctx context.Context, msg interface{}) error {
	return p.send(ctx, msg)
}

func (p *ClientPort) receive(opts ...PortOpt) (interface{}, error) {
	options := defaultPortOpts
	for _, o := range opts {
		o(&options)
	}

	select {
	case <-time.Tick(options.timeout):
		return nil, errors.Errorf("failed to receive  message, deadline exeeded %s")
	case result := <-p.callResultC:
		if result.err != nil {
			return nil, result.err
		}
		return result.resp, nil
	}
}

func (p *ClientPort) send(ctx context.Context, msg interface{}) error {
	v, ok := p.emd[reflect.TypeOf(msg)]
	if !ok {
		return errors.Errorf("port doesn't support message type %T", msg)
	}
	stop := make(chan struct{})
	go func() {
		for {
			state := p.conn.GetState()
			fmt.Printf("PORT state: %s\n", state)
			time.Sleep(time.Second)
			select {
			case <-stop:
				return
			default:
				continue
			}
		}
	}()
	go func() {
		defer func() {
			close(stop)
		}()
		out := reflect.New(v.RespType.Elem()).Interface()
		fmt.Printf("sending %T ... \n", msg)
		if err := p.conn.Invoke(ctx, v.Endpoint, msg, out, grpc.WaitForReady(true)); err != nil {
			fmt.Printf("sending %T error \n", msg)
			go func() {
				p.callResultC <- callResult{
					err:  err,
					resp: nil,
				}
			}()
			return
		}

		fmt.Printf("sending %T OK \n", msg)
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
	}()
	return nil
}
