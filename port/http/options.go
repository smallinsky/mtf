package port

import (
	"time"
)

type Opt func(*options)

type options struct {
	err     error
	timeout time.Duration
}

var defaultRcvOptions = options{
	err:     nil,
	timeout: time.Second * 5,
}
