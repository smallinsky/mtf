package main

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
)

const (
	queueSize = 100
)

func NewPubsub() *Pubsub {
	ps := &Pubsub{
		messages: make(chan *pubsub.Message, queueSize),
	}
	ctx := context.Background()
	conn, err := pubsub.NewClient(ctx, "ingird-local_test")
	if err != nil {
		panic(err)
	}

	ps.topic = conn.Topic("first-topic")

	sub := conn.Subscription("sub1")
	ok, err := sub.Exists(ctx)
	if err != nil {
		panic(err)
	}

	if !ok {
		sub, err = conn.CreateSubscription(ctx, "sub1", pubsub.SubscriptionConfig{
			Topic:       ps.topic,
			AckDeadline: time.Second * 10,
		})
		if err != nil {
			panic(err)
		}
	}
	sub.ReceiveSettings.Synchronous = true
	go func() {
		err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			ps.messages <- msg
			msg.Ack()
		})
		if err != nil {
			panic(err)
		}
	}()
	return ps

}

type Pubsub struct {
	messages chan *pubsub.Message
	topic    *pubsub.Topic
}

func (p *Pubsub) Receive(i interface{}) {
	select {
	case msg := <-p.messages:
		fmt.Println("Got message: ", string(msg.Data))
	case <-time.Tick(time.Second * 5):
		panic("timout during pubsub.receive")
	}
}

func (p *Pubsub) Send(i interface{}) {
	ctx := context.Background()
	msg := &pubsub.Message{
		Data: []byte("ala ma kota"),
	}

	if _, err := p.topic.Publish(ctx, msg).Get(ctx); err != nil {
		panic(err)
	}
}

func main() {
	p := NewPubsub()
	p.Send(nil)
}
