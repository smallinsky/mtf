// +build mtf

package framework

import (
	"net/http"
	"testing"

	"github.com/smallinsky/mtf/framework"
	"github.com/smallinsky/mtf/port"
)

func TestMain(m *testing.M) {
	framework.TestEnv(m).WithSUT(framework.SutSettings{
		Dir:         "./service",
		RuntimeType: framework.RuntimeTypeCommand,
	}).Run()
}

func TestEchoService(t *testing.T) {
	framework.Run(t, new(SuiteTest))
}

func (st *SuiteTest) Init(t *testing.T) {
	st.httpPort = port.NewHTTPPort()
}

type SuiteTest struct {
	httpPort *port.Port
}

func (st *SuiteTest) TestHTTP(t *testing.T) {
	st.httpPort.Receive(t, &port.HTTPRequest{
		Method: "GET",
		Host:   "example.com",
		URL:    "/urlpath",
	})

	st.httpPort.Send(t, &port.HTTPResponse{
		Body:   []byte(`{"value":{"joke":"42"}}`),
		Status: http.StatusOK,
	})
}
