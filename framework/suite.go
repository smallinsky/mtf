package framework

import (
	"log"
	"testing"

	"github.com/smallinsky/mtf/framework/components"
)

func NewSuite(testID string, m *testing.M) *Suite {
	return newSuite(testID, m.Run)
}

type Suite struct {
	testID string
	mRunFn runFn
}

type Comper interface {
	Start()
	Stop()
	Ready()
}

func (s *Suite) Run() {
	// TODO: setup testing env and all dependency in docker
	// and before triggering testcases run rediness check.

	comps := []Comper{
		&components.Net{},
		&components.MySQL{},
		&components.Redis{},
		&components.MigrateDB{},
	}

	for _, comp := range comps {
		log.Printf("--- Staring %T component ---\n", comp)
		comp.Start()
		comp.Ready()
		log.Printf("--- Component %T is ready ---\n", comp)
	}
	s.mRunFn()

	// TODO: clear all dependency, add leazy teardown for most
	// time consuming components like database conatiner, right now DB start
	// in docker can take around 15s.

	// reverse order
	for i := len(comps) - 1; i >= 0; i-- {
		// TODO defer during component start.
		comp := comps[i]
		log.Printf("--- Stopping %T component ---\n", comp)
		comp.Stop()
	}
}

type runFn func() int

func newSuite(testID string, run runFn) *Suite {
	return &Suite{
		mRunFn: run,
		testID: testID,
	}
}
