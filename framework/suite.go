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
	"github.com/smallinsky/mtf/pkg/cert"
	"github.com/smallinsky/mtf/pkg/docker"
)

func NewSuite(m *testing.M) *Suite {
	return newSuite(m.Run)
}

type Suite struct {
	mRunFn      runFn
	sutEnv      []string
	migratePath string
	sutPath     string
	settings    Settings
}

func (s *Suite) WithMySQL(c MysqlSettings) *Suite {
	s.settings.mysql = &c
	return s
}

func (s *Suite) WithSut(c SutSettings) *Suite {
	s.settings.sut = &c
	return s
}

func (s *Suite) WithPubSub(c PubSubSettings) *Suite {
	s.settings.pubsub = &c
	return s
}

func (s *Suite) WithRedis(c RedisSettings) *Suite {
	s.settings.redis = &c
	return s
}

type Settings struct {
	mysql  *MysqlSettings
	sut    *SutSettings
	pubsub *PubSubSettings
	redis  *RedisSettings
}

type MysqlSettings struct {
	DatabaseName string
	MigrationDir string
	Password     string
	Port         string
}

type SutSettings struct {
	Envs []string
	Dir  string
}

type PubSubSettings struct {
}

type RedisSettings struct {
	Port     string
	Password string
}

type Comper interface {
	Start() error
	Stop() error
	Ready() error
	StartPriority() int
}

func (s *Suite) Run() {

	kvpair, err := cert.GenCert([]string{"localhost", "host.docker.internal"})
	if err != nil {
		log.Fatalf("[ERR] failed to generate certs")
	}
	cert.WriteCert(kvpair)
	components := s.getComponents()

	start := time.Now()
	fmt.Println("=== PREPERING TEST ENV")
	components.Start()
	start = time.Now()
	fmt.Printf("=== PREPERING TEST ENV DONE - %v\n\n", time.Now().Sub(start))
	s.mRunFn()
	fmt.Printf("=== TEST RUN DONE - %v\n", time.Now().Sub(start))
	components.Stop()
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

func (s *Suite) getComponents() Components {
	cli, err := docker.NewClient()
	if err != nil {
		log.Fatalf("faield to get docker client: %v", err)
	}

	components := Components{
		m: make(map[startPriority][]Comper),
	}

	components.Add(network.New(cli, network.NetworkConfig{
		Name:          "mtf_net",
		AttachIfExist: true,
	}))

	if s.settings.redis != nil {
		components.Add(redis.NewRedis(cli, redis.RedisConfig{
			Password: s.settings.redis.Password,
		}))
	}

	if s.settings.pubsub != nil {
		components.Add(pubsub.NewPubsub(cli))
	}

	if s.settings.sut != nil {
		s.settings.sut.Envs = append(s.settings.sut.Envs, "PUBSUB_EMULATOR_HOST=host.docker.internal:8085")
		components.Add(sut.NewSUT(cli, sut.SutConfig{
			Path: s.settings.sut.Dir,
			Env:  s.settings.sut.Envs,
		}))
	}

	if cfg := s.settings.mysql; cfg != nil {
		components.Add(mysql.NewMySQL(cli, mysql.MySQLConfig{
			Database:      cfg.DatabaseName,
			Password:      cfg.Password,
			AttachIfExist: true,
		}))
	}

	if cfg := s.settings.mysql; cfg != nil && cfg.MigrationDir != "" {
		components.Add(migrate.NewMigrate(cli, migrate.MigrateConfig{
			Path:     cfg.MigrationDir,
			Password: cfg.Password,
			Port:     "3306",
			Hostname: "mysql_mtf",
			Database: cfg.DatabaseName,
		}))
	}

	return components
}

type startPriority int

type Components struct {
	m       map[startPriority][]Comper
	maxPrio int
}

func (c *Components) Add(v Comper) {
	if v.StartPriority() > c.maxPrio {
		c.maxPrio = v.StartPriority()
	}
	c.m[startPriority(v.StartPriority())] = append(c.m[startPriority(v.StartPriority())], v)

}

func (c *Components) Start() {
	for i := 0; i < c.maxPrio+1; i++ {
		cc, ok := c.m[startPriority(i)]
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
}

func (co *Components) Stop() {
	for i := co.maxPrio; i >= 0; i-- {
		cc, ok := co.m[startPriority(i)]
		if !ok {
			continue
		}

		var wg sync.WaitGroup
		for _, c := range cc {
			if c == nil {
				continue
			}
			sutC, ok := c.(*sut.SUT)
			if ok {
				buff, err := sutC.Logs()
				if err != nil {
					fmt.Println("got error during sut logs call")
					return
				}

				err = ioutil.WriteFile(fmt.Sprintf("%s/sut.logs", "runlogs"), buff, 0644)
				if err != nil {
					fmt.Println("failed to write sut logs: ", err)
					return
				}
			}
			wg.Add(1)
			go func(cp Comper) {
				defer wg.Done()
				if err := cp.Stop(); err != nil {
					fmt.Printf("failed to stop %T: %v", cp, err)
				}
			}(c)
		}
		wg.Wait()
	}
}

type runFn func() int

func newSuite(run runFn) *Suite {
	return &Suite{
		mRunFn: run,
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
