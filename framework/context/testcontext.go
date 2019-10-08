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

var contextMap = map[string]*TestContext{}
var mtx sync.Mutex

func Get(t *testing.T) *TestContext {
	mtx.Lock()
	defer mtx.Unlock()
	return contextMap[getTestPrefix(t)]
}

func getTestPrefix(t *testing.T) string {
	return strings.Split(t.Name(), "/")[0]
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

	name := strings.Replace(t.Name(), "/", "_", -1)
	var err error
	c.file, err = os.Create(fmt.Sprintf("%s/%s.log", logDir, name))
	if err != nil {
		panic(err)
	}
	c.log = log.New(c.file, "", log.Lmicroseconds)

	mtx.Lock()
	defer mtx.Unlock()
	contextMap[getTestPrefix(t)] = c
}

func RemoveTextContext(t *testing.T) {
	mtx.Lock()
	defer mtx.Unlock()
	delete(contextMap, getTestPrefix(t))
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
	buff, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return "", err
	}

	return string(buff), nil
}
