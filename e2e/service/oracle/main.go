package main

import (
	"log"
	"net"

	_ "github.com/go-sql-driver/mysql"
	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"

	pb "github.com/smallinsky/mtf/e2e/proto/oracle"
)

type config struct {
	GrpcPort string `envconfig:"GRPC_PORT" default:":8001"`

	TLSRootPath string `envconfig:TLS_ROOT_PATH`
	TLSCertPath string `envconfig:TLS_CERT_PATH`
	TLSKeyPath  string `envconfig:TLS_KEY_PATH`
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
	pb.RegisterOracleServer(s, &server{})

	log.Println("Starting echo server")
	if err := s.Serve(l); err != nil {
		log.Fatalf("error during grpc.server: %v", err)
	}
}
