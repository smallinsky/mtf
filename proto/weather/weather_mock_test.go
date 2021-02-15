package weather

import (
	"context"
	"testing"
)

func TestHelloWorld(t *testing.T) {
	server := NewScaleConvServerMock()
	client := NewScaleConvClientMock(server.Addr())

	server.CelsiusToFahrenheit(func(ctx context.Context, req *CelsiusToFahrenheitRequest) (*CelsiusToFahrenheitResponse, error) {
		if got, want := req.GetValue(), int64(170); got != want {
			t.Fatalf("got: %v, want: %v", got, want)
		}
		return &CelsiusToFahrenheitResponse{
			Value: 100,
		}, nil
	})

	resp, err := client.CelsiusToFahrenheit(context.Background(), &CelsiusToFahrenheitRequest{
		Value: 170,
	})

	if err != nil {
		t.Fatalf("got err: %v", err)
	}

	if got, want := resp.GetValue(), int64(100); got != want {
		t.Fatalf("got: %v, want: %v", got, want)
	}
}
