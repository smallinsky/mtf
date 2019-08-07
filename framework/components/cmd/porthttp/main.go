package main

import (
	"time"

	"github.com/smallinsky/mtf/port"
)

func main() {
	p, _ := port.NewHTTP()
	p = p
	time.Sleep(time.Hour)
}
