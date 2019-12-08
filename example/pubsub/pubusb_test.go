// +build mtf

package framework

import (
	"testing"

	"github.com/golang/protobuf/proto/proto3_proto"

	"github.com/smallinsky/mtf/framework"
	"github.com/smallinsky/mtf/port"
)

func TestMain(m *testing.M) {
	framework.TestEnv(m).
		WithSUT(framework.SutSettings{
			Envs: []string{
				"ORACLE_ADDR=" + framework.GetDockerHostAddr(8002),
			},
			Dir:         "./service",
			RuntimeType: framework.RuntimeTypeCommand,
		}).
		WithPubSub(framework.PubSubSettings{
			ProjectID: "test-project-id",
			TopicSubscriptions: []framework.TopicSubscriptions{
				{
					Topic:         "testtopic",
					Subscriptions: []string{"testsub"},
				},
			},
		}).
		Run()
}

func TestPubSub(t *testing.T) {
	framework.Run(t, new(SuiteTest))
}

func (st *SuiteTest) Init(t *testing.T) {
	pusbus, err := port.NewPubsub("test-project-id", "localhost:8085")
	if err != nil {
		t.Fatalf("Failed to create pbusub %v", err)
	}
	st.pubsub = pusbus
}

type SuiteTest struct {
	pubsub *port.Port
}

func (st *SuiteTest) TestPubsub(t *testing.T) {
	st.pubsub.Send(&proto3_proto.Message{
		Name: "test message",
	})
	st.pubsub.Receive(&proto3_proto.Message{
		Name: "bar",
	})
}
