package grpc

import (
	"fmt"
	"reflect"
	"strings"
)

type serverDesc struct {
	Name        string
	methodsDesc []methodDesc
}

type methodDesc struct {
	InType  reflect.Type
	OutType reflect.Type
	Name    string
}

func getGrpcDetails(s interface{}) serverDesc {
	desc := serverDesc{}
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	ps := strings.Split(t.PkgPath(), "/")
	name := t.Name()
	for _, suffix := range []string{"Server", "Client"} {
		name = strings.TrimSuffix(name, suffix)
	}
	// TODO Get full pacakge name with protobuf prefix like example.Server.NameOfServer
	if false {
		desc.Name = ps[len(ps)-1] + "." + name
	}

	desc.Name = name

	for i := 0; i < t.NumMethod(); i++ {
		// TODO: distinguish stream methods
		m := t.Method(i)
		desc.methodsDesc = append(desc.methodsDesc, methodDesc{
			Name:    m.Name,
			InType:  m.Type.In(1),
			OutType: m.Type.Out(0),
		})
	}
	fmt.Println(desc)
	return desc
}
