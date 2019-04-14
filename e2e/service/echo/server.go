package main

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/smallinsky/mtf/e2e/proto/echo"
)

type server struct {
}

func (s *server) Repeat(ctx context.Context, req *pb.RepeatRequest) (*pb.RepeatResponse, error) {
	return &pb.RepeatResponse{
		Data: req.GetData(),
	}, nil
}

func (s *server) Scream(ctx context.Context, req *pb.ScreamRequest) (*pb.ScreamResponse, error) {
	return &pb.ScreamResponse{
		Data: fmt.Sprintf("%s !!!!", strings.ToUpper(req.GetData())),
	}, nil
}

func (s *server) AskGoogle(ctx context.Context, req *pb.AskGoogleRequest) (*pb.AskGoogleResponse, error) {
	return nil, status.New(codes.Unimplemented, "unimplemented").Err()
}

func (s *server) AskDB(ctx context.Context, req *pb.AskDBRequest) (*pb.AskDBResponse, error) {
	return nil, status.New(codes.Unimplemented, "unimplemented").Err()
}

func (s *server) AskRedis(ctx context.Context, req *pb.AskRedisRequest) (*pb.AskRedisResponse, error) {
	return nil, status.New(codes.Unimplemented, "unimplemented").Err()
}

func (s *server) AskOracle(ctx context.Context, req *pb.AskOracleRequest) (*pb.AskOracleResponse, error) {
	return nil, status.New(codes.Unimplemented, "unimplemented").Err()
}
