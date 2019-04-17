package main

import (
	"fmt"
	"time"

	pb "github.com/smallinsky/mtf/e2e/proto/echo"
	pbo "github.com/smallinsky/mtf/e2e/proto/oracle"
	"github.com/smallinsky/mtf/port/grpc"
)

func main() {
	grpcEcho := grpc.NewClient((*pb.EchoClient)(nil), "localhost:8001")
	grpcOracle := grpc.NewServer((*pbo.OracleServer)(nil), ":8002")

	// Wait for sut connection to grpc oracle server port.
	// TODO: Prabably condition if test can be start can be deduce based on grpc connection state (status).
	time.Sleep(time.Second * 2)

	grpcEcho.Send(&pb.AskOracleRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})

	grpcOracle.Receive(&pbo.AskDeepThroughRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})

	fmt.Println("---> send echo.AskDeepThroughRespnse")
	grpcOracle.Send(&pbo.AskDeepThroughRespnse{
		Data: "42",
	})

	fmt.Println("<--- receive echo.AskOracleResponse")
	grpcEcho.Receive(&pb.AskOracleResponse{
		Data: "42",
	})
}
