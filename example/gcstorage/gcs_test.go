// +build mtf

package gcstorage

import (
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
	var err error
	st.gcsPort, err = port.NewGCSPort()
	if err != nil {
		t.Fatal("unexpected error ", err)
	}
}

type SuiteTest struct {
	gcsPort *port.Port
}

func (st *SuiteTest) TestGCStorage(t *testing.T) {
	st.gcsPort.Receive(t, &port.StorageGetRequest{
		Bucket: "bucket/path",
		Object: "file.txt",
	})

	st.gcsPort.Send(t, &port.StorageGetResponse{
		Content: []byte("awesomefile.txt file content"),
	})

	st.gcsPort.Receive(t, &port.StorageInsertRequest{
		Bucket:  "bucket/path/bak",
		Object:  "file.txt.bak",
		Content: []byte("awesomefile.txt file content"),
	})

	st.gcsPort.Send(t, &port.StorageInsertResponse{})
}
