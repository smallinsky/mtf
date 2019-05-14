package core

import (
	"fmt"
	"reflect"
	"testing"
)

func newQueue(processFn func(interface{}) interface{}) *queue {
	return &queue{
		r:  make(chan interface{}),
		fn: processFn,
	}
}

type queue struct {
	r  chan interface{}
	fn func(interface{}) interface{}
}

func (q *queue) send(args ...interface{}) {
	if len(args) == 1 {
		go func() {
			q.r <- q.fn(args[0]) // send message
		}()
	} else {
		panic("wrong number or arguments")
	}
}

type Foo struct {
	err error
	rcv func() interface{}
}

func (f *Foo) Receive(i interface{}) error {
	got := f.rcv()
	if !reflect.DeepEqual(got, i) {
		fmt.Printf("message are not equal \nGot: %+v\nExp: %+v", got, i)
		return nil
	}
	fmt.Printf("got %+v, exp %+v\n", got, i)
	return nil
}

func (q *queue) Send(i interface{}) *Foo {
	c := make(chan interface{})
	go func() {
		r := q.fn(i)
		c <- r
	}()

	return &Foo{
		rcv: func() interface{} {
			return <-c

		},
		err: nil,
	}
}

type sendFnArgsResult struct {
	t *testing.T
}

func (q *queue) receive(i interface{}) error {
	msg := <-q.r
	if !reflect.DeepEqual(i, msg) {
		return fmt.Errorf("message are not equal \nGot: %+v\nExp: %+v", i, msg)
	}
	return nil
}

func (q *queue) receiveT(t *testing.T, i interface{}) error {
	if err := q.receive(i); err != nil {
		t.Fatal("Got error durint receive, err: ", err)
	}
	return nil
}

func OnReceiveFn(fn func(i interface{})) AfterFn {
	return AfterFn{
		fn: fn,
	}
}

type AfterFn struct {
	fn func(i interface{})
}

type Step struct {
	// Di
	Send    interface{}
	Receive interface{}

	resp interface{}
	err  error
}
