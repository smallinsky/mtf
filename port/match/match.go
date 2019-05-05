package match

import (
	"reflect"

	"github.com/pkg/errors"
)

var (
	ErrMatchFnInvalidArg = errors.New("match fn invalid arg")
)

type PayloadMatcher struct {
	Msg interface{}
}

func Payload(message interface{}) PayloadMatcher {
	return PayloadMatcher{
		Msg: message,
	}
}

type FnMatcher struct {
	Calls []interface{}
}

func Fn(fn ...interface{}) FnMatcher {
	return FnMatcher{
		Calls: fn,
	}
}

type MatchResult struct {
	MatchFn MatchFn
	ArgType reflect.Type
}

type MatchFn func(err error, i interface{})

func PayloadMatchFucs(args ...interface{}) (*MatchResult, error) {
	if err := validate(args); err != nil {
		return nil, err
	}

	var matchFuncs []func(interface{})
	vmfs := reflect.ValueOf(&matchFuncs)

	res := &MatchResult{}
	for _, arg := range args {
		v := reflect.ValueOf(arg)
		if res.ArgType == nil {
			res.ArgType = v.Type().In(0)
		}
		fn := func(i interface{}) {
			v.Call([]reflect.Value{reflect.ValueOf(i)})
		}
		vmfs.Elem().Set(reflect.Append(vmfs.Elem(), reflect.ValueOf(fn)))
	}
	res.MatchFn = func(err error, i interface{}) {
		for _, fn := range matchFuncs {
			fn(i)
		}
	}
	return res, nil
}

func validate(args []interface{}) error {
	if args == nil {
		return errors.Wrap(ErrMatchFnInvalidArg, "got nil argument")
	}
	var t reflect.Type
	if len(args) == 0 {
		return errors.Wrapf(ErrMatchFnInvalidArg, "0 arguments")
	}
	for _, arg := range args {
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
			return errors.Wrap(ErrMatchFnInvalidArg, "match functions supose to have same prototype")
		}
	}
	return nil
}
