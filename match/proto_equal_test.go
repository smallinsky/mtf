package match

import (
	"testing"

	pb "github.com/golang/protobuf/proto/proto3_proto"
	"github.com/pkg/errors"
)

func TestProtoEqual(t *testing.T) {
	cases := []struct {
		name string
		exp  interface{}
		got  interface{}
		err  error
	}{
		{
			name: "proto samge types",
			exp: &pb.Message{
				Name: "ptr message 1",
			},
			got: &pb.Message{
				Name: "ptr message 2",
			},
			err: ErrNotEq,
		},
		{
			name: "proto samge type and value",
			exp: &pb.Message{
				Name: "42",
			},
			got: &pb.Message{
				Name: "42",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := ProtoEqual(tc.exp)
			err := a.Match(tc.got)
			if errors.Cause(err) != errors.Cause(tc.err) {
				t.Fatalf("Unexpecte error: %v", err)
			}
		})
	}
}
