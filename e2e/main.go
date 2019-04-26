package main

import (
	"fmt"

	pb "github.com/smallinsky/mtf/e2e/proto/echo"
	pbo "github.com/smallinsky/mtf/e2e/proto/oracle"
	"github.com/smallinsky/mtf/port"
)

func main() {
	grpcEcho := port.NewGRPCClient((*pb.EchoClient)(nil), "localhost:8001")
	grpcOracle := port.NewGRPCServer((*pbo.OracleServer)(nil), ":8002")
	http := port.NewHTTP()

	grpcEcho.Send(&pb.AskOracleRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	grpcOracle.Receive(&pbo.AskDeepThroughRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	grpcOracle.Send(&pbo.AskDeepThroughRespnse{
		Data: "42",
	})
	grpcEcho.Receive(&pb.AskOracleResponse{
		Data: "42",
	})
	fmt.Println("PASS: Integration with other grpc service")

	// -- checking http forwarding
	grpcEcho.Send(&pb.AskGoogleRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	http.Receive(port.HttpRequest{
		Method: "GET",
		URL:    "http://api.icndb.com/jokes/random?firstName=John&amp;lastName=Doe",
	})
	http.Send(port.HttpResponse{
		Body: `{"value":{"joke":"42"}}`,
	})
	grpcEcho.Receive(&pb.AskGoogleResponse{
		Data: "42",
	})

	fmt.Println("PASS: Handling HTTP trafic from SUT")

	grpcEcho.Send(&pb.AskDBRequest{
		Data: "the dirty fork",
	})

	grpcEcho.Receive(&pb.AskDBResponse{
		Data: "Lucky we didn't say anything about the dirty knife",
	})

	fmt.Println("PASS: DB integration")
}
