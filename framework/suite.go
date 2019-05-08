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

	net := components.NewNet()
	net.Start()
	net.Ready()
	defer net.Stop()

	comps := []Comper{
		components.NewMySQL(),
		components.NewRedis(),
		//	components.NewPubsub(),
	}

	for _, comp := range comps {
		go func(comp Comper) {
			log.Printf("--- Staring %T component ---\n", comp)
			comp.Start()
		}(comp)
		log.Printf("--- Component %T is ready ---\n", comp)
	}

	for _, comp := range comps {
		comp.Ready()
		log.Printf("--- Component %T is ready ---\n", comp)
	}

	m := components.MigrateDB{}
	m.Start()
	m.Ready()
	defer m.Stop()

	sut := components.SUT{}

	log.Printf("starting sut")
	sut.Start()
	sut.Ready()

	defer func() {
		log.Printf("stopping sut")
		sut.Stop()
	}()
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
