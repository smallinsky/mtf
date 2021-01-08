// +build mtf

package ftptest

import (
	"testing"

	"github.com/smallinsky/mtf/framework"
	"github.com/smallinsky/mtf/port"
)

func TestMain(m *testing.M) {
	framework.TestEnv(m).
		WithFTP(framework.FTPSettings{
			Addr: "mtf_ftp:21",
			User: "test",
			Pass: "test",
		}).
		WithSUT(framework.SutSettings{
			Dir:         "./service",
			RuntimeType: framework.RuntimeTypeCommand,
		}).
		Run()
}

func TestEchoService(t *testing.T) {
)
	framework.Run(t, new(SuiteTest))
}

func (st *SuiteTest) Init(t *testing.T) {
	var err error
	if st.ftpPort, err = port.NewFTPPort("", "", ""); err != nil {
		t.Fatalf("failed to init ftp port: %v", err)
	}
}

type SuiteTest struct {
	ftpPort *port.Port
}

func (s *SuiteTest) TestFTPUpload(t *testing.T) {
	s.ftpPort.Receive(t, &port.FTPEvent{
		Path:    "/ftp/randomfile.txt",
		Payload: []byte("random file content"),
	})
}
