package main

import (
	"flag"

	"github.com/smallinsky/mtf/pkg/fswatch"
)

var (
	addr = flag.String("addr", "localhost:44441", "watcher address")
	dir  = flag.String("dir", ".", "dir path to watch")
)

func main() {
	flag.Parse()
	fswatch.Monitor(*addr, *dir)
}
