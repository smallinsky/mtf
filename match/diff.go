package match

import (
	"github.com/pkg/errors"

	"github.com/google/go-cmp/cmp"
)

type DiffType struct {
	exp interface{}
}

func Diff(exp interface{}) *DiffType {
	return &DiffType{
		exp: exp,
	}
}

func (m *DiffType) Match(got interface{}) error {
	if str := cmp.Diff(got, m.exp); str != "" {
		return errors.Wrapf(ErrNotEq, "diff: %v", str)
	}
	return nil
}
