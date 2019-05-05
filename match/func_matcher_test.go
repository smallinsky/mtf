package match

import (
	"testing"

	"github.com/pkg/errors"
)

type firstType struct {
	Data string
}

type secondType struct {
	Data string
}

func TestFuncValidate(t *testing.T) {
	cases := []struct {
		name string
		args []interface{}
		err  error
	}{
		{
			name: "NilArgument",
			args: nil,
			err:  ErrMatchFnInvalidArg,
		},
		{
			name: "PaylaodMatchFunc",
			args: []interface{}{
				func(p *firstType) {
				},
			},
			err: nil,
		},
		{
			name: "PaylaodMatchFunc",
			args: []interface{}{
				func(p firstType) {
				},
			},
			err: ErrMatchFnInvalidArg,
		},
		{
			name: "PaylaodAndErrMatch",
			args: []interface{}{
				func(p *firstType) {
				},
				func(err error) {
				},
			},
			err: ErrMatchFnInvalidArg,
		},
		{
			name: "DiffrentMessagesType",
			args: []interface{}{
				func(p *firstType) {
				},
				func(p *secondType) {
				},
			},
			err: ErrMatchFnInvalidArg,
		},
		{
			name: "DiffrentMessagesTypeNoPointers",
			args: []interface{}{
				func(p firstType) {
				},
				func(p secondType) {
				},
			},
			err: ErrMatchFnInvalidArg,
		},
		{
			name: "WrongMatchFnPrototype",
			args: []interface{}{
				func(s string) {
				},
			},
			err: ErrMatchFnInvalidArg,
		},
		{
			name: "Not match func or struct",
			args: []interface{}{
				string("first args"),
			},
			err: ErrMatchFnInvalidArg,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := Fn(tc.args...)
			err := m.Validate()
			if err != nil && tc.err == nil {
				t.Fatalf("got unexpected error '%v'", err)
			}
			if errors.Cause(err) != tc.err {
				t.Fatalf("got '%v', exp '%v'", err, tc.err)
			}
			t.Logf("[DEBUG] err: '%v'", err)
		})
	}
}

func TestFuncMatch(t *testing.T) {
	data := "example data"
	s := &firstType{
		Data: data,
	}
	m := Fn(
		func(p *firstType) {
			if got, exp := p.Data, data; got != exp {
				t.Fatalf("got '%v', exp '%v'", got, exp)
			}
		},
		func(p *firstType) {
			if len(p.Data) <= 8 {
				t.Fatalf("data lenght should be greater than 8")
			}
		},
	)
	if err := m.Validate(); err != nil {
		t.Fatalf("unexpected error during validate call: %v", err)
	}
	m.Match(nil, s)
}
