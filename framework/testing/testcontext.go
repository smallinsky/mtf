package testing

import (
	"testing"
)

type Cleaner interface {
	CleanPort()
}

type TestContext interface {
	WorkDir() string
	CreateDirectoryOrFail(name string) string
}

type Context struct {
	T     *testing.T
	Ports []Cleaner
}

type T interface {
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fail()
	FailNow()
	Failed() bool
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Log(args ...interface{})
	Logf(format string, args ...interface{})
	Name() string
	Skip(args ...interface{})
	SkipNow()
	Skipf(format string, args ...interface{})
	Skipped() bool
	Helper()
}

var _ T = &Context{}

func (t *Context) Error(args ...interface{}) {
	t.T.Error(args...)
}

func (t *Context) Errorf(format string, args ...interface{}) {
	t.T.Errorf(format, args...)
}

func (t *Context) Fail() {
	t.T.Fail()
}

func (t *Context) FailNow() {
	t.T.FailNow()
}

func (t *Context) Failed() bool {
	return t.T.Failed()
}

func (t *Context) Fatal(args ...interface{}) {
	t.T.Fatal()
}

func (t *Context) Fatalf(format string, args ...interface{}) {
	t.T.Fatalf(format, args...)
}

func (t *Context) Log(args ...interface{}) {
	t.T.Log(args...)
}

func (t *Context) Logf(format string, args ...interface{}) {
	t.T.Logf(format, args...)
}

func (t *Context) Name() string {
	return t.T.Name()
}

func (t *Context) Skip(args ...interface{}) {
	t.T.Skip()
}

func (t *Context) SkipNow() {
	t.T.SkipNow()
}

func (t *Context) Skipf(format string, args ...interface{}) {
	t.T.Skipf(format, args...)
}

func (t *Context) Skipped() bool {
	return t.T.Skipped()
}

func (t *Context) Helper() {
	t.T.Helper()
}
