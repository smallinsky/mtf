// +build mtf

package framework

import (
	"context"
	"testing"

	"github.com/smallinsky/mtf/framework"
	"github.com/smallinsky/mtf/port"
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
	st.scaleConvClient = pb.NewScaleConvClientMock(":8082")
}

type SuiteTest struct {
	weatherPort     *port.Port
	convPort        *port.Port
	httpPort        *port.Port
	scaleConvServer *pb.ScaleConvServiceMock
	scaleConvClient pb.ScaleConvClient
}

//func (st *SuiteTest) TestWeatherCelsius(t *testing.T) {
//	st.weatherPort.Send(t, &pb.AskAboutWeatherRequest{
//		City: "Wroclaw",
//	})
//	st.httpPort.Receive(t, &port.HTTPRequest{
//		Method: "GET",
//		Host:   "api.weather.com",
//		URL:    "/",
//	})
//	st.httpPort.Send(t, &port.HTTPResponse{
//		Status: 200,
//		Body:   []byte("15 Celsius Degrees"),
//	})
//	st.weatherPort.Receive(t, &pb.AskAboutWeatherResponse{
//		Result: "15 Celsius Degrees",
//	})
//}

//func (st *SuiteTest) TestFahrenheitConvErr(t *testing.T) {
//	st.weatherPort.Send(t, &pb.AskAboutWeatherRequest{
//		City:  "Wroclaw",
//		Scale: pb.Scale_FAHRENHEIT,
//	})
//	st.httpPort.Receive(t, &port.HTTPRequest{
//		Method: "GET",
//		Host:   "api.weather.com",
//		URL:    "/",
//	})
//	st.httpPort.Send(t, &port.HTTPResponse{
//		Status: 200,
//		Body:   []byte("15 Celsius Degrees"),
//	})
//	st.weatherPort.Receive(t, match.GRPCErr(codes.Internal, "failed to convert http response to int"))
//}

func (st *SuiteTest) TestFahrenheit(t *testing.T) {
	st.scaleConvServer.CelsiusToFahrenheit(func(ctx context.Context, req *pb.CelsiusToFahrenheitRequest) (*pb.CelsiusToFahrenheitResponse, error) {
		if got, want := req.GetValue(), int64(11); got != want {
			t.Fatalf("invalid req.value got: %v, want: %v", want, got)
		}
		return &pb.CelsiusToFahrenheitResponse{Value: 10029}, nil
	})

	st.scaleConvClient.CelsiusToFahrenheit(context.Background(), &pb.CelsiusToFahrenheitRequest{
		Value: 17,
	})
}

//func (st *SuiteTest) TestFahrenheitConvGRPCErr(t *testing.T) {
//	st.weatherPort.Send(t, &pb.AskAboutWeatherRequest{
//		City:  "Wroclaw",
//		Scale: pb.Scale_FAHRENHEIT,
//	})
//	st.httpPort.Receive(t, &port.HTTPRequest{
//		Method: "GET",
//		Host:   "api.weather.com",
//		URL:    "/",
//	})
//	st.httpPort.Send(t, &port.HTTPResponse{
//		Status: 200,
//		Body:   []byte("15"),
//	})
//	st.convPort.Receive(t, &pb.CelsiusToFahrenheitRequest{
//		Value: 15,
//	})
//	st.convPort.Send(t, &port.GRPCErr{
//		Err: status.Error(codes.FailedPrecondition, "not supported"),
//	})
//	st.weatherPort.Receive(t, match.GRPCErr(codes.Internal, "cale conv client failed"))
//}
