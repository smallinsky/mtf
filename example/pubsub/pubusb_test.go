package framework

import (
	"testing"

	"github.com/golang/protobuf/proto/proto3_proto"
	"github.com/smallinsky/mtf/framework"
	"github.com/smallinsky/mtf/port"
)

func TestMain(m *testing.M) {
	framework.NewSuite(m).
		WithSut(framework.SutSettings{
			Envs: []string{
				"ORACLE_ADDR=host.docker.internal:8002",
			},
			Dir: "./service",
		}).
		WithPubSub(framework.PubSubSettings{}).
		Run()
}

func TestPubSub(t *testing.T) {
	framework.Run(t, new(SuiteTest))
}

func (st *SuiteTest) Init(t *testing.T) {
	pusbus, err := port.NewPubsub("test-project-id", "localhost:8085", port.PubSubConfig{
		TopicSubscriptions: []port.TopicSubscriptions{
			{
				Topic:         "testtopic",
				Subscriptions: []string{"testsub"},
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to create pbusub %v", err)
	}
	st.pubsub = pusbus

}

type SuiteTest struct {
	pubsub *port.Port
}

func (st *SuiteTest) TestPubsub(t *testing.T) {
	st.pubsub.Send(t, &proto3_proto.Message{
		Name: "test message",
	})
	st.pubsub.Receive(t, &proto3_proto.Message{
		Name: "bar",
	})
}
