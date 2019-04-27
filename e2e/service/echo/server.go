package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/go-redis/redis"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/smallinsky/mtf/e2e/proto/echo"
	pbo "github.com/smallinsky/mtf/e2e/proto/oracle"
)

type server struct {
	OracleClient pbo.OracleClient
	DB           *sql.DB
	RedisClient  *redis.Client
}

func (s *server) Repeat(ctx context.Context, req *pb.RepeatRequest) (*pb.RepeatResponse, error) {
	return &pb.RepeatResponse{
		Data: req.GetData(),
	}, nil
}

func (s *server) Scream(ctx context.Context, req *pb.ScreamRequest) (*pb.ScreamResponse, error) {
	log.Panicln("Scream enpoint called")
	return &pb.ScreamResponse{
		Data: fmt.Sprintf("%s !!!!", strings.ToUpper(req.GetData())),
	}, nil
}

func (s *server) AskGoogle(ctx context.Context, req *pb.AskGoogleRequest) (*pb.AskGoogleResponse, error) {
	log.Println("AskGoogle endpoint called")
	r, err := http.NewRequest(http.MethodGet, "http://api.icndb.com/jokes/random?firstName=John&amp;lastName=Doe", nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create get request %v", err)
	}
	resp, err := http.DefaultClient.Do(r.WithContext(ctx))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed t call http %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, status.Errorf(codes.Internal, "failed to call api: status code %v", resp.StatusCode)
	}

	defer resp.Body.Close()
	buff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read reasponse body: %v", err)
	}

	type incdbResp struct {
		Value struct {
			Joke string `json:"joke,omitempty"`
		} `json:"value,omitempty"`
	}

	var ir incdbResp

	if err := json.Unmarshal(buff, &ir); err != nil {
		return nil, status.Errorf(codes.Internal, "failed unmarshal body: %v", err)
	}

	return &pb.AskGoogleResponse{
		Data: ir.Value.Joke,
	}, nil
}

func (s *server) AskDB(ctx context.Context, req *pb.AskDBRequest) (*pb.AskDBResponse, error) {
	var punchline string
	err := s.DB.QueryRowContext(ctx, "SELECT punchline FROM happy_table WHERE query=?", req.GetData()).Scan(&punchline)
	if err == sql.ErrNoRows {
		return nil, status.New(codes.NotFound, "no punchline").Err()
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch from db, err: %v", err)
	}

	return &pb.AskDBResponse{
		Data: punchline,
	}, nil
}

func (s *server) AskRedis(ctx context.Context, req *pb.AskRedisRequest) (*pb.AskRedisResponse, error) {
	resp, err := s.RedisClient.Get(req.GetData()).Result()
	if err == redis.Nil {
		return nil, status.Errorf(codes.NotFound, "key '%v' not found", req.GetData())
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get key value from redis, err %v", err)
	}

	return &pb.AskRedisResponse{
		Data: resp,
	}, nil
}

func (s *server) AskOracle(ctx context.Context, req *pb.AskOracleRequest) (*pb.AskOracleResponse, error) {
	log.Println("AskOrace ongoing....")
	resp, err := s.OracleClient.AskDeepThrough(context.Background(), &pbo.AskDeepThroughRequest{
		Data: req.GetData(),
	})
	if err != nil {
		return nil, err
	}

	return &pb.AskOracleResponse{
		Data: resp.GetData(),
	}, nil
}
