package port

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/pkg/errors"

	"github.com/smallinsky/mtf/pkg/fswatch"
	pb "github.com/smallinsky/mtf/proto/fswatch"
)

type FTPPort struct {
	ftpEventC chan *pb.EventRequest
	conn      *ftp.ServerConn
}

func NewFTPPort(addr, user, pass string) (*Port, error) {
	p, err := NewFTP(addr, user, pass)
	if err != nil {
		return nil, err
	}

	return &Port{
		impl: p,
	}, nil
}

func NewFTP(addr, user, pass string) (*FTPPort, error) {
	conn, err := dialFTP("localhost:21", "test", "test")
	if err != nil {
		return nil, fmt.Errorf("faield to dial ftp: %v", err)
	}

	ftpPort := &FTPPort{
		ftpEventC: make(chan *pb.EventRequest),
		conn:      conn,
	}

	go func() {
		fswatch.Subscriber(":4441", func(event *pb.EventRequest) {
			fmt.Println("ftp port got event", event.String())
			ftpPort.ftpEventC <- event
		})
	}()

	return ftpPort, nil
}

type FTPEvent struct {
	Path    string
	Payload []byte
}

func (p *FTPPort) Send(ctx context.Context, i interface{}) error {
	event, ok := i.(*FTPEvent)
	if !ok {
		return fmt.Errorf("FTPPort send supports only FTPEvent type")
	}

	if err := p.conn.Stor(event.Path, bytes.NewBuffer(event.Payload)); err != nil {
		return fmt.Errorf("ftp file upload failed %v", err)
	}

	return nil
}

func (p *FTPPort) Receive(ctx context.Context) (interface{}, error) {
	select {
	case msg := <-p.ftpEventC:
		return &FTPEvent{
			Path:    msg.GetPath(),
			Payload: msg.GetContent(),
		}, nil
	case <-time.NewTimer(time.Second * 6).C:
		return nil, errors.Errorf("fialed to recive message, deadline exeeded")
	}
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
