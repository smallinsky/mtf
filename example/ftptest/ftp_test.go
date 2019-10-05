package ftptest

import (
	"fmt"
	"testing"

	"github.com/jlaffaye/ftp"
	"github.com/pkg/errors"

	"github.com/smallinsky/mtf/framework"
	"github.com/smallinsky/mtf/port"
)

func TestMain(m *testing.M) {
	framework.NewSuite(m).
		WithFTP(framework.FTPSettings{
			Addr: "mtf_ftp:21",
			User: "test",
			Pass: "test",
		}).
		WithSut(framework.SutSettings{
			Dir: "./service",
		}).
		Run()
}

func TestEchoService(t *testing.T) {
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
	conn, err := dialFTP("localhost:21", "test", "test")
	if err != nil {
		t.Fatalf("failed to dial ftp server: %v", err)
	}
	conn = conn
	s.ftpPort.Receive(t, &port.FTPEvent{
		Path:    "/ftp/randomfile.txt",
		Payload: []byte("cmFuZG9tIGZpbGUgY29udGVudA=="),
	})
}

func dialFTP(addr string, user, pass string) (*ftp.ServerConn, error) {
	connection, err := ftp.Connect(addr)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to %q", addr)
	}
	if err := connection.Login(user, pass); err != nil {
		return nil, fmt.Errorf("failed to login to %q: %v", addr, err)
	}
	return connection, nil
}
