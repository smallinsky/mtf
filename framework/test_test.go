package framework

import (
	"testing"
)

func TestMain(m *testing.M) {
	NewSuite("suite_first", m).Run()
}

func TestFoo(t *testing.T) {
}

func TestBar(t *testing.T) {
}
