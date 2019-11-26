package match

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

var (
	ErrMatchFnInvalidArg = errors.New("match fn invalid arg")
)

type FnType struct {
	Args []interface{}
}

func Fn(args ...interface{}) *FnType {
	return &FnType{
		Args: args,
	}
}

func (m *FnType) Match(err error, got interface{}) error {
	var matchFuncs []func(interface{})
	vmfs := reflect.ValueOf(&matchFuncs)

	var ht reflect.Type
	for _, arg := range m.Args {
		v := reflect.ValueOf(arg)
		if ht == nil {
			ht = v.Type().In(0)
		}
		fn := func(i interface{}) {
			v.Call([]reflect.Value{reflect.ValueOf(i)})
		}
		vmfs.Elem().Set(reflect.Append(vmfs.Elem(), reflect.ValueOf(fn)))
	}

	if gt := reflect.TypeOf(got); gt != ht {
		return fmt.Errorf("received '%v' message, but handler functions are defined for '%v'", gt, ht)
	}
	for _, fn := range matchFuncs {
		fn(got)
	}
	return nil
}

func (m *FnType) Validate() error {
	if m.Args == nil {
		return errors.Wrap(ErrMatchFnInvalidArg, "got nil argument")
	}
	var t reflect.Type
	if len(m.Args) == 0 {
		return errors.Wrapf(ErrMatchFnInvalidArg, "0 arguments")
	}
	for _, arg := range m.Args {
		v := reflect.ValueOf(arg)
		if v.Type().Kind() != reflect.Func {
			return errors.Wrapf(ErrMatchFnInvalidArg, "expected function but got '%v'", v.Type())
		}
		if v.Type().NumIn() != 1 {
			return errors.Wrapf(ErrMatchFnInvalidArg, "match function prototype expected one match arg but got '%v'", v.Type())
		}
		it := v.Type().In(0)
		if it.Kind() != reflect.Ptr {
			return errors.Wrapf(ErrMatchFnInvalidArg, "match function args should be a pointer to struct, but got %v in '%v'", it.Kind(), it)
		}
		it = it.Elem()
		if it.Kind() != reflect.Struct {
			return errors.Wrapf(ErrMatchFnInvalidArg, "match function args should be a pointer to struct, but got ptr to %v in '%v'", it.Kind(), it)
		}

		if v.Type().NumOut() != 0 {
			return errors.Wrapf(ErrMatchFnInvalidArg, "match function prototype should return void but got '%v'", v.Type())
		}
		if t == nil {
			t = v.Type()
		} else if v.Type() != t {
			return errors.Wrap(ErrMatchFnInvalidArg, "match functions suppose to have same prototype")
		}
	}
	return nil
}
