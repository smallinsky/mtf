package framework

import (
	"reflect"
	"strings"
	"testing"

	"github.com/smallinsky/mtf/framework/context"
)

type Initable interface {
	Init(*testing.T)
}

func Run(t *testing.T, i interface{}) {
	if v, ok := i.(Initable); ok {
		v.Init(t)
	}
	context.CreateDirectory()

	for _, test := range getInternalTests(i) {
		if testenv.settings.SUT.RuntimeType == RuntimeTypeCommand {
			err := testenv.StartSutInCommandMode()
			if err != nil {
				t.Fatalf("[MTF ERROR] Failed to start sut component in cmd mode: %v", err)
			}
		}
		t.Run(test.Name, test.F)

		if testenv.settings.SUT.RuntimeType == RuntimeTypeCommand {
			err := testenv.StopSutInCommandMode(test.Name)
			if err != nil {
				t.Fatalf("[MTF ERROR] Failed to stop sut component in cmd mode: %v", err)
			}
		}
	}
}

func getInternalTests(i interface{}) []testing.InternalTest {
	var tests []testing.InternalTest
	v := reflect.ValueOf(i)
	if v.Type().Kind() != reflect.Ptr && v.Type().Kind() != reflect.Struct {
		panic("invalid argument, expect ptr to struct")
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
				context.CreateTestContext(t)
				m.Call([]reflect.Value{reflect.ValueOf(t)})
				context.RemoveTextContext(t)
			},
		})
	}

	return tests
}
