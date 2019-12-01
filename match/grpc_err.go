package match

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCErrType struct {
	Code    codes.Code
	Message string
}

func GRPCErr(code codes.Code, msg string) *GRPCErrType {
	return &GRPCErrType{
		Code:    code,
		Message: msg,
	}
}

func GRPCStatusCode(code codes.Code) *GRPCErrType {
	return &GRPCErrType{
		Code: code,
	}
}

func (m *GRPCErrType) Match(err error) error {
	s, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("invalid error type")
	}
	if got, want := s.Code(), m.Code; got != want {
		return fmt.Errorf("got unexpected error status code:\nGot:  '%v'\nWant: '%v'", got, want)
	}
	if m.Message == "" {
		return nil
	}
	if got, want := s.Message(), m.Message; !strings.Contains(got, want) {
		return fmt.Errorf("unexpected grpc error message:\n Got: '%s'\nWant: '%s'", got, want)
	}
	return nil
}
