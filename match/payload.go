package match

import (
	"fmt"
	"strings"

	"github.com/go-test/deep"
)

type PayloadMatcher struct {
	Exp interface{}
}

func Payload(message interface{}) *PayloadMatcher {
	return &PayloadMatcher{
		Exp: message,
	}
}

func (m *PayloadMatcher) Validate() error {
	return nil
}

func (m *PayloadMatcher) Match(err error, got interface{}) error {
	if err != nil {
		return fmt.Errorf("received unexpected error during %T matcher call, err: %v", m, err)
	}
	if errs := deep.Equal(got, m.Exp); err != nil {
		err := strings.Join(errs, "\n")
		return fmt.Errorf(err)
	}
	return nil
}
