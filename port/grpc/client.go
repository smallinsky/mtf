package grpc

import (
	"context"
	"log"
	"reflect"
	"strings"

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
	}

	d := getGrpcDetails(i)
	for _, m := range d.methodsDesc {
		p.emd[m.InType] = EndpointRespTypePair{
			RespType: m.OutType,
			Endpoint: strings.Join([]string{options.pkgName, d.Name}, ".") + "/" + m.Name,
		}
		log.Printf("Endpoint url: %s\n", p.emd[m.InType].Endpoint)
	}
	p.connect(target, options.clientCertPath)
	return p
}

type ClientPort struct {
	conn *grpc.ClientConn

	resp interface{}
	emd  MsgTypeMap
}

func (p *ClientPort) connect(addr, certfile string) {
	options := []grpc.DialOption{grpc.WithInsecure()}
	if certfile != "" {
		creds, err := credentials.NewClientTLSFromFile(certfile, "service-labels") //strings.Split(addr, ":")[0])
		if err != nil {
			log.Fatalf("Failed to load credentials: %s", err)
		}
		options[0] = grpc.WithTransportCredentials(creds)
	}
	var err error
	log.Println("dial: ", addr)
	p.conn, err = grpc.Dial(addr, options...)
	if err != nil {
		log.Fatal("Failed to dial target address: ", err)
	}
}

func (p *ClientPort) Close() {
	p.conn.Close()
}

func (p *ClientPort) Send(msg interface{}) {
	v, ok := p.emd[reflect.TypeOf(msg)]
	if !ok {
		log.Fatalln("Failed to map type %T to endpoint url")
	}

	out := reflect.New(v.RespType.Elem()).Interface()
	if err := p.conn.Invoke(context.Background(), v.Endpoint, msg, out); err != nil {
		log.Fatalf("Failed to invoke: %v", err)
	}

	rv := reflect.ValueOf(&p.resp)
	rv.Elem().Set(reflect.New(v.RespType))
	rv.Elem().Set(reflect.ValueOf(out))
}

func (p *ClientPort) Receive(msg interface{}, opts ...PortOpt) {
	options := defaultPortOpts
	for _, o := range opts {
		o(&options)
	}

	//TODO Use template pattern matching
	if err := deep.Equal(msg, p.resp); err != nil {
		log.Fatalf("Struct not eq: %v", err)
	}
}
