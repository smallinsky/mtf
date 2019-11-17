package main

import (
	"flag"
	"fmt"

	"github.com/smallinsky/mtf/pkg/fswatch"
)

var (
	addr = flag.String("addr", "host.docker.internal:4441", "watcher address")
	dir  = flag.String("dir", ".", "dir path to watch")
)

func main() {
	flag.Parse()
	fmt.Println("[INFO] watcher remote addr: ", *addr)
	fswatch.Monitor(*addr, *dir)
}
