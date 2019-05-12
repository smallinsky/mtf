package port

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/pkg/errors"
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

func getGrpcDetails(s interface{}) (*serverDesc, error) {
	desc := serverDesc{}
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	name := t.Name()
	for _, suffix := range []string{"Server", "Client"} {
		name = strings.TrimSuffix(name, suffix)
	}

	sn, err := getPackageName(s)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get package name")
	}

	desc.Name = sn + "." + name

	for i := 0; i < t.NumMethod(); i++ {
		// TODO: distinguish stream methods
		m := t.Method(i)
		desc.methodsDesc = append(desc.methodsDesc, methodDesc{
			Name:    m.Name,
			InType:  m.Type.In(1),
			OutType: m.Type.Out(0),
		})
	}
	return &desc, nil
}

func getProtoDescFromBuff(buff []byte) (*descriptor.FileDescriptorProto, error) {
	r, err := gzip.NewReader(bytes.NewReader(buff))
	if err != nil {
		return nil, fmt.Errorf("failed to open gzip reader: %v", err)
	}
	defer r.Close()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read: %v", err)
	}

	fd := &descriptor.FileDescriptorProto{}
	if err := proto.Unmarshal(b, fd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to %T: %v", fd, err)
	}
	return fd, nil
}

func getPackageName(s interface{}) (string, error) {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		n1 := reflect.Zero(m.Type.In(1))
		mm := n1.MethodByName("Descriptor")
		if !mm.IsValid() {
			continue
		}
		r := mm.Call([]reflect.Value{})
		descBuff := r[0].Interface().([]byte)
		desc, err := getProtoDescFromBuff(descBuff)
		if err != nil {
			return "", errors.Wrapf(err, "failed to get proto descriptor")
		}
		if name := desc.GetPackage(); name != "" {
			return name, nil
		}
	}
	return "", errors.Errorf("package name not found")
}
