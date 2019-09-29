package fswatch

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	pb "github.com/smallinsky/mtf/pkg/fswatch/proto"
)

func TestDirWatcher(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatalf("faield to create tmp dir: %v", err)
	}

	event := make(chan *pb.EventRequest)
	go func() {
		Subscriber("localhost:11132", func(req *pb.EventRequest) {
			event <- req

		})
	}()
	go func() {
		Monitor("localhost:11132", tmpDir)
	}()

	time.Sleep(time.Millisecond * 100)

	tmpFile := fmt.Sprintf("%s/%s", tmpDir, "tmpFile.txt")
	if err := ioutil.WriteFile(tmpFile, []byte("tmp content"), 0644); err != nil {
		t.Fatalf("failed to writie to file: %v", err)
	}

	select {
	case got := <-event:
		exp := &pb.EventRequest{
			Path:    tmpFile,
			Action:  pb.Action_ADDED,
			Content: []byte("tmp content"),
		}
		if !reflect.DeepEqual(got, exp) {
			t.Fatalf("got: %+v\nexp: %+v", got, exp)
		}
	case <-time.After(time.Second):
		t.Fatalf("file event was not recived")
	}

	if err := os.Remove(tmpFile); err != nil {
		t.Fatalf("failed to remove tmp file %v", err)
	}

	select {
	case got := <-event:
		exp := &pb.EventRequest{
			Path:   tmpFile,
			Action: pb.Action_REMOVED,
		}
		if !reflect.DeepEqual(got, exp) {
			t.Fatalf("got: %+v\nexp: %+v", got, exp)
		}
	case <-time.After(time.Second):
		t.Fatalf("file event was not recived")
	}
}
