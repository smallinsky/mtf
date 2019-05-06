package framework

import (
	"testing"

	"github.com/smallinsky/mtf/framework/components"
)

func TestMain(m *testing.M) {
	NewSuite("suite_first", m).Run()
}

func TestFoo(t *testing.T) {
	mysql := components.MySQL{}
	mysql.Start()
	defer mysql.Stop()
}

func TestBar(t *testing.T) {
}
