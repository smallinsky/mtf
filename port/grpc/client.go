package grpc

import (
	"context"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-test/deep"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type EndpointRespTypePair struct {
	RespType reflect.Type
	Endpoint string
}

type MsgTypeMap map[reflect.Type]EndpointRespTypePair

func NewClient(i interface{}, target string, opts ...PortOpt) ClientPort {
	options := defaultPortOpts
	for _, o := range opts {
		o(&options)
	}
	p := ClientPort{
		emd: make(map[reflect.Type]EndpointRespTypePair),
		out: make(chan interface{}, 1),
	}

	d := getGrpcDetails(i)
	for _, m := range d.methodsDesc {
		p.emd[m.InType] = EndpointRespTypePair{
			RespType: m.OutType,
			Endpoint: d.Name + "/" + m.Name,
		}
		log.Printf("Endpoint url: %s\n", p.emd[m.InType].Endpoint)
	}
	p.connect(target, options.clientCertPath)
	return p
}

type connection interface {
	Invoke(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error
	Close() error
}

type ClientPort struct {
	conn connection

	emd     MsgTypeMap
	out     chan interface{}
	sendMtx sync.Mutex
}

func (p *ClientPort) connect(addr, certfile string) {
	options := []grpc.DialOption{grpc.WithInsecure()}
	if certfile != "" {
		// TODO: set dynamic authority header file.
		creds, err := credentials.NewClientTLSFromFile(certfile, strings.Split(addr, ":")[0])
		if err != nil {
			log.Fatalf("Failed to load credentials: %s", err)
		}
		options[0] = grpc.WithTransportCredentials(creds)
	}
	var err error
	p.conn, err = grpc.Dial(addr, options...)
	if err != nil {
		log.Fatal("Failed to dial target address: ", err)
		p.Close()
	}
}

func (p *ClientPort) Close() {
	p.conn.Close()
}

func (p *ClientPort) Send(msg interface{}) {
	startSync.Wait()
	go func() {
		done := make(chan struct{})
		go func() {
			defer func() {
				done <- struct{}{}

			}()

			v, ok := p.emd[reflect.TypeOf(msg)]
			if !ok {
				log.Fatalf("Failed to map type %T to endpoint url", msg)
			}

			out := reflect.New(v.RespType.Elem()).Interface()
			if err := p.conn.Invoke(context.Background(), v.Endpoint, msg, out); err != nil {
				log.Fatalf("Failed to invoke: %v", err)
			}

			var resp interface{}
			rv := reflect.ValueOf(&resp)
			rv.Elem().Set(reflect.New(v.RespType))
			rv.Elem().Set(reflect.ValueOf(out))
			p.out <- resp
		}()

		deadline := time.Tick(time.Second * 2)
		select {
		case <-deadline:
			log.Fatalf("Failed to send %T message deadline exeeded", msg)
		case <-done:
			return
		}
	}()
}

func (p *ClientPort) Receive(msg interface{}, opts ...PortOpt) {
	options := defaultPortOpts
	for _, o := range opts {
		o(&options)
	}

	deadlineC := time.Tick(options.timeout)

	select {
	case <-deadlineC:
		log.Fatalf("Deadline during receving %T message", msg)
	case m := <-p.out:
		//TODO Use template pattern matching
		if err := deep.Equal(msg, m); err != nil {
			log.Fatalf("Struct not eq: %v", err)
		}
	}
}
