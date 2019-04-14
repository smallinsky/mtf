package main

import (
	"log"
	"net"

	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"

	pb "github.com/smallinsky/mtf/e2e/proto/echo"
	"github.com/smallinsky/mtf/e2e/proto/oracle"
)

type config struct {
	GrpcPort   string `envconfig:"GRPC_PORT" default:":8001"`
	OracleAddr string `envconfig:"ORACLE_ADDR" default:"localhost:8002"`

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
	pb.RegisterEchoServer(s, &server{})

	log.Println("Starting Echo Server")
	if err := s.Serve(l); err != nil {
		log.Fatalf("Error during grpc.Server: %v", err)
	}
}

func oracleClient(cfg config) oracle.OracleClient {
	conn, err := grpc.Dial(cfg.OracleAddr, grpc.WithInsecure())
	if err != nil {
		log.Printf("Faield to dial oracle service")
		return nil
	}

	return oracle.NewOracleClient(conn)
}
