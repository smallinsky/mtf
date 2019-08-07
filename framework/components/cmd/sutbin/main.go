package main

import (
	"github.com/smallinsky/mtf/framework/components"
)

func main() {
	err := components.BuildGoBinary("../e2e/service/echo")
	if err != nil {
		panic(err)
	}
}
