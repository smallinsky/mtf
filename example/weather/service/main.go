package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/kelseyhightower/envconfig"

	pb "github.com/smallinsky/mtf/proto/weather"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type config struct {
	GrpcPort    string `envconfig:"GRPC_PORT" default:":8001"`
	TLSCertPath string `envconfig:"TLS_CERT_PATH"`
	TLSKeyPath  string `envconfig:"TLS_KEY_PATH"`

	ScaleConvAddr string `envconfig:"SCALE_CONV_ADDR"`
}

const (
	weatherEndpointAPI = "https://api.weather.com/"
)

type server struct {
	scaleConvClient pb.ScaleConvClient
}

func (s *server) AskAboutWeather(ctx context.Context, req *pb.AskAboutWeatherRequest) (*pb.AskAboutWeatherResponse, error) {

	cresp, err := s.scaleConvClient.CelsiusToFahrenheit(ctx, &pb.CelsiusToFahrenheitRequest{Value: 133})
	if err != nil {
		if status.Code(err) == codes.FailedPrecondition {
			return nil, status.Errorf(codes.Internal, "scale conv client failed: %v", err)
		}
		return nil, err
	}
	return &pb.AskAboutWeatherResponse{
		Result: fmt.Sprintf("%v", cresp.GetValue()),
	}, nil

}

func scalConvClient(cfg config) pb.ScaleConvClient {
	conn, err := grpc.Dial(cfg.ScaleConvAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Faield to dial oracle service: %v", err)
	}
	return pb.NewScaleConvClient(conn)
}

func main() {
	var cfg config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("Failed to parse env config: %v", err)
	}

	l, err := net.Listen("tcp", cfg.GrpcPort)
	if err != nil {
		log.Fatalf("Failed to start net listener: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterWeatherServer(s, &server{
		scaleConvClient: scalConvClient(cfg),
	})

	if err := s.Serve(l); err != nil {
		log.Fatalf("Failed to server rpc server: %v", err)
	}
}
