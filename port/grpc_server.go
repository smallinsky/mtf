package port

import (
	"context"
	"log"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"github.com/go-test/deep"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/smallinsky/mtf/port/match"
)

type processFunc func(i interface{}) (interface{}, error)

type outValues struct {
	msg interface{}
	err error
}

type PortIn struct {
	reqC  chan interface{}
	respC chan outValues
}

func NewGRPCServer(i interface{}, port string, opts ...PortOpt) *PortIn {
	options := defaultPortOpts
	for _, o := range opts {
		o(&options)
	}

	p := &PortIn{
		reqC:  make(chan interface{}),
		respC: make(chan outValues),
	}

	fn := func(i interface{}) (interface{}, error) {
		go func() {
			p.reqC <- i
		}()
		retV := <-p.respC
		return retV.msg, retV.err
	}

	// TODO Add tls support
	lis, err := listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}

	grpcOpts := []grpc.ServerOption{}
	if options.serverCertPath != "" && options.serverKeyPath != "" {
		creds, err := credentials.NewServerTLSFromFile(options.serverCertPath, options.serverKeyPath)
		if err != nil {
			log.Fatalf("failed to load TLS certs")
		}
		grpcOpts = append(grpcOpts, grpc.Creds(creds))
	}

	s := registerInterface(grpc.NewServer(grpcOpts...), i, fn, options)

	startSync.Add(1)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to server %v", err)
		}
	}()
	return p
}

func (p *PortIn) Receive(i interface{}, opts ...Opt) {
	options := defaultPortOpts
	for _, o := range opts {
		o(&options)
	}

	// TODO handle messages by type and add erro on unexpected msg recived
	select {
	case v := <-p.reqC:
		//TODO Use template pattern matching
		if err := deep.Equal(v, i); err != nil {
			log.Fatalf("Struct not eq: %v", err)
		}
	case <-time.NewTimer(options.timeout).C:
		log.Printf("Timeout, expected message %T not received\n", i)
	}
}

func (p *PortIn) ReceiveMatch(i ...interface{}) {
	r, err := match.PayloadMatchFucs(i...)
	if err != nil {
		panic(err)
	}

	select {
	case v := <-p.reqC:
		r.MatchFn(nil, v)
	case <-time.NewTimer(time.Second * 5).C:
		log.Printf("Timeout, expected message %T not received\n", r.ArgType)
	}
}

func (p *PortIn) Send(msg interface{}, opts ...PortOpt) {
	options := defaultPortOpts
	for _, o := range opts {
		o(&options)
	}

	p.respC <- outValues{
		msg: msg,
		err: options.err,
	}
}

func registerInterface(s *grpc.Server, i interface{}, procCall processFunc, opts portOpts) *grpc.Server {
	sv := reflect.ValueOf(&s).Elem().Elem().FieldByName("m")
	nsv := reflect.New(sv.Type().Elem().Elem())
	mdv := nsv.Elem().FieldByName("md")
	//TODO register handler for stream methods

	z := reflect.New(mdv.Type().Elem().Elem())
	mv := allocMap(mdv)

	desc := getGrpcDetails(i)
	for _, mdesc := range desc.methodsDesc {
		fn := func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
			v := reflect.New(mdesc.InType)
			v.Elem().Set(reflect.New(mdesc.InType.Elem()))
			dec(v.Elem().Interface())
			return procCall(v.Elem().Interface())
		}
		z.Elem().FieldByName("MethodName").SetString(mdesc.Name)
		z.Elem().FieldByName("Handler").Set(reflect.ValueOf(fn))

		serverName := ""
		// TODO: pkgName is probably not needed during server registartion, check it
		if opts.pkgName != "" {
			serverName = strings.Join([]string{opts.pkgName, mdesc.Name}, ".")
		} else {
			serverName = mdesc.Name
		}
		mv.Elem().SetMapIndex(reflect.ValueOf(serverName), z)
	}

	mv = allocMap(sv)
	mv.Elem().SetMapIndex(reflect.ValueOf(desc.Name), nsv)
	return s
}

func allocMap(v reflect.Value) reflect.Value {
	vm := reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr()))
	vm.Elem().Set(reflect.MakeMap(vm.Elem().Type()))
	return vm
}

func getServerDesc(s interface{}) (name string, methods []string) {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	ps := strings.Split(t.PkgPath(), "/")
	name = ps[len(ps)-1] + strings.TrimSuffix(t.Name(), "Server")
	for i := 0; i < t.NumMethod(); i++ {
		// TODO: distinguish stream methods
		methods = append(methods, t.Method(i).Name)
	}
	return
}
