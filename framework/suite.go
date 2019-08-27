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

	"github.com/docker/docker/api/types"

	"github.com/smallinsky/mtf/framework/components/migrate"
	"github.com/smallinsky/mtf/framework/components/mysql"
	"github.com/smallinsky/mtf/framework/components/network"
	"github.com/smallinsky/mtf/framework/components/pubsub"
	"github.com/smallinsky/mtf/framework/components/redis"
	"github.com/smallinsky/mtf/framework/components/sut"
	"github.com/smallinsky/mtf/framework/context"
	"github.com/smallinsky/mtf/pkg/docker"
)

func NewSuite(testID string, m *testing.M) *Suite {
	return newSuite(testID, m.Run)
}

type Suite struct {
	testID      string
	mRunFn      runFn
	sutEnv      []string
	migratePath string
	sutPath     string
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
	stopFn, err := s.startComponents()
	if err != nil {
		fmt.Println("failed to run components: ", err)
		return
	}
	fmt.Printf("=== PREPERING TEST ENV DONE - %v\n\n", time.Now().Sub(start))

	defer stopFn()

	start = time.Now()
	s.mRunFn()
	fmt.Printf("=== TEST RUN DONE - %v\n", time.Now().Sub(start))
}
func (s *Suite) SUTEnv(m map[string]string) *Suite {
	for k, v := range m {
		s.sutEnv = append(s.sutEnv, fmt.Sprintf("%s=%s", k, v))
	}
	return s
}

func (s *Suite) SetMigratePath(path string) *Suite {
	s.migratePath = path
	return s
}

func (s *Suite) SetSUTPath(path string) *Suite {
	s.sutPath = path
	return s
}

type Attachable interface {
	StartOrAttachIfAlreadyExits([]types.Container)
}

func (s *Suite) startComponents() (stopFn func(), err error) {
	cli, err := docker.NewClient()
	if err != nil {
		return nil, err
	}

	netCom := network.New(cli, network.NetworkConfig{
		Name:          "mtf_net",
		AttachIfExist: true,
	})

	if err != nil {
		return nil, err
	}
	pub := pubsub.NewPubsub(cli)
	s.sutEnv = append(s.sutEnv, "PUBSUB_EMULATOR_HOST=host.docker.internal:8085")

	sutCom, err := sut.NewSUT(cli, sut.SutConfig{
		Path: s.sutPath,
		Env:  s.sutEnv,
	})
	if err != nil {
		return nil, err
	}

	dbConfig := mysql.MySQLConfig{
		Database:      "test_db",
		Password:      "test",
		AttachIfExist: true,
	}

	mysqlCom := mysql.NewMySQL(cli, dbConfig)

	redisCom := redis.NewRedis(cli, redis.RedisConfig{
		Password: "test",
	})

	migrate := migrate.NewMigrate(cli, migrate.MigrateConfig{
		Path:     s.migratePath,
		Password: dbConfig.Password,
		Port:     "3306",
		Hostname: "mysql_mtf",
		Database: dbConfig.Database,
	})
	comps := []Comper{
		netCom,
		sutCom,
		pub,
	}

	if false {
		comps = []Comper{
			netCom,
			sutCom,
			mysqlCom,
			redisCom,
			migrate,
			pub,
		}
	}

	m := make(map[int][]Comper)
	for _, v := range comps {
		m[v.StartPriority()] = append(m[v.StartPriority()], v)
	}

	stopFn = func() {
		start := time.Now()
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
		fmt.Printf("=== STOPING ENV DONE - %v\n", time.Since(start))
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

	return stopFn, nil
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
