package port

import (
	"context"
	"log"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/go-test/deep"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/smallinsky/mtf/match"
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

func NewGRPCServer(i interface{}, port string, opts ...PortOpt) (*PortIn, error) {
	options := defaultPortOpts
	for _, o := range opts {
		o(&options)
	}

	portIn := &PortIn{
		reqC:  make(chan interface{}),
		respC: make(chan outValues),
	}

	fn := func(i interface{}) (interface{}, error) {
		go func() {
			portIn.reqC <- i
		}()
		retV := <-portIn.respC
		return retV.msg, retV.err
	}

	// TODO Add tls support
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

	s, err := registerInterface(grpc.NewServer(grpcOpts...), i, fn, options)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to reqigster server interface")
	}

	startSync.Add(1)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to server %v", err)
		}
	}()
	return portIn, nil
}

func (p *PortIn) Receive(i interface{}, opts ...Opt) error {
	options := defaultPortOpts
	for _, o := range opts {
		o(&options)
	}

	select {
	case v := <-p.reqC:
		if diff := deep.Equal(v, i); diff != nil {
			return errors.Errorf("Struct not eq: \n diff: '%v'", diff)
		}
	case <-time.NewTimer(options.timeout).C:
		return errors.Errorf("failed to receive  message, deadline exeeded")
	}
	return nil
}

func (p *PortIn) ReceiveM(m match.Matcher, opts ...Opt) error {
	options := defaultPortOpts
	for _, o := range opts {
		o(&options)
	}
	if err := m.Validate(); err != nil {
		return errors.Wrapf(err, "invalid marcher argument")
	}

	select {
	case got := <-p.reqC:
		if err := m.Match(nil, got); err != nil {
			return errors.Wrapf(err, "%T message match failed", m)
		}
	case <-time.NewTimer(options.timeout).C:
		return errors.Errorf("failed to receive  message, deadline exeeded")
	}
	return nil
}

func (p *PortIn) Send(msg interface{}, opts ...PortOpt) error {
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

	z := reflect.New(mdv.Type().Elem().Elem())
	mv := allocMap(mdv)

	desc, err := getGrpcDetails(i)
	if err != nil {
		return nil, errors.Wrapf(err, "failed ot get grpc details")
	}
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
