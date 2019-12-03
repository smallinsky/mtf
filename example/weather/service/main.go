package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/kelseyhightower/envconfig"

	pb "github.com/smallinsky/mtf/proto/weather"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
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
	r, err := http.NewRequest(http.MethodGet, weatherEndpointAPI, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to build http request: %v", err)
	}

	resp, err := http.DefaultClient.Do(r.WithContext(ctx))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reach weather api: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, status.Errorf(codes.Internal, "failed to reach weather api: invalid http status %s", resp.Status)
	}

	defer resp.Body.Close()
	buff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read response body content: %v", err)
	}

	if req.GetScale() == pb.Scale_FAHRENHEIT {
		val, err := strconv.ParseInt(string(buff), 10, 32)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to convert http response to int: %v", err)
		}
		cresp, err := s.scaleConvClient.CelciusToFarenheit(ctx, &pb.CelciusToFarenheitRequest{Value: val})
		if err != nil {
			if status.Code(err) == codes.FailedPrecondition {
				return nil, status.Errorf(codes.Internal, "scale conv client failed: %v", err)
			}
			return nil, err
		}
		return &pb.AskAboutWeatherResponse{
			Result: fmt.Sprintf("%d Farenheit Degrees", cresp.GetValue()),
		}, nil
	}

	return &pb.AskAboutWeatherResponse{
		Result: string(buff),
	}, nil
}

func scalConvClient(cfg config) pb.ScaleConvClient {
	creds, err := credentials.NewClientTLSFromFile(cfg.TLSCertPath, "")
	conn, err := grpc.Dial(cfg.ScaleConvAddr, grpc.WithTransportCredentials(creds))
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

	creds, err := credentials.NewServerTLSFromFile(cfg.TLSCertPath, cfg.TLSKeyPath)
	if err != nil {
		log.Fatalf("Failed to get creds: %v", err)
	}

	l, err := net.Listen("tcp", cfg.GrpcPort)
	if err != nil {
		log.Fatalf("Failed to start net listener: %v", err)
	}

	s := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterWeatherServer(s, &server{
		scaleConvClient: scalConvClient(cfg),
	})

	if err := s.Serve(l); err != nil {
		log.Fatalf("Failed to server rpc server: %v", err)
	}
}
