package core

import (
	"fmt"
	"testing"
)

type result struct {
	i interface{}
}

type typeAReq struct {
	ID string
}

type typeAResp struct {
	ID string
}

type typeBReq struct {
	ID string
}

type typeBResp struct {
	ID string
}

func process(i interface{}) interface{} {
	switch t := i.(type) {
	case *typeAReq:
		return &typeAResp{
			ID: t.ID,
		}
	case *typeBReq:
		return &typeBResp{
			ID: t.ID,
		}
	default:
		panic("unexpected type")
	}
}

func TestCoreQueue(t *testing.T) {
	p := newQueue(process)

	for i := 0; i < 100; i++ {
		rcv := p.Send(&typeAReq{ID: fmt.Sprintf("%v", i)})

		rcv.Receive(&typeAResp{ID: fmt.Sprintf("%v", i)})
	}
}
