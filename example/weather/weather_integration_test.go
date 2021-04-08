// +build mtf

package framework

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/smallinsky/mtf/framework"
	"github.com/smallinsky/mtf/mockhttp"
	pb "github.com/smallinsky/mtf/proto/weather"
)

func TestMain(m *testing.M) {
	framework.TestEnv(m).
		WithSUT(framework.SutSettings{
			Dir: "./service",
			Envs: []string{
				"GRPC_PORT=:8082",
				"TLS_CERT_PATH=" + framework.GetTLSCertPath(),
				"TLS_KEY_PATH=" + framework.GetTLSKeyPath(),
				"SCALE_CONV_ADDR=" + framework.GetDockerHostAddr(8083),
			},
			Ports: []int{8082},
		}).
		WithTLS(framework.TLSSettings{
			Hosts: []string{
				"api.weather.com",
			},
		}).
		Run()
}

func TestWeatherService(t *testing.T) {
	framework.Run(t, new(SuiteTest))
}

func (st *SuiteTest) Init(t *testing.T) {
	st.scaleConvServer = pb.NewScaleConvServerMock(":8083")
	st.weatherClient = pb.NewWeatherClientMock(":8082")
	st.httpServer = mockhttp.New()
}

type SuiteTest struct {
	scaleConvServer *pb.ScaleConvServiceMock
	weatherClient   pb.WeatherClient
	httpServer      *mockhttp.Server
}

func (st *SuiteTest) TestFahrenheit(t *testing.T) {
	if true {
		time.Sleep(time.Millisecond * 900)
	}

	st.scaleConvServer.CelsiusToFahrenheit(func(ctx context.Context, req *pb.CelsiusToFahrenheitRequest) (*pb.CelsiusToFahrenheitResponse, error) {
		if got, want := req.GetValue(), int64(12); got != want {
			t.Fatalf("invalid req.value got: %v, want: %v", want, got)
		}
		return &pb.CelsiusToFahrenheitResponse{Value: 10029}, nil
	})

	st.httpServer.Handle(func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(rw, "18")
	})

	resp, err := st.weatherClient.AskAboutWeather(context.Background(), &pb.AskAboutWeatherRequest{
		City:  "Wroclaw",
		Scale: pb.Scale_FAHRENHEIT,
	})

	if err != nil {
		t.Fatal(err)
	}
	if got, want := resp.GetResult(), "fofdoa"; got != want {
		t.Fatalf("mismwatch got: %v, want: %v", got, want)
	}
}
