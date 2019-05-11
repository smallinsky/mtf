package framework

import (
	"testing"

	pb "github.com/smallinsky/mtf/e2e/proto/echo"
	"github.com/smallinsky/mtf/port"
)

func TestMain(m *testing.M) {
	NewSuite("suite_first", m).Run()
}

type SuiteTest struct {
	echoPort port.ClientPort
	httpPort port.HTTPPort
}

func (st *SuiteTest) Init(t *testing.T) {
	st.echoPort = port.NewGRPCClient((*pb.EchoClient)(nil), "localhost:8001")
	st.httpPort = port.NewHTTP()
}

func (st *SuiteTest) TestRedis(t *testing.T) {
	st.echoPort.Send(&pb.AskRedisRequest{
		Data: "make me sandwitch",
	})

	st.echoPort.Receive(&pb.AskRedisResponse{
		Data: "what? make it yourself",
	})

	st.echoPort.Send(&pb.AskRedisRequest{
		Data: "sudo make me sandwitch",
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
	})
	st.httpPort.Send(&port.HTTPResponse{
		Body: []byte(`{"value":{"joke":"42"}}`),
	})
	st.echoPort.Receive(&pb.AskGoogleResponse{
		Data: "42",
	})
}

func TestEchoService(t *testing.T) {
	Run(t, new(SuiteTest))
}
