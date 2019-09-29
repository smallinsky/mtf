package match

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

type ProtoEqualType struct {
	exp interface{}
}

func ProtoEqual(exp interface{}) *ProtoEqualType {
	return &ProtoEqualType{
		exp: exp,
	}
}

func (m *ProtoEqualType) Match(got interface{}) error {
	gotP, ok := got.(proto.Message)
	if !ok {
		return fmt.Errorf("%T is not proto message", got)
	}

	expP, ok := m.exp.(proto.Message)
	if !ok {
		return fmt.Errorf("%T is not proto message", got)
	}
	if !proto.Equal(gotP, expP) {
		return errors.Wrapf(ErrNotEq, "deep equal: \n got: '%v'\n exp: '%v'\n", gotP.String(), expP.String())
	}
	return nil
}
