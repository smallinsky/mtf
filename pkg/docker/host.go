package docker

import "fmt"

var hostAddr string

func init() {
	host, err := HostIP()
	if err != nil {
		panic(err)
	}
	hostAddr = host
}

func HostAddr(port int) string {
	return fmt.Sprintf("%s:%d", hostAddr, port)
}
