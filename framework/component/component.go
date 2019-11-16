package component

import (
	"context"
	"io"
)

type Component interface {
	Start() error
	Stop() error
}

type Loggable interface {
	Logs(context.Context) (io.Reader, error)
	Name() string
}
