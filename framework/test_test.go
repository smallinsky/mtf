package framework

import (
	"testing"

	pb "github.com/smallinsky/mtf/e2e/proto/echo"
	pbo "github.com/smallinsky/mtf/e2e/proto/oracle"
	"github.com/smallinsky/mtf/port"
)

func TestMain(m *testing.M) {
	NewSuite("suite_first", m).Run()
}

func TestEchoService(t *testing.T) {
	Run(t, new(SuiteTest))
}

func (st *SuiteTest) Init(t *testing.T) {
	var err error
	if st.echoPort, err = port.NewGRPCClient((*pb.EchoClient)(nil), "localhost:8001"); err != nil {
		t.Fatalf("failed to init grpc client port")
	}
	if st.httpPort, err = port.NewHTTP(port.WithTLSHost("*.icndb.com")); err != nil {
		t.Fatalf("failed to init http port")
	}
	if st.oraclePort, err = port.NewGRPCServer((*pbo.OracleServer)(nil), ":8002"); err != nil {
		t.Fatalf("failed to init grpc oracle server")
	}
}

type SuiteTest struct {
	echoPort   *port.ClientPort
	httpPort   *port.HTTPPort
	oraclePort *port.PortIn
}

func (st *SuiteTest) TestRedis(t *testing.T) {
	st.echoPort.SendT(t, &pb.AskRedisRequest{
		Data: "make me sandwitch",
	})
	st.echoPort.ReceiveT(t, &pb.AskRedisResponse{
		Data: "what? make it yourself",
	})
	st.echoPort.SendT(t, &pb.AskRedisRequest{
		Data: "sudo make me sandwitch",
	})
	st.echoPort.ReceiveT(t, &pb.AskRedisResponse{
		Data: "okey",
	})
}

func (st *SuiteTest) TestHTTP(t *testing.T) {
	st.echoPort.SendT(t, &pb.AskGoogleRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.httpPort.ReceiveT(t, &port.HTTPRequest{
		Method: "GET",
	})
	st.httpPort.SendT(t, &port.HTTPResponse{
		Body: []byte(`{"value":{"joke":"42"}}`),
	})
	st.echoPort.ReceiveT(t, &pb.AskGoogleResponse{
		Data: "42",
	})
}

func (st *SuiteTest) TestClientServerGRPC(t *testing.T) {
	st.echoPort.SendT(t, &pb.AskOracleRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.oraclePort.ReceiveT(t, &pbo.AskDeepThroughRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.oraclePort.SendT(t, &pbo.AskDeepThroughRespnse{
		Data: "42",
	})
	st.echoPort.ReceiveT(t, &pb.AskOracleResponse{
		Data: "42",
	})
}

func (st *SuiteTest) TestFetchDataFromDB(t *testing.T) {
	st.echoPort.SendT(t, &pb.AskDBRequest{
		Data: "the dirty fork",
	})
	st.echoPort.ReceiveT(t, &pb.AskDBResponse{
		Data: "Lucky we didn't say anything about the dirty knife",
	})
}
