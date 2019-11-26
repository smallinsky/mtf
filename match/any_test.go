package match

import (
	"testing"

	pb "github.com/golang/protobuf/proto/proto3_proto"
	"github.com/pkg/errors"
)

func TestType(t *testing.T) {
	cases := []struct {
		name string
		exp  interface{}
		got  interface{}
		err  error
	}{
		{
			name: "int same types",
			exp:  int(1),
			got:  int(2),
		},
		{
			name: "int and int32 different types",
			exp:  int(1),
			got:  int32(2),
			err:  ErrNotEq,
		},
		{
			name: "proto ptr vs obj different types",
			exp: &pb.Message{
				Name: "ptr message",
			},
			got: pb.Message{
				Name: "obj message",
			},
			err: ErrNotEq,
		},
		{
			name: "proto samge types",
			exp: &pb.Message{
				Name: "ptr message 1",
			},
			got: pb.Message{
				Name: "ptr message 2",
			},
			err: ErrNotEq,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := Type(tc.exp)
			err := a.Match(tc.got)
			if errors.Cause(err) != errors.Cause(tc.err) {
				t.Fatalf("Unexpecte error: %v", err)
			}

		})
	}
}
