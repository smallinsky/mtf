package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/pkg/errors"
)

func main() {
	conn, err := dialFTP("ftp_mtf:21", "test", "test")
	if err != nil {
		log.Fatalf("failed to connect to ftp server")
	}

	time.Sleep(time.Second * 1)
	err = conn.Stor("randomfile.txt", strings.NewReader("random file content"))
	if err != nil {
		log.Fatalf("failed to upload file: %v", err)
	}

	conn = conn
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
