package main

import (
	pb "github.com/smallinsky/mtf/e2e/proto/echo"
	pbo "github.com/smallinsky/mtf/e2e/proto/oracle"
	"github.com/smallinsky/mtf/port"
)

type TestComponent struct {
	Echo   *port.ClientPort
	Oracle *port.PortIn
	HTTP   *port.HTTPPort
}

func main() {
	httpPort := port.NewHTTP()
	echoPort := port.NewGRPCClient((*pb.EchoClient)(nil), "localhost:8001")
	oraclePort := port.NewGRPCServer((*pbo.OracleServer)(nil), ":8002")

	tc := &TestComponent{
		Echo:   &echoPort,
		Oracle: oraclePort,
		HTTP:   &httpPort,
	}

	testFetchDataFromOtherService(tc)
	testFetchDataFromExternalAPIViaHTTP(tc)
	testFetchDataDB(tc)
	testFetchDataFromRedis(tc)

}

func testFetchDataFromOtherService(tc *TestComponent) {
	tc.Echo.Send(&pb.AskOracleRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	tc.Oracle.Receive(&pbo.AskDeepThroughRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	tc.Oracle.Send(&pbo.AskDeepThroughRespnse{
		Data: "42",
	})
	tc.Echo.Receive(&pb.AskOracleResponse{
		Data: "42",
	})
}

func testFetchDataFromExternalAPIViaHTTP(tc *TestComponent) {
	tc.Echo.Send(&pb.AskGoogleRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	tc.HTTP.Receive(port.HttpRequest{
		Method: "GET",
		URL:    "http://api.icndb.com/jokes/random?firstName=John&amp;lastName=Doe",
	})
	tc.HTTP.Send(port.HttpResponse{
		Body: `{"value":{"joke":"42"}}`,
	})
	tc.Echo.Receive(&pb.AskGoogleResponse{
		Data: "42",
	})
}

func testFetchDataDB(tc *TestComponent) {
	tc.Echo.Send(&pb.AskDBRequest{
		Data: "the dirty fork",
	})

	tc.Echo.Receive(&pb.AskDBResponse{
		Data: "Lucky we didn't say anything about the dirty knife",
	})
}

func testFetchDataFromRedis(tc *TestComponent) {
	tc.Echo.Send(&pb.AskRedisRequest{
		Data: "make me sandwitch",
	})

	tc.Echo.Receive(&pb.AskRedisResponse{
		Data: "what? make it yourself",
	})

	tc.Echo.Send(&pb.AskRedisRequest{
		Data: "sudo make me sandwitch",
	})

	tc.Echo.Receive(&pb.AskRedisResponse{
		Data: "okey",
	})
}
