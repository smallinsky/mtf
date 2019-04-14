package main

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smallinsky/mtf/e2e/proto/oracle"
)

type server struct {
}

func (s *server) AskDeepThrough(context.Context, *oracle.AskDeepThroughRequest) (*oracle.AskDeepThroughRespnse, error) {
	return nil, status.New(codes.Unimplemented, "unimplemented").Err()
}
