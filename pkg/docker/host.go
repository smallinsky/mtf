package docker

import (
	"fmt"
	"log"
)

var hostAddr string

func init() {
	host, err := HostIP()
	if err != nil {
		log.Fatalf("[ERROR] Failed to init host ip: %v", err)
	}
	hostAddr = host
}

func HostAddr(port int) string {
	return fmt.Sprintf("%s:%d", hostAddr, port)
}
