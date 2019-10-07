package main

import (
	"flag"

	"github.com/smallinsky/mtf/pkg/fswatch"
)

var (
	addr = flag.String("addr", "host.docker.internal:4441", "watcher address")

	dir = flag.String("dir", ".", "dir path to watch")
)

func main() {
	flag.Parse()
	fswatch.Monitor(*addr, *dir)
}
