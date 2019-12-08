// +build mtf

package framework

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smallinsky/mtf/framework"
	"github.com/smallinsky/mtf/match"
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
	var err error
	if st.weatherPort, err = port.NewGRPCClientPort((*pb.WeatherClient)(nil), "localhost:8082", port.WithTLS()); err != nil {
		t.Fatalf("failed to init grpc client port")
	}

	if st.convPort, err = port.NewGRPCServerPort((*pb.ScaleConvServer)(nil), ":8083", port.WithTLS()); err != nil {
		t.Fatalf("failed to init grpc server port")
	}
	st.httpPort = port.NewHTTPPort()
}

type SuiteTest struct {
	weatherPort *port.Port
	convPort    *port.Port
	httpPort    *port.Port
}

func (st *SuiteTest) TestWeatherCelsius(t *testing.T) {
	st.weatherPort.Send(&pb.AskAboutWeatherRequest{
		City: "Wroclaw",
	})
	st.httpPort.Receive(&port.HTTPRequest{
		Method: "GET",
		Host:   "api.weather.com",
		URL:    "/",
	})
	st.httpPort.Send(&port.HTTPResponse{
		Status: 200,
		Body:   []byte("15 Celsius Degrees"),
	})
	st.weatherPort.Receive(&pb.AskAboutWeatherResponse{
		Result: "15 Celsius Degrees",
	})
}

func (st *SuiteTest) TestFarenheitConvErr(t *testing.T) {
	st.weatherPort.Send(&pb.AskAboutWeatherRequest{
		City:  "Wroclaw",
		Scale: pb.Scale_FAHRENHEIT,
	})
	st.httpPort.Receive(&port.HTTPRequest{
		Method: "GET",
		Host:   "api.weather.com",
		URL:    "/",
	})
	st.httpPort.Send(&port.HTTPResponse{
		Status: 200,
		Body:   []byte("15 Celsius Degrees"),
	})
	st.weatherPort.Receive(match.GRPCErr(codes.Internal, "failed to convert http response to int"))
}

func (st *SuiteTest) TestFarenheit(t *testing.T) {
	st.weatherPort.Send(&pb.AskAboutWeatherRequest{
		City:  "Wroclaw",
		Scale: pb.Scale_FAHRENHEIT,
	})
	st.httpPort.Receive(&port.HTTPRequest{
		Method: "GET",
		Host:   "api.weather.com",
		URL:    "/",
	})
	st.httpPort.Send(&port.HTTPResponse{
		Status: 200,
		Body:   []byte("15"),
	})
	st.convPort.Receive(&pb.CelciusToFarenheitRequest{
		Value: 15,
	})
	st.convPort.Send(&pb.CelciusToFarenheitResponse{
		Value: 59,
	})
	st.weatherPort.Receive(&pb.AskAboutWeatherResponse{
		Result: "59 Farenheit Degrees",
	})
}

func (st *SuiteTest) TestFarenheitConvGRPCErr(t *testing.T) {
	st.weatherPort.Send(&pb.AskAboutWeatherRequest{
		City:  "Wroclaw",
		Scale: pb.Scale_FAHRENHEIT,
	})
	st.httpPort.Receive(&port.HTTPRequest{
		Method: "GET",
		Host:   "api.weather.com",
		URL:    "/",
	})
	st.httpPort.Send(&port.HTTPResponse{
		Status: 200,
		Body:   []byte("15"),
	})
	st.convPort.Receive(&pb.CelciusToFarenheitRequest{
		Value: 15,
	})
	st.convPort.Send(&port.GRPCErr{
		Err: status.Error(codes.FailedPrecondition, "not supported"),
	})
	st.weatherPort.Receive(match.GRPCErr(codes.Internal, "cale conv client failed"))
}
