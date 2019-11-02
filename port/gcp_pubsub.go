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
		topicMap: make(map[string]*pubsub.Topic),
	}
	ctx := context.Background()
	conn, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	for _, ts := range config.TopicSubscriptions {
		var topic *pubsub.Topic
		topic = conn.Topic(ts.Topic)
		exists, err := topic.Exists(ctx)
		if err != nil {
			return nil, err
		}
		if !exists {
			fmt.Println("creating", ts.Topic)
			topic, err = conn.CreateTopic(ctx, ts.Topic)
			if err != nil {
				return nil, err
			}
		}
		ps.topicMap[ts.Topic] = topic
		ps.topic = topic

		for _, subscription := range ts.Subscriptions {
			sub := conn.Subscription(subscription)
			ok, err := sub.Exists(ctx)
			if err != nil {
				return nil, err
			}

			if !ok {
				_, err = conn.CreateSubscription(ctx, subscription, pubsub.SubscriptionConfig{
					Topic:       topic,
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
					Topic:       topic,
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

//func waitForTopicCreated(cli *pubsub.Client, cfg PubSubConfig) error {
//	var wg sync.WaitGroup
//
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
//	defer cancel()
//
//	for _, v := range cfg.TopicSubscriptions {
//		wg.Add(1)
//		go func(topic string) {
//			defer wg.Done()
//			for {
//				select {
//				case <-ctx.Done():
//					return
//				default:
//					exists, err := cli.Topic(topic).Exists(context.Background())
//					if err != nil {
//						log.Fatalf("failed to check if topic exists: %v", err)
//					}
//					if !exists {
//						continue
//					}
//					fmt.Println("topic exists")
//					return
//				}
//			}
//
//		}(v.Topic)
//	}
//
//	wg.Wait()
//	return nil
//}

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
			fmt.Println("receiving done from internal pubsub")
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
			fmt.Println("receiving done from internal pubsub")
			return nil
		case <-timeout:
			return fmt.Errorf("failed to send message")
		}
	}
}
