package main

import (
	"database/sql"
	"log"
	"net"

	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"

	_ "github.com/go-sql-driver/mysql"
	pb "github.com/smallinsky/mtf/e2e/proto/echo"
	"github.com/smallinsky/mtf/e2e/proto/oracle"
	"github.com/smallinsky/mtf/pkg"
)

type config struct {
	GrpcPort   string `envconfig:"GRPC_PORT" default:":8001"`
	OracleAddr string `envconfig:"ORACLE_ADDR" default:"localhost:8002"`

	TLSRootPath string `envconfig:TLS_ROOT_PATH`
	TLSCertPath string `envconfig:TLS_CERT_PATH`
	TLSKeyPath  string `envconfig:TLS_KEY_PATH`
	DBDsn       string `envconfig:"DB_DSN" default:"root:test@tcp(localhost:3306)/test_db"`
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

	oracleCli := oracleClient(cfg)
	s := grpc.NewServer()
	pb.RegisterEchoServer(s, &server{
		Client: oracleCli,
		DB:     initDB(cfg.DBDsn),
	})

	if err := s.Serve(l); err != nil {
		log.Fatalf("Error during grpc.Server: %v", err)
	}
}

func oracleClient(cfg config) oracle.OracleClient {
	conn, err := grpc.Dial(cfg.OracleAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Faield to dial oracle service")
		return nil
	}
	pkg.StartMonitor(conn)
	return oracle.NewOracleClient(conn)
}

func initDB(dsn string) *sql.DB {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to connect to %v, err: %v\n", dsn, err)
	}
	return db
}
