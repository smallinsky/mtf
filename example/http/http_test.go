package framework

import (
	"net/http"
	"testing"
	//"time"

	"github.com/smallinsky/mtf/framework"
	"github.com/smallinsky/mtf/port"
)

func TestMain(m *testing.M) {
	framework.NewSuite(m).WithSut(framework.SutSettings{
		Dir: "./service",
	}).Run()
}

func TestEchoService(t *testing.T) {
	framework.Run(t, new(SuiteTest))
}

func (st *SuiteTest) Init(t *testing.T) {
	var err error
	if st.httpPort, err = port.NewHTTPPort(port.WithTLSHost("example.com")); err != nil {
		t.Fatalf("failed to init http port")
	}
	//	time.Sleep(time.Millisecond * 300)
}

type SuiteTest struct {
	httpPort *port.Port
}

func (st *SuiteTest) TestHTTP(t *testing.T) {
	st.httpPort.Receive(t, &port.HTTPRequest{
		Body:   []byte{},
		Method: "GET",
		Host:   "example.com",
		URL:    "/urlpath",
	})
	st.httpPort.Send(t, &port.HTTPResponse{
		Body:   []byte(`{"value":{"joke":"42"}}`),
		Status: http.StatusOK,
	})
}
