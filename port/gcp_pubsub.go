package port

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
)

const (
	queueSize = 100
)

func NewPubsub(projectID, addr string) (*Port, error) {
	if err := os.Setenv("PUBSUB_EMULATOR_HOST", addr); err != nil {
		return nil, err
	}

	ps := &Pubsub{
		messages: make(chan proto.Message, queueSize),
		topicMap: make(map[string]*pubsub.Topic),
	}
	ctx := context.Background()
	conn, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	titer := conn.Topics(ctx)
	for {
		t, err := titer.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			panic(err)
		}

		ps.topicMap[portTopicName(t.String())] = t
		ps.topic = t

		siter := t.Subscriptions(ctx)

		for {
			sub, err := siter.Next()
			if err != nil {
				if err == iterator.Done {
					break
				}
				panic(err)
			}

			sm, err := conn.CreateSubscription(ctx, portSubscriptionName(sub.String()), pubsub.SubscriptionConfig{
				Topic:       t,
				AckDeadline: time.Second * 10,
			})
			if err != nil {
				return nil, err
			}
			go func() {
				err := sm.Receive(ctx, ps.handle)
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

func portSubscriptionName(s string) string {
	ss := strings.SplitAfter(s, "/subscriptions/")
	if len(ss) != 2 {
		panic("corupted subscription name")
	}
	return fmt.Sprintf("%s_mtf_port_receiver", ss[1])
}

func portTopicName(s string) string {
	ss := strings.SplitAfter(s, "/topics/")
	if len(ss) != 2 {
		panic("corupted topic name")
	}
	return ss[1]
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

type Pubsub struct {
	messages chan proto.Message
	queue    []proto.Message
	topic    *pubsub.Topic
	topicMap map[string]*pubsub.Topic
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

type PubSubSendRequest struct {
	Topic   string
	Message proto.Message
}

func (p *Pubsub) sendToTopic(msg *PubSubSendRequest) error {
	ctx := context.Background()
	anyMsg, err := ptypes.MarshalAny(msg.Message)
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

	topic, ok := p.topicMap[msg.Topic]
	if !ok {
		return fmt.Errorf("topic not found")
	}

	if _, err := topic.Publish(ctx, m).Get(ctx); err != nil {
		return err
	}

	timeout := time.After(time.Second * 2)
	for {
		select {
		case f := <-p.messages:
			if !proto.Equal(f, msg.Message) {
				continue
			}
			return nil
		case <-timeout:
			return fmt.Errorf("failed to send message")
		}
	}
}

func (p *Pubsub) send(i interface{}) error {
	if msg, ok := i.(*PubSubSendRequest); ok {
		return p.sendToTopic(msg)
	}

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

	timeout := time.After(time.Second * 2)
	for {
		select {
		case f := <-p.messages:
			if !proto.Equal(f, msg) {
				continue
			}
			return nil
		case <-timeout:
			return fmt.Errorf("failed to send message")
		}
	}
}
