package main

import (
	_ "runtime"
	_ "unsafe"
)

////go:linkname foo time.Now
//func foo() time.Time {
//	return time.Time{}
//}
//
//func sleep() {
//	time.Sleep(time.Second * 10)
//	time.Now()
//}

type mutex struct {
	key uintptr
}

//go:linkname fof runtime.timeSleep
func fof(ns int64)

//go:linkname zz runtime.timeSleepUntil
func zz() int64

//go:linkname aa runtime.time_now
func aa() (sec int64, nsec int32, mono int64)

func main() {
	aa()
}
