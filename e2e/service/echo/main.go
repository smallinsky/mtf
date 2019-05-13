package main

import (
	"database/sql"
	"log"
	"net"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"

	pb "github.com/smallinsky/mtf/e2e/proto/echo"
	"github.com/smallinsky/mtf/e2e/proto/oracle"
)

type config struct {
	GrpcPort   string `envconfig:"GRPC_PORT" default:":8001"`
	OracleAddr string `envconfig:"ORACLE_ADDR" default:"oracle_mtf:8002"`

	TLSRootPath string `envconfig:"TLS_ROOT_PATH"`
	TLSCertPath string `envconfig:"TLS_CERT_PATH"`
	TLSKeyPath  string `envconfig:"TLS_KEY_PATH"`
	DBDsn       string `envconfig:"DB_DSN" default:"root:test@tcp(mysql_mtf:3306)/test_db"`

	RedisAdrr string `envconfig:"REDIS_ADDR" default:"redis_mtf:6379"`
	RedisPass string `envconfig:"REDIS_PASS" default:"test"`
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
		OracleClient: oracleCli,
		DB:           initDB(cfg.DBDsn),
		RedisClient:  initRedis(cfg),
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
	return oracle.NewOracleClient(conn)
}

func initDB(dsn string) *sql.DB {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to connect to %v, err: %v\n", dsn, err)
	}
	return db
}

func initRedis(cfg config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAdrr,
		Password: cfg.RedisPass,
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Fatalf("failed to coonect to reddis: %v", err)
	}

	if _, err := client.Set("make me sandwitch", "what? make it yourself", 0).Result(); err != nil {
		log.Fatalf("redis: faield to set key %v", err)
	}

	if _, err := client.Set("sudo make me sandwitch", "okey", 0).Result(); err != nil {
		log.Fatalf("redis: faield to set key %v", err)
	}

	return client
}
