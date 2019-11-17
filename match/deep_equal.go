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
	if reflect.DeepEqual(got, m.exp) {
		return nil
	}

	gotDump, err := toJsonString(got)
	if err != nil {
		return err
	}
	expDump, err := toJsonString(m.exp)
	if err != nil {
		return err
	}

	return errors.Wrapf(ErrNotEq, "deep equal: \n got:'%v'\n exp: '%v'\n", gotDump, expDump)
}

func toJsonString(i interface{}) (string, error) {
	buff, err := json.MarshalIndent(i, "", " ")
	if err != nil {
		return "", nil
	}
	return string(buff), nil
}
