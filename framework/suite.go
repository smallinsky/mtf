package framework

import (
	"testing"
)

func NewSuite(testID string, m *testing.M) *Suite {
	return newSuite(testID, m.Run)
}

type Suite struct {
	testID string
	mRunFn runFn
}

func (s *Suite) Run() {
	// TODO: setup testing env and all dependency in docker
	// and before triggering testcases run rediness check.

	s.mRunFn()

	// TODO: clear all dependency, add leazy teardown for most
	// time consuming components like database conatiner, right now DB start
	// in docker can take around 15s.
}

type runFn func() int

func newSuite(testID string, run runFn) *Suite {
	return &Suite{
		mRunFn: run,
		testID: testID,
	}
}
