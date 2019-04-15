package main

import (
	"fmt"
	"time"

	pb "github.com/smallinsky/mtf/e2e/proto/echo"
	pbo "github.com/smallinsky/mtf/e2e/proto/oracle"
	"github.com/smallinsky/mtf/port/grpc"
)

func main() {
	grpcOracle := grpc.NewServer((*pbo.OracleServer)(nil), ":8002")
	grpcEcho := grpc.NewClient((*pb.EchoClient)(nil), "localhost:8001")

	if false {
		go func() {
			for {
				grpcOracle.Receive(&pbo.AskDeepThroughRequest{
					Data: "alamakota",
				})

				grpcOracle.Send(&pbo.AskDeepThroughRespnse{
					Data: "42",
				})
			}
		}()
	}
	time.Sleep(time.Second * 3)

	fmt.Println("---> send echo.AskOracleRequest")
	grpcEcho.Send(&pb.AskOracleRequest{
		Data: "ala ma kota",
	})

	fmt.Println("<--- receive oracle.AskDeepThroughRequest")
	grpcOracle.Receive(&pbo.AskDeepThroughRequest{
		Data: "ala ma kota",
	})

	fmt.Println("---> send oracle.AskDeepThroughResponse")
	grpcOracle.Send(&pbo.AskDeepThroughRespnse{
		Data: "42",
	})

	fmt.Println("<--- receive echo.AskOracleResponse")

	grpcEcho.Receive(&pb.AskOracleResponse{
		Data: "42",
	})

	time.Sleep(time.Second * 20)

	fmt.Println("vim-go")
	time.Sleep(time.Second * 2)
}
