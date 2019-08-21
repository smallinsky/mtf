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

type PubSubConfig struct {
	TopicSubscriptions []TopicSubscriptions
}

type TopicSubscriptions struct {
	Topic         string
	Subscriptions []string
}

func NewPubsub(projectID, addr string, config PubSubConfig) (*Port, error) {
	if err := os.Setenv("PUBSUB_EMULATOR_HOST", addr); err != nil {
		return nil, err
	}

	ps := &Pubsub{
		messages: make(chan proto.Message, queueSize),
	}
	ctx := context.Background()
	conn, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	for _, ts := range config.TopicSubscriptions {
		ps.topic = conn.Topic(ts.Topic)
		exists, err := ps.topic.Exists(ctx)
		if err != nil {
			return nil, err
		}
		if !exists {
			ps.topic, err = conn.CreateTopic(ctx, ts.Topic)
			if err != nil {
				return nil, err
			}
		}

		for _, subscription := range ts.Subscriptions {
			sub := conn.Subscription(subscription)
			ok, err := sub.Exists(ctx)
			if err != nil {
				return nil, err
			}

			if !ok {
				_, err = conn.CreateSubscription(ctx, subscription, pubsub.SubscriptionConfig{
					Topic:       ps.topic,
					AckDeadline: time.Second * 10,
				})
				if err != nil {
					return nil, err
				}
			}
			subtmp := subscription + "hash"

			sub = conn.Subscription(subtmp)
			ok, err = sub.Exists(ctx)
			if err != nil {
				return nil, err
			}

			if !ok {
				_, err = conn.CreateSubscription(ctx, subtmp, pubsub.SubscriptionConfig{
					Topic:       ps.topic,
					AckDeadline: time.Second * 10,
				})
				if err != nil {
					return nil, err
				}
			}
			sub.ReceiveSettings.Synchronous = true
			go func() {
				err := sub.Receive(ctx, ps.handle)
				if err != nil {
					panic(err)
				}
			}()
		}
	}

	return &Port{
		impl: ps,
	}, nil
}

func (ps *Pubsub) handle(ctx context.Context, msg *pubsub.Message) {
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

func (p *Pubsub) Receive(ctx context.Context) (interface{}, error) {
	return p.receive()
}

func (p *Pubsub) Send(ctx context.Context, i interface{}) error {
	return p.send(i)
}

func (p *Pubsub) receive(opts ...Opt) (interface{}, error) {
	for {
		select {
		case msg := <-p.messages:
			return msg, nil
		case <-time.Tick(time.Second * 10):
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
		return err
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
	for {
		select {
		case f := <-p.messages:
			if !proto.Equal(f, msg) {
				continue
			}
			fmt.Println("receiving done from internal pubsub")
			return nil
		case <-timeout:
			return fmt.Errorf("failed to send message")
		}
	}
}
