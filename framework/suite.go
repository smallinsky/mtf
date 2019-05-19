package framework

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"
	"time"

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
	Start() error
	Stop() error
	Ready() error
}

func (s *Suite) Run() {
	// TODO: setup testing env and all dependency in docker
	// and before triggering testcases run rediness check.
	start := time.Now()

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
			if err := comp.Start(); err != nil {
				log.Fatalf("failed to start %T, err %v", comp, err)
			}
		}(comp)
	}

	for _, comp := range comps {
		if err := comp.Ready(); err != nil {
			log.Fatalf("faield to call ready for %T, err: %v", comp, err)
		}
	}

	sut := components.NewSUT(
		"/Users/Marek/Go/src/github.com/smallinsky/mtf/e2e/service/echo/",
		"ORACLE_ADDR=host.docker.internal:8002",
	)

	sut.Start()
	sut.Ready()

	defer func() {
		sut.Stop()
	}()

	fmt.Printf("Components start time: %v\n", time.Now().Sub(start))
	s.mRunFn()

	// TODO: clear all dependency, add leazy teardown for most
	// time consuming components like database conatiner, right now DB start
	// in docker can take around 15s.

	// reverse order
	for i := len(comps) - 1; i >= 0; i-- {
		// TODO defer during component start.
		comp := comps[i]
		if err := comp.Stop(); err != nil {
			log.Fatalf("faild to stop %T, err: %v", comp, err)

		}
	}
}

type runFn func() int

func newSuite(testID string, run runFn) *Suite {
	return &Suite{
		mRunFn: run,
		testID: testID,
	}
}

func Run(t *testing.T, i interface{}) {
	if v, ok := i.(interface{ Init(*testing.T) }); ok {
		v.Init(t)
	}

	for _, test := range getTests(i) {
		t.Run(test.Name, test.F)
	}
}

func getTests(i interface{}) []testing.InternalTest {
	var out []testing.InternalTest
	v := reflect.ValueOf(i)
	if v.Type().Kind() != reflect.Ptr && v.Type().Kind() != reflect.Struct {
		panic("arg is not a ptr to a struct")
	}
	for i := 0; i < v.Type().NumMethod(); i++ {
		tm := v.Type().Method(i)
		if !strings.HasPrefix(tm.Name, "Test") {
			continue
		}
		m := v.Method(i)
		if m.Type().NumIn() != 1 {
			continue
		}
		if m.Type().NumOut() != 0 {
			continue
		}
		if m.Type().In(0) != reflect.TypeOf((*testing.T)(nil)) {
			continue
		}
		out = append(out, testing.InternalTest{
			Name: tm.Name,
			F: func(t *testing.T) {
				m.Call([]reflect.Value{reflect.ValueOf(t)})
			},
		})
	}
	return out
}
