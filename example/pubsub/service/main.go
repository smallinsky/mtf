package main

import (
	"context"
	"log"

	"cloud.google.com/go/pubsub"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/proto/proto3_proto"
	"github.com/golang/protobuf/ptypes"
)

func main() {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "test-project-id")
	if err != nil {
		log.Fatalf("[ERROR] Failed to init pubsub client: %v", err)
	}
	sub := client.Subscription("testsub")

	sub.ReceiveSettings.Synchronous = true
	c := make(chan struct{})
	go func() {
		err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			defer msg.Ack()
			close(c)
		})
		if err != nil {
			log.Fatalf("[ERROR] Failed to init pubsub receiver function: %v", err)
		}
	}()
	<-c

	topic := client.Topic("testtopic")

	pb := &proto3_proto.Message{
		Name: "bar",
	}
	a, err := ptypes.MarshalAny(pb)
	if err != nil {
		log.Fatalf("[ERROR] Failed to marshal any: %v", err)
	}
	bb, err := proto.Marshal(a)
	if err != nil {
		log.Fatalf("[ERROR] Failed to marshal proto message: %v", err)
	}

	r := topic.Publish(ctx, &pubsub.Message{
		Data: bb,
	})
	_, err = r.Get(ctx)
	if err != nil {
		log.Fatalf("[ERROR] Failed to publish message: %v", err)
	}
}
