package component

import (
	"context"
	"io"
)

// Component is a interface that allows to start
// and stop docker component.
type Component interface {
	// Start start docker component.
	Start(context.Context) error
	// Stop stops docker components.
	Stop(context.Context) error
}

// Loggable allows bo obtains logs from docker container.
type Loggable interface {
	// Logs returns reader for buffer that contains
	// preformated container logs.
	Logs(context.Context) (io.Reader, error)
	// Name returns container name.
	Name() string
}
