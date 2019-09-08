package match

import (
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"
)

type DeepEqualType struct {
	exp interface{}
}

func DeepEqual(exp interface{}) *DeepEqualType {
	return &DeepEqualType{
		exp: exp,
	}
}

func (m *DeepEqualType) Match(got interface{}) error {
	if !reflect.DeepEqual(got, m.exp) {
		return errors.Wrapf(ErrNotEq, "deep equal: \n got:'%v'\n exp: '%v'\n", toJsonString(got), toJsonString(m.exp))
	}
	return nil
}

func toJsonString(i interface{}) string {
	buff, err := json.MarshalIndent(i, "", " ")
	if err != nil {
		panic(err)
	}
	return string(buff)
}
