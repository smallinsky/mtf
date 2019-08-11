package port

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"google.golang.org/grpc"

	"github.com/smallinsky/mtf/match"
)

func TestGrpcClientPort(t *testing.T) {
	port := Port{
		impl: &ClientPort{
			emd: map[reflect.Type]EndpointRespTypePair{
				reflect.TypeOf((*FirstRequest)(nil)): EndpointRespTypePair{
					Endpoint: "FirstMessageHandler",
					RespType: reflect.TypeOf((*FirstResponse)(nil)),
				},
			},
			callResultC: make(chan callResult),
			conn:        &mockConnection{t: t},
		},
	}

	t.Run("SendReceiveOneMessgeSameType", func(t *testing.T) {
		port.Send(t, &FirstRequest{
			ID: 1,
		})
		port.Receive(t, match.Payload(&FirstResponse{
			ID: 1,
		}))
	})

	t.Run("SendReceiveTwoMessageSameType", func(t *testing.T) {
		t.Skipf("fix async call and queue messages")
		port.Send(t, &FirstRequest{
			ID: 1,
		})
		port.Send(t, &FirstRequest{
			ID: 2,
		})

		//TODO: Fix async order based on send call
		port.Receive(t, match.Payload(&FirstResponse{
			ID: 1,
		}))
		port.Receive(t, match.Payload(&FirstResponse{
			ID: 2,
		}))
	})

	t.Run("SendReciveParaller", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				i := i

				port.Send(t, &FirstRequest{
					ID: i,
				})
				port.Receive(t, match.Payload(&FirstResponse{
					ID: i,
				}))
			}()
			wg.Wait()
		}
	})

	t.Run("RecieveMatchFn", func(t *testing.T) {
		port.Send(t, &FirstRequest{
			ID: 10,
		})
		port.Receive(t, match.Fn(
			func(r *FirstResponse) {
				if r.ID != 10 {
					t.Fatalf("expected response id = 10 but got: %v", r.ID)
				}
			},
		))
	})
}

type mockConnection struct {
	t   *testing.T
	err error
}

func (m *mockConnection) Invoke(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error {
	switch t := args.(type) {
	case *FirstRequest:
		resp := FirstResponse{ID: t.ID}
		reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(resp))
	case *SecondRequest:
		resp := FirstResponse{ID: t.ID}
		reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(resp))
	}

	return m.err
}
func (m *mockConnection) Close() error {
	return nil
}

type FirstRequest struct {
	ID int
}

type FirstResponse struct {
	ID int
}

type SecondRequest struct {
	ID int
}

type SecondResponse struct {
	ID int
}
