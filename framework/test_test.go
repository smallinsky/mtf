package framework

import (
	"testing"

	pb "github.com/smallinsky/mtf/e2e/proto/echo"
	"github.com/smallinsky/mtf/port"
)

func TestMain(m *testing.M) {
	NewSuite("suite_first", m).Run()
}

func TestFoo(t *testing.T) {
	echoPort := port.NewGRPCClient((*pb.EchoClient)(nil), "localhost:8001")
	httpPort := port.NewHTTP()

	echoPort.Send(&pb.AskRedisRequest{
		Data: "make me sandwitch",
	})

	echoPort.Receive(&pb.AskRedisResponse{
		Data: "what? make it yourself",
	})

	echoPort.Send(&pb.AskRedisRequest{
		Data: "sudo make me sandwitch",
	})
	echoPort.Receive(&pb.AskRedisResponse{
		Data: "okey",
	})

	echoPort.Send(&pb.AskGoogleRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	httpPort.Receive(&port.HTTPRequest{
		Method: "GET",
	})
	httpPort.Send(&port.HTTPResponse{
		Body: []byte(`{"value":{"joke":"42"}}`),
	})
	echoPort.Receive(&pb.AskGoogleResponse{
		Data: "42",
	})
}

func TestBar(t *testing.T) {
}
