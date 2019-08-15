package framework

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/smallinsky/mtf/framework/components/migrate"
	"github.com/smallinsky/mtf/framework/components/mysql"
	"github.com/smallinsky/mtf/framework/components/network"
	"github.com/smallinsky/mtf/framework/components/redis"
	"github.com/smallinsky/mtf/framework/components/sut"
	"github.com/smallinsky/mtf/framework/context"
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
	StartPriority() int
}

func (s *Suite) Run() {
	fmt.Println("=== PREPERING TEST ENV")
	start := time.Now()

	netCom := network.New()

	sutCom := sut.NewSUT(
		"/Users/Marek/Go/src/github.com/smallinsky/mtf/e2e/service/echo/",
		"ORACLE_ADDR=host.docker.internal:8002",
	)
	mysqlCom := mysql.NewMySQL()
	redisCom := redis.NewRedis()
	migrate := &migrate.MigrateDB{}

	comps := []Comper{
		netCom,
		sutCom,
		mysqlCom,
		redisCom,
		migrate,
	}

	m := make(map[int][]Comper)
	for _, v := range comps {
		m[v.StartPriority()] = append(m[v.StartPriority()], v)
	}

	defer func() {
		for i := 9; i >= 0; i-- {
			cc, ok := m[i]
			if !ok {
				continue
			}
			var wg sync.WaitGroup
			for _, c := range cc {
				wg.Add(1)
				go func(c Comper) {
					defer wg.Done()
					if err := c.Stop(); err != nil {
						fmt.Printf("failed to stop %T: %v", c, err)
					}
				}(c)
			}
			wg.Wait()
		}
	}()

	for i := 0; i < 10; i++ {
		cc, ok := m[i]
		if !ok {
			continue
		}
		var wg sync.WaitGroup
		for _, c := range cc {
			wg.Add(1)
			go func(c Comper) {
				defer wg.Done()
				fmt.Printf("%T is starting \n", c)
				start := time.Now()
				if err := c.Start(); err != nil {
					log.Fatalf("failed to start %T, err %v\n", c, err)
				}
				if err := c.Ready(); err != nil {
					log.Fatalf("faield to call ready for %T, err: %v", c, err)
				}
				fmt.Printf("%T is ready - %v\n", c, time.Since(start))
			}(c)
		}
		wg.Wait()
	}
	fmt.Printf("=== PREPERING TEST ENV DONE - %v\n\n", time.Now().Sub(start))
	s.mRunFn()
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
