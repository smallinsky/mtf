package port

import (
	"bufio"
	"bytes"
	"fmt"
	"runtime"
	"testing"
	"unsafe"
)

func getT() *testing.T {
	var buf [8192]byte
	n := runtime.Stack(buf[:], false)
	sc := bufio.NewScanner(bytes.NewReader(buf[:n]))
	for sc.Scan() {
		var p uintptr
		n, _ := fmt.Sscanf(sc.Text(), "testing.tRunner(%v", &p)
		if n != 1 {
			continue
		}
		return (*testing.T)(unsafe.Pointer(p))
	}
	panic("T stack args not found")
}
