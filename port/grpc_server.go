package port

import (
	"context"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

func (p *PortIn) Send(ctx context.Context, i interface{}) error {
	return p.send(i)
}

func (p *PortIn) Receive(ctx context.Context) (interface{}, error) {
	return p.receive()
}

func NewGRPCServersPort(ii []interface{}, port string, opts ...PortOpt) (*Port, error) {
	p, err := NewGRPCServers(ii, port, opts...)
	if err != nil {
		return nil, err
	}
	return &Port{
		impl: p,
	}, nil
}

func NewGRPCServerPort(i interface{}, port string, opts ...PortOpt) (*Port, error) {
	p, err := NewGRPCServer(i, port, opts...)
	if err != nil {
		return nil, err
	}
	return &Port{
		impl: p,
	}, nil
}

func NewGRPCServers(ii []interface{}, port string, opts ...PortOpt) (*PortIn, error) {
	options := defaultPortOpts
	for _, o := range opts {
		o(&options)
	}

	portIn := &PortIn{
		reqC:  make(chan interface{}),
		respC: make(chan outValues),
	}

	lis, err := listen("tcp", port)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create net listener")
	}

	grpcOpts := []grpc.ServerOption{}
	if options.serverCertPath != "" && options.serverKeyPath != "" {
		creds, err := credentials.NewServerTLSFromFile(options.serverCertPath, options.serverKeyPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load TLS certs")
		}
		grpcOpts = append(grpcOpts, grpc.Creds(creds))
	}

	s, err := registerInterfaces(grpc.NewServer(grpcOpts...), ii, portIn.rpcCallHandler, options)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to reqigster server interface")
	}

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to server %v", err)
		}
	}()

	return portIn, nil
}

func NewGRPCServer(i interface{}, port string, opts ...PortOpt) (*PortIn, error) {
	options := defaultPortOpts
	for _, o := range opts {
		o(&options)
	}

	portIn := &PortIn{
		reqC:  make(chan interface{}),
		respC: make(chan outValues),
	}

	lis, err := listen("tcp", port)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create net listener")
	}

	grpcOpts := []grpc.ServerOption{}
	if options.serverCertPath != "" && options.serverKeyPath != "" {
		creds, err := credentials.NewServerTLSFromFile(options.serverCertPath, options.serverKeyPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load TLS certs")
		}
		grpcOpts = append(grpcOpts, grpc.Creds(creds))
	}

	s, err := registerInterface(grpc.NewServer(grpcOpts...), i, portIn.rpcCallHandler, options)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to reqigster server interface")
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Done()
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to server %v", err)
		}
	}()
	wg.Wait()
	return portIn, nil
}

func (p *PortIn) rpcCallHandler(req interface{}) (interface{}, error) {
	go func() {
		p.reqC <- req
	}()
	resp := <-p.respC
	return resp.msg, resp.err
}

func (p *PortIn) receive(opts ...Opt) (interface{}, error) {
	options := defaultPortOpts
	for _, o := range opts {
		o(&options)
	}

	select {
	case <-time.NewTimer(options.timeout).C:
		return nil, errors.Errorf("failed to receive  message, deadline exeeded")
	case v := <-p.reqC:
		return v, nil
	}
}

func (p *PortIn) send(msg interface{}, opts ...PortOpt) error {
	options := defaultPortOpts
	for _, o := range opts {
		o(&options)
	}

	p.respC <- outValues{
		msg: msg,
		err: options.err,
	}
	return nil
}

func registerInterface(server *grpc.Server, i interface{}, procCall processFunc, opts portOpts) (*grpc.Server, error) {
	sv := reflect.ValueOf(&server).Elem().Elem().FieldByName("m")
	nsv := reflect.New(sv.Type().Elem().Elem())
	mdv := nsv.Elem().FieldByName("md")
	//TODO register handler for stream methods

	mv := allocMap(mdv)

	desc, err := getGrpcDetails(i)
	if err != nil {
		return nil, errors.Wrapf(err, "failed ot get grpc details")
	}
	for _, mdesc := range desc.MethodsDesc {
		mdesc := mdesc
		fn := func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
			v := reflect.New(mdesc.InType)
			v.Elem().Set(reflect.New(mdesc.InType.Elem()))
			dec(v.Elem().Interface())
			return procCall(v.Elem().Interface())
		}

		z := reflect.New(mdv.Type().Elem().Elem())
		z.Elem().FieldByName("MethodName").SetString(mdesc.Name)
		z.Elem().FieldByName("Handler").Set(reflect.ValueOf(fn))

		serverName := ""
		if opts.pkgName != "" {
			serverName = strings.Join([]string{opts.pkgName, mdesc.Name}, ".")
		} else {
			serverName = mdesc.Name
		}
		mv.Elem().SetMapIndex(reflect.ValueOf(serverName), z)
	}

	mv = allocMap(sv)
	mv.Elem().SetMapIndex(reflect.ValueOf(desc.Name), nsv)
	return server, nil
}

func registerInterfaces(server *grpc.Server, ii []interface{}, procCall processFunc, opts portOpts) (*grpc.Server, error) {
	sv := reflect.ValueOf(&server).Elem().Elem().FieldByName("m")

	svm := allocMap(sv)

	for _, i := range ii {
		nsv := reflect.New(sv.Type().Elem().Elem())
		mdv := nsv.Elem().FieldByName("md")
		mv := allocMap(mdv)
		//TODO register handler for stream methods

		desc, err := getGrpcDetails(i)
		if err != nil {
			return nil, errors.Wrapf(err, "failed ot get grpc details")
		}
		for _, mdesc := range desc.MethodsDesc {
			mdesc := mdesc
			fn := func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
				v := reflect.New(mdesc.InType)
				v.Elem().Set(reflect.New(mdesc.InType.Elem()))
				dec(v.Elem().Interface())

				return procCall(v.Elem().Interface())
			}

			z := reflect.New(mdv.Type().Elem().Elem())
			z.Elem().FieldByName("MethodName").SetString(mdesc.Name)
			z.Elem().FieldByName("Handler").Set(reflect.ValueOf(fn))

			serverName := ""
			if opts.pkgName != "" {
				serverName = strings.Join([]string{opts.pkgName, mdesc.Name}, ".")
			} else {
				serverName = mdesc.Name
			}
			mv.Elem().SetMapIndex(reflect.ValueOf(serverName), z)
		}

		svm.Elem().SetMapIndex(reflect.ValueOf(desc.Name), nsv)
	}

	return server, nil
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
