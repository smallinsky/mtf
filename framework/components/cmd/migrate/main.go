package main

import (
	"github.com/smallinsky/mtf/framework/components"
)

func main() {
	migrate := &components.MigrateDB{}
	migrate.Start()
}
