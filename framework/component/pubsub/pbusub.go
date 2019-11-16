package pubsub

import (
	"context"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/smallinsky/mtf/pkg/docker"
)

type Component struct {
	Config    Config
	Container docker.Container
}

func New(cli *docker.Docker, config Config) (*Component, error) {
	containerConf, err := BuildContainerConfig()
	if err != nil {
		return nil, err
	}

	container, err := cli.NewContainer(*containerConf)
	if err != nil {
		return nil, err
	}

	return &Component{
		Config:    config,
		Container: container,
	}, nil
}

func (c *Component) Start(ctx context.Context) error {
	if err := c.Container.Start(ctx); err != nil {
		return err
	}

	if err := os.Setenv("PUBSUB_EMULATOR_HOST", ":8085"); err != nil {
		return err
	}

	conn, err := pubsub.NewClient(ctx, c.Config.ProjectID)
	if err != nil {
		return err
	}

	for _, ts := range c.Config.TopicSubscriptions {
		topic := conn.Topic(ts.Topic)
		exists, err := topic.Exists(ctx)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		topic, err = conn.CreateTopic(ctx, ts.Topic)
		if err != nil {
			return err
		}

		for _, sn := range ts.Subscriptions {
			sub := conn.Subscription(sn)
			exists, err := sub.Exists(ctx)
			if err != nil {
				return err
			}
			if !exists {
				_, err = conn.CreateSubscription(ctx, sn, pubsub.SubscriptionConfig{
					Topic:       topic,
					AckDeadline: time.Second * 10,
				})
			}
		}
	}

	//  Give some time to pubsub emulator to process requests
	time.Sleep(time.Millisecond * 500)

	return nil
}

func (c *Component) Stop(ctx context.Context) error {
	return c.Container.Stop(ctx)
}
