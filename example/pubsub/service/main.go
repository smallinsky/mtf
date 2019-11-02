package main

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/proto/proto3_proto"
	"github.com/golang/protobuf/ptypes"
)

func main() {
	// fix pbusub Healthcheck
	time.Sleep(time.Second * 3)
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "test-project-id")
	if err != nil {
		panic(err)
	}
	sub := client.Subscription("testsub")

	fmt.Println("receiveing ...")
	c := make(chan struct{})
	go func() {
		err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			defer msg.Ack()
			fmt.Println("received: ", msg.ID, " ", msg.Data)
			close(c)
		})
	}()
	<-c
	if err != nil {
		fmt.Println("failed to receive from sub: ", err)
		panic(err)
	}

	topic := client.Topic("testtopic")

	pb := &proto3_proto.Message{
		Name: "bar",
	}
	a, err := ptypes.MarshalAny(pb)
	if err != nil {
		panic(err)
	}
	bb, err := proto.Marshal(a)
	if err != nil {
		panic(err)
	}

	fmt.Println("publish message")
	r := topic.Publish(ctx, &pubsub.Message{
		Data: bb,
	})
	_, err = r.Get(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println("publish message done")

}
