package framework

import (
	"flag"
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/smallinsky/mtf/framework/context"
	"github.com/smallinsky/mtf/pkg/cert"
	"github.com/smallinsky/mtf/pkg/docker"
)

var GetDockerHostAddr = docker.HostAddr

func NewSuite(m *testing.M) *Suite {
	flag.Parse()
	s := newSuite(m.Run)
	return s
}

type Suite struct {
	mRunFn   runFn
	settings Settings
}

func (s *Suite) Run() {
	fmt.Println("=== PREPERING TEST ENV")
	start := time.Now()

	kvpair, err := cert.GenCert([]string{"localhost", "host.docker.internal"})
	if err != nil {
		log.Fatalf("[ERR] failed to generate certs: %v", err)
	}

	if err := cert.WriteCert(kvpair); err != nil {
		log.Fatalf("[ERR] failed to write certs: %v ", err)
	}

	cli, err := docker.New()
	if err != nil {
		panic(err)
	}
	network, err := cli.CreateNetwork("mtf_net")
	if err != nil {
		log.Fatalf("faield to get docker client: %v", err)
	}
	defer network.Remove()

	containersConfig, err := s.GetContainersConfig()
	if err != nil {
		log.Fatalf("faield to get containers settings: %v", err)
	}

	var containers []*docker.ContainerType
	for _, conf := range containersConfig {
		container, err := cli.NewContainer(*conf)
		if err != nil {
			log.Fatalf("[ERR] failed to run container: %v", err)
		}
		containers = append(containers, container)
	}

	for _, container := range containers {
		start := time.Now()
		fmt.Printf("    Starting %s ", container.Name())
		err := container.Start()
		if err != nil {
			log.Fatalf("\nstart err: %v", err)
		}
		fmt.Printf("-  %v\n", time.Now().Sub(start))
	}
	fmt.Printf("=== TEST RUN DONE - %v\n", time.Now().Sub(start))
	s.mRunFn()

	for _, container := range containers {
		err := container.Stop()
		if err != nil {
			log.Fatalf("stop err: %v", err)
		}
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
