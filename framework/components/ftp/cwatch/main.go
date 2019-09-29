package main

import (
	"fmt"

	"github.com/smallinsky/mtf/pkg/fswatch"
	pb "github.com/smallinsky/mtf/pkg/fswatch/proto"
)

func main() {
	fswatch.Subscriber(":4441", func(event *pb.EventRequest) {
		fmt.Println("content: ", string(event.GetContent()))
		fmt.Println(event.String())
	})
}
