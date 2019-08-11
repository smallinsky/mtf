package framework

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/smallinsky/mtf/framework/components"
	"github.com/smallinsky/mtf/framework/components/mysql"
	"github.com/smallinsky/mtf/framework/components/redis"
	"github.com/smallinsky/mtf/framework/components/sut"
	"github.com/smallinsky/mtf/framework/context"
	"github.com/smallinsky/mtf/framework/core"
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
		mysql.NewMySQL(),
		redis.NewRedis(),
	}

	for _, comp := range comps {
		//	go func(comp Comper) {
		if err := comp.Start(); err != nil {
			log.Fatalf("failed to start %T, err %v", comp, err)
		}
		fmt.Printf("started %T \n", comp)
		//	}(comp)
	}

	for _, comp := range comps {
		if err := comp.Ready(); err != nil {
			log.Fatalf("faield to call ready for %T, err: %v", comp, err)
		}
	}

	aa := sut.NewSUT(
		"/Users/Marek/Go/src/github.com/smallinsky/mtf/e2e/service/echo/",
		"ORACLE_ADDR=host.docker.internal:8002",
	)

	err := aa.Start()
	if err != nil {
		log.Fatalf("failed to run sut: %v", err)
	}
	aa.Ready()

	defer func() {
		aa.Stop()
	}()

	fmt.Printf("Components start time: %v\n", time.Now().Sub(start))
	s.mRunFn()

	// TODO: clear all dependency, add leazy teardown for most
	// time consuming components like database conatiner, right now DB start
	// in docker can take around 15s.

	if core.Settings.StopComponentsAfterExit {
		// reverse order
		for i := len(comps) - 1; i >= 0; i-- {
			// TODO defer during component start.
			comp := comps[i]
			if err := comp.Stop(); err != nil {
				log.Fatalf("faild to stop %T, err: %v", comp, err)
			}
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

	context.CreateDirectory()

	fmt.Println("runing test cases")
	for _, test := range getTests(i) {
		// Create context and tmp dir
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
		if _, ok := m.Interface().(func(*testing.T)); !ok {
			continue
		}
		out = append(out, testing.InternalTest{
			Name: tm.Name,
			F: func(t *testing.T) {
				// create test dir
				context.CreateTestContext(t)
				m.Call([]reflect.Value{reflect.ValueOf(t)})
				context.RemoveTextContext(t)
				// get all port and run cleanup func
			},
		})
	}
	return out
}
