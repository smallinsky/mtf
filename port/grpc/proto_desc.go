package grpc

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

func getProtoDescFromFile(file string) (*descriptor.FileDescriptorProto, error) {
	b := proto.FileDescriptor(file)
	if b != nil {
		return nil, fmt.Errorf("file %s not registered in proto package", file)
	}
	return getProtoDescFromBuff(b)
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
