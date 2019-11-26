// +build mtf

package framework

import (
	"testing"

	"github.com/smallinsky/mtf/framework"
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
	st.echoPort.Send(t, &pb.AskRedisRequest{
		Data: "make me sandwich",
	})
	st.echoPort.Receive(t, &pb.AskRedisResponse{
		Data: "what? make it yourself",
	})
	st.echoPort.Send(t, &pb.AskRedisRequest{
		Data: "sudo make me sandwich",
	})
	st.echoPort.Receive(t, &pb.AskRedisResponse{
		Data: "okey",
	})
}

func (st *SuiteTest) TestHTTP(t *testing.T) {
	st.echoPort.Send(t, &pb.AskGoogleRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.httpPort.Receive(t, &port.HTTPRequest{
		Body:   []byte{},
		Method: "GET",
		Host:   "api.icndb.com",
		URL:    "/jokes/random?firstName=John\u0026amp;lastName=Doe",
	})
	st.httpPort.Send(t, &port.HTTPResponse{
		Body: []byte(`{"value":{"joke":"42"}}`),
	})
	st.echoPort.Receive(t, &pb.AskGoogleResponse{
		Data: "42",
	})
}

func (st *SuiteTest) TestClientServerGRPC(t *testing.T) {
	st.echoPort.Send(t, &pb.AskOracleRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.oraclePort.Receive(t, &pbo.AskDeepThoughtRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.oraclePort.Send(t, &pbo.AskDeepThoughtResponse{
		Data: "42",
	})
	st.echoPort.Receive(t, &pb.AskOracleResponse{
		Data: "42",
	})
}

func (st *SuiteTest) TestFetchDataFromDB(t *testing.T) {
	st.echoPort.Send(t, &pb.AskDBRequest{
		Data: "the dirty fork",
	})
	st.echoPort.Receive(t, &pb.AskDBResponse{
		Data: "Lucky we didn't say anything about the dirty knife",
	})
}
