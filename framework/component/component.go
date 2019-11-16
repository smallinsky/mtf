package component

import (
	"context"
	"io"
)

type Component interface {
	Start(context.Context) error
	Stop(context.Context) error
}

type Loggable interface {
	Logs(context.Context) (io.Reader, error)
	Name() string
}
