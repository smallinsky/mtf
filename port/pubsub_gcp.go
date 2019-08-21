package port

import (
	"context"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	log "github.com/sirupsen/logrus"
)

const (
	queueSize = 100
)

func NewPubsub(projectID, topicName string, addr string) *Port {
	if err := os.Setenv("PUBSUB_EMULATOR_HOST", addr); err != nil {
		log.Fatalf("failed to set env: %v", err)
	}

	ps := &Pubsub{
		messages: make(chan proto.Message, queueSize),
	}
	ctx := context.Background()
	conn, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		panic(err)
	}

	ps.topic = conn.Topic(topicName)
	exists, err := ps.topic.Exists(ctx)
	if err != nil {
		panic(err)
	}

	if !exists {
		ps.topic, err = conn.CreateTopic(ctx, topicName)
		if err != nil {
			panic(err)
		}
	}

	subName := fmt.Sprintf("sub-%d", time.Now().Nanosecond()/100000)
	sub := conn.Subscription(subName)
	ok, err := sub.Exists(ctx)
	if err != nil {
		panic(err)
	}

	if !ok {
		sub, err = conn.CreateSubscription(ctx, subName, pubsub.SubscriptionConfig{
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
			m := any.Any{}
			err := proto.Unmarshal(msg.Data, &m)
			if err != nil {
				log.Infof("recived messge ID: %v Data:%v\n", msg.ID, string(msg.Data))
				panic(err)
			}

			dynAny := ptypes.DynamicAny{}
			if err := ptypes.UnmarshalAny(&m, &dynAny); err != nil {
				panic(err)
			}
			msg.Ack()
			ps.messages <- dynAny.Message

		})
		if err != nil {
			panic(err)
		}
	}()

	return &Port{
		impl: ps,
	}
}

func (ps *Pubsub) Kind() Kind {
	return KIND_MESSAGE_QEUEU
}

func (ps *Pubsub) Name() string {
	return "pubsub_message_queue"
}

type Pubsub struct {
	messages chan proto.Message
	queue    []proto.Message
	topic    *pubsub.Topic
}

func (p *Pubsub) Receive() (interface{}, error) {
	return p.Receive()
}

func (p *Pubsub) Send(i interface{}) error {
	return p.send(i)
}

func (p *Pubsub) receive(opts ...Opt) (interface{}, error) {
	for {
		select {
		case msg := <-p.messages:
			return msg, nil
		case <-time.Tick(time.Second * 5):
			return nil, fmt.Errorf("timout during pubsub.receive")
		}
	}
}

func (p *Pubsub) send(i interface{}) error {
	msg, ok := i.(proto.Message)
	if !ok {
		return fmt.Errorf("message is not a proto.Message")
	}

	ctx := context.Background()
	anyMsg, err := ptypes.MarshalAny(msg)
	if err != nil {
		panic(err)
	}

	buf, err := proto.Marshal(anyMsg)
	if err != nil {
		return err
	}

	m := &pubsub.Message{
		Data: buf,
	}

	if _, err := p.topic.Publish(ctx, m).Get(ctx); err != nil {
		return err
	}

	timeout := time.After(time.Second * 5)
	// TODO: stop other recivers
	// resend recived message to queue.
	for {
		select {
		case f := <-p.messages:
			if !proto.Equal(f, msg) {
				continue
			}
		case <-timeout:
			return fmt.Errorf("failed to send message")
		}
	}
}
