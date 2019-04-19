package main

import (
	"fmt"

	pb "github.com/smallinsky/mtf/e2e/proto/echo"
	pbo "github.com/smallinsky/mtf/e2e/proto/oracle"
	"github.com/smallinsky/mtf/port/grpc"
)

func main() {
	grpcOracle := grpc.NewServer((*pbo.OracleServer)(nil), ":8002")
	grpcEcho := grpc.NewClient((*pb.EchoClient)(nil), "localhost:8001")

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
