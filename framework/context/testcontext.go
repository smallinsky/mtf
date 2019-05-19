package context

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"testing"
)

var contextMap = map[*testing.T]*TestContext{}

func Get(t *testing.T) *TestContext {
	return contextMap[t]
}

type TestContext struct {
	file *os.File
	mtx  sync.Mutex
	log  *log.Logger
}

const (
	logDir = "runlogs"
)

func CreateDirectory() error {
	if err := os.RemoveAll(logDir); err != nil {
		panic(err)
	}
	if err := os.Mkdir(logDir, os.ModePerm); err != nil {
		panic(err)
	}
	return nil
}

func (c *TestContext) Clear() {
	c.file.Close()
}

func CreateTestContext(t *testing.T) {
	c := &TestContext{}

	name := strings.ReplaceAll(t.Name(), "/", "_")
	var err error
	c.file, err = os.Create(fmt.Sprintf("%s/%s.log", logDir, name))
	if err != nil {
		panic(err)
	}
	c.log = log.New(c.file, "", log.Lmicroseconds)
	contextMap[t] = c
}

func RemoveTextContext(t *testing.T) {
	delete(contextMap, t)
}

func (c *TestContext) LogReceive(name string, i interface{}) {
	payload, _ := dump(i)
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.log.Printf("%s <- %T:\n%s", name, i, payload)

}

func (c *TestContext) LogSend(name string, i interface{}) {
	payload, _ := dump(i)
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.log.Printf("%s -> %T:\n%s", name, i, payload)
}

func dump(i interface{}) (string, error) {
	buff, err := json.MarshalIndent(i, " ", " ")
	if err != nil {
		return "", err
	}

	return string(buff), nil
}
