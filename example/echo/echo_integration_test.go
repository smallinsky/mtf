// +build mtf

package framework

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smallinsky/mtf/framework"
	"github.com/smallinsky/mtf/match"
	"github.com/smallinsky/mtf/port"
	pbecho "github.com/smallinsky/mtf/proto/echo"
	pboracle "github.com/smallinsky/mtf/proto/oracle"
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
		WithMigration([]*framework.MigrationSettings{
			{
				Password: "test",
				Port:     "3306",
				DBName:   "events_db",
				Dir:      "./service/migrations",
			},
			{
				Password: "test",
				Port:     "3306",
				DBName:   "test_db",
				Dir:      "./service/migrations",
			},
		}).
		WithMySQL(framework.MysqlSettings{
			DatabaseName: "test_db",
			Databases:    []string{"events_db"},
			MigrationDir: "./service/migrations",
			Password:     "test",
		}).Run()
}

func TestEchoService(t *testing.T) {
	framework.Run(t, new(SuiteTest))
}

func (st *SuiteTest) Init(t *testing.T) {
	var err error
	if st.echoPort, err = port.NewGRPCClientPort((*pbecho.EchoClient)(nil), "localhost:8001"); err != nil {
		t.Fatalf("failed to init grpc client port")
	}
	st.httpPort = port.NewHTTPPort()
	if st.oraclePort, err = port.NewGRPCServerPort((*pboracle.OracleServer)(nil), ":8002"); err != nil {
		t.Fatalf("failed to init grpc oracle server")
	}
}

type SuiteTest struct {
	echoPort   *port.Port
	httpPort   *port.Port
	oraclePort *port.Port
}

func (st *SuiteTest) TestRedis(t *testing.T) {
	st.echoPort.Send(t, &pbecho.AskRedisRequest{
		Data: "make me sandwich",
	})
	st.echoPort.Receive(t, &pbecho.AskRedisResponse{
		Data: "what? make it yourself",
	})
	st.echoPort.Send(t, &pbecho.AskRedisRequest{
		Data: "sudo make me sandwich",
	})
	st.echoPort.Receive(t, &pbecho.AskRedisResponse{
		Data: "okey",
	})
}

func (st *SuiteTest) TestHTTP(t *testing.T) {
	st.echoPort.Send(t, &pbecho.AskGoogleRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.httpPort.Receive(t, &port.HTTPRequest{
		Method: "GET",
		Host:   "api.icndb.com",
		URL:    "/jokes/random?firstName=John\u0026amp;lastName=Doe",
	})
	st.httpPort.Send(t, &port.HTTPResponse{
		Body: []byte(`{"value":{"joke":"42"}}`),
	})
	st.echoPort.Receive(t, &pbecho.AskGoogleResponse{
		Data: "42",
	})
}

func (st *SuiteTest) TestClientServerGRPC(t *testing.T) {
	st.echoPort.Send(t, &pbecho.AskOracleRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.oraclePort.Receive(t, &pboracle.AskDeepThoughtRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.oraclePort.Send(t, &pboracle.AskDeepThoughtResponse{
		Data: "42",
	})
	st.echoPort.Receive(t, &pbecho.AskOracleResponse{
		Data: "42",
	})
}

func (st *SuiteTest) TestClientServerGRPCError(t *testing.T) {
	st.echoPort.Send(t, &pbecho.AskOracleRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.oraclePort.Receive(t, &pboracle.AskDeepThoughtRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.oraclePort.Send(t, &port.GRPCErr{
		Err: status.Errorf(codes.FailedPrecondition, "Deepthought error"),
	})
	st.echoPort.Receive(t, &pbecho.AskOracleResponse{
		Data: "Come back after seven and a half million years",
	})
}

func (st *SuiteTest) TestClientServerGRPCErrorMatch(t *testing.T) {
	st.echoPort.Send(t, &pbecho.AskOracleRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.oraclePort.Receive(t, &pboracle.AskDeepThoughtRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.oraclePort.Send(t, &port.GRPCErr{
		Err: status.Errorf(codes.Internal, "internal error"),
	})
	st.echoPort.Receive(t, match.GRPCStatusCode(codes.Internal))
}

func (st *SuiteTest) TestFetchDataFromDB(t *testing.T) {
	st.echoPort.Send(t, &pbecho.AskDBRequest{
		Data: "the dirty fork",
	})
	st.echoPort.Receive(t, &pbecho.AskDBResponse{
		Data: "Lucky we didn't say anything about the dirty knife",
	})
}

func (st *SuiteTest) TestHTTPMatcher(t *testing.T) {
	st.echoPort.Send(t, &pbecho.AskGoogleRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.httpPort.Receive(t, match.Fn(func(req *port.HTTPRequest) {
		if got, want := req.Host, "api.icndb.com"; got != want {
			t.Fatalf("host mismatch, got: %v want: %v", got, want)
		}
	}))
	st.httpPort.Send(t, &port.HTTPResponse{
		Body: []byte(`{"value":{"joke":"42"}}`),
	})
	st.echoPort.Receive(t, match.Fn(func(resp *pbecho.AskGoogleResponse) {
		if got, want := resp.GetData(), "42"; got != want {
			t.Fatalf("data mismatch, got: %v want: %v", got, want)
		}
	}))
}
