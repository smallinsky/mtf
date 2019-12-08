// +build mtf

package framework

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smallinsky/mtf/framework"
	"github.com/smallinsky/mtf/match"
	"github.com/smallinsky/mtf/port"
	pb "github.com/smallinsky/mtf/proto/echo"
	pbo "github.com/smallinsky/mtf/proto/oracle"
)

func TestMain(m *testing.M) {
	framework.TestEnv(m).
		WithSUT(framework.SutSettings{
			Dir:   "./service",
			Ports: []int{8001},
			Envs: []string{
				"ORACLE_ADDR=" + framework.GetDockerHostAddr(8002),
			}}).
		WithRedis(framework.RedisSettings{
			Password: "test",
		}).
		WithMySQL(framework.MysqlSettings{
			DatabaseName: "test_db",
			MigrationDir: "./service/migrations",
			Password:     "test",
		}).Run()
}

func TestEchoService(t *testing.T) {
	framework.Run(t, new(SuiteTest))
}

func (st *SuiteTest) Init(t *testing.T) {
	var err error
	if st.echoPort, err = port.NewGRPCClientPort((*pb.EchoClient)(nil), "localhost:8001"); err != nil {
		t.Fatalf("failed to init grpc client port")
	}
	st.httpPort = port.NewHTTPPort()
	if st.oraclePort, err = port.NewGRPCServerPort((*pbo.OracleServer)(nil), ":8002"); err != nil {
		t.Fatalf("failed to init grpc oracle server")
	}
}

type SuiteTest struct {
	echoPort   *port.Port
	httpPort   *port.Port
	oraclePort *port.Port
}

func (st *SuiteTest) TestRedis(t *testing.T) {
	st.echoPort.Send(&pb.AskRedisRequest{
		Data: "make me sandwich",
	})
	st.echoPort.Receive(&pb.AskRedisResponse{
		Data: "what? make it yourself",
	})
	st.echoPort.Send(&pb.AskRedisRequest{
		Data: "sudo make me sandwich",
	})
	st.echoPort.Receive(&pb.AskRedisResponse{
		Data: "okey",
	})
}

func (st *SuiteTest) TestHTTP(t *testing.T) {
	st.echoPort.Send(&pb.AskGoogleRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.httpPort.Receive(&port.HTTPRequest{
		Method: "GET",
		Host:   "api.icndb.com",
		URL:    "/jokes/random?firstName=John\u0026amp;lastName=Doe",
	})
	st.httpPort.Send(&port.HTTPResponse{
		Body: []byte(`{"value":{"joke":"42"}}`),
	})
	st.echoPort.Receive(&pb.AskGoogleResponse{
		Data: "42",
	})
}

func (st *SuiteTest) TestClientServerGRPC(t *testing.T) {
	st.echoPort.Send(&pb.AskOracleRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.oraclePort.Receive(&pbo.AskDeepThoughtRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.oraclePort.Send(&pbo.AskDeepThoughtResponse{
		Data: "42",
	})
	st.echoPort.Receive(&pb.AskOracleResponse{
		Data: "42",
	})
}

func (st *SuiteTest) TestClientServerGRPCError(t *testing.T) {
	st.echoPort.Send(&pb.AskOracleRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.oraclePort.Receive(&pbo.AskDeepThoughtRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.oraclePort.Send(&port.GRPCErr{
		Err: status.Errorf(codes.FailedPrecondition, "Deepthought error"),
	})
	st.echoPort.Receive(&pb.AskOracleResponse{
		Data: "Come back after seven and a half million years",
	})
}

func (st *SuiteTest) TestClientServerGRPCErrorMatch(t *testing.T) {
	st.echoPort.Send(&pb.AskOracleRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.oraclePort.Receive(&pbo.AskDeepThoughtRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.oraclePort.Send(&port.GRPCErr{
		Err: status.Errorf(codes.Internal, "internal error"),
	})
	st.echoPort.Receive(match.GRPCStatusCode(codes.Internal))
}

func (st *SuiteTest) TestFetchDataFromDB(t *testing.T) {
	st.echoPort.Send(&pb.AskDBRequest{
		Data: "the dirty fork",
	})
	st.echoPort.Receive(&pb.AskDBResponse{
		Data: "Lucky we didn't say anything about the dirty knife",
	})
}

func (st *SuiteTest) TestHTTPMatcher(t *testing.T) {
	st.echoPort.Send(&pb.AskGoogleRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.httpPort.Receive(match.Fn(func(req *port.HTTPRequest) {
		if got, want := req.Host, "api.icndb.com"; got != want {
			t.Fatalf("host mismatch, got: %v want: %v", got, want)
		}
	}))
	st.httpPort.Send(&port.HTTPResponse{
		Body: []byte(`{"value":{"joke":"42"}}`),
	})
	st.echoPort.Receive(match.Fn(func(resp *pb.AskGoogleResponse) {
		if got, want := resp.GetData(), "42"; got != want {
			t.Fatalf("data mismatch, got: %v want: %v", got, want)
		}
	}))
}
