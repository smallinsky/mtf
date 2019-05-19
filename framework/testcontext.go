package framework

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"
	"testing"
)

var contextMap = map[*testing.T]*TestContext{}

func GetContextFor(t *testing.T) *TestContext {
	return contextMap[t]
}

type TestContext struct {
	file *os.File
	mtx  sync.Mutex
	log  *log.Logger
}

func CreateDirectory() error {
	err := os.Mkdir("./runlogs", os.ModePerm)
	if err != nil {
		panic(err)
	}
	return nil
}

func (c *TestContext) Clear() {
	c.file.Close()
}

func createTestContext(t *testing.T) {
	c := &TestContext{}

	name := strings.ReplaceAll(t.Name(), "/", "_")
	var err error
	c.file, err = os.Create(name)
	if err != nil {
		panic(err)
	}
	c.log = log.New(c.file, "", log.Lmicroseconds)
	contextMap[t] = c
}

func removeTextContext(t *testing.T) {
	delete(contextMap, t)
}

func (c *TestContext) LogReceive(name, string, i interface{}) {
	payload, _ := dump(i)
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.log.Printf("%s <- Received:\n%s", name, payload)

}

func (c *TestContext) LogSend(name string, i interface{}) {
	payload, _ := dump(i)
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.log.Printf("%s <- Send:\n%s", name, payload)
}

func dump(i interface{}) (string, error) {
	buff, err := json.MarshalIndent(i, " ", " ")
	if err != nil {
		return "", err
	}

	return string(buff), nil
}
