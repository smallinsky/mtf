package framework

import (
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/client"

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
	sutEnv []string
}

type Comper interface {
	Start() error
	Stop() error
	Ready() error
	StartPriority() int
}

func (s *Suite) Run() {
	start := time.Now()

	fmt.Println("=== PREPERING TEST ENV")
	stopFn := s.startComponents()
	fmt.Printf("=== PREPERING TEST ENV DONE - %v\n\n", time.Now().Sub(start))

	defer stopFn()

	s.mRunFn()
}
func (s *Suite) SUTEnv(m map[string]string) *Suite {
	for k, v := range m {
		s.sutEnv = append(s.sutEnv, fmt.Sprintf("%s=%s", k, v))
	}
	return s
}

func (s *Suite) startComponents() (stopFn func()) {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	netCom := network.New(cli, network.NetworkConfig{
		Name: "mtf_net",
	})

	fmt.Println(s.sutEnv)
	sutCom := sut.NewSUT(cli, sut.SutConfig{
		Path: "/Users/marek/Go/src/github.com/smallinsky/mtf/e2e/service/echo/",
		Env:  s.sutEnv,
	})

	dbConfig := mysql.MySQLConfig{
		Database: "test_db",
		Password: "test",
	}

	mysqlCom := mysql.NewMySQL(cli, dbConfig)
	redisCom := redis.NewRedis(cli, redis.RedisConfig{
		Password: "test",
	})
	migrate := migrate.NewMigrate(cli, migrate.MigrateConfig{
		Path:     "../../e2e/migrations",
		Password: dbConfig.Password,
		Port:     "3306",
		Hostname: "mysql_mtf",
		Database: dbConfig.Database,
	})

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

	stopFn = func() {
		for i := 9; i >= 0; i-- {
			cc, ok := m[i]
			if !ok {
				continue
			}
			var wg sync.WaitGroup
			for _, c := range cc {
				sutC, ok := c.(*sut.SUT)
				if ok {
					buff, err := sutC.Logs()
					if err != nil {
						fmt.Println("got error during sut logs call")
					}

					err = ioutil.WriteFile(fmt.Sprintf("%s/sut.logs", "runlogs"), buff, 0644)
					if err != nil {
						fmt.Println("failed to write sut logs: ", err)
					}
				}
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
	}

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

	return stopFn
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

	for _, test := range getInternalTests(i) {
		t.Run(test.Name, test.F)
	}
}

func getInternalTests(i interface{}) []testing.InternalTest {
	var tests []testing.InternalTest
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
		tests = append(tests, testing.InternalTest{
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
	return tests
}
