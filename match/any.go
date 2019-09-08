package match

import (
	"reflect"

	"github.com/pkg/errors"
)

type TypeT struct {
	expectedMessageType interface{}
}

func Type(expectedMessageType interface{}) *TypeT {
	return &TypeT{
		expectedMessageType: expectedMessageType,
	}
}

func (m *TypeT) Match(receiveMessageType interface{}) error {
	if exp, got := reflect.TypeOf(m.expectedMessageType).Name(), reflect.TypeOf(receiveMessageType).Name(); got != exp {
		return errors.Wrapf(ErrNotEq, "types: %v != %v", got, exp)
	}
	return nil
}
