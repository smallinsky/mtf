package fswatch

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"time"

	"google.golang.org/grpc"

	pb "github.com/smallinsky/mtf/pkg/fswatch/proto"
)

func Monitor(addr, dir string) {
	client, err := newWatcherClient(addr)
	if err != nil {
		log.Fatalf("failed to create watcher client: %v", err)
	}

	pub := &ActionHandler{
		Client: client,
	}

	w := &Watcher{
		Dir:          dir,
		EventHandler: pub,
		stop:         make(chan struct{}),
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		w.Stop()
	}()

	if err := w.Run(); err != nil {
		log.Fatalf("[ERR] watcher run: %v", err)
	}
}

func newWatcherClient(addr string) (pb.WatcherClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return pb.NewWatcherClient(conn), nil
}

// ActionHandler publish directory changes over to remote grpc server.
type ActionHandler struct {
	Client pb.WatcherClient
}

// OnFileCreated sends Create action details to remote events server.
func (g *ActionHandler) OnFileCreated(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file '%v': %s", path, err)
	}
	time.Sleep(time.Millisecond * 100)

	buff, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read file '%v': %s", path, err)
	}
	defer f.Close()

	_, err = g.Client.Event(context.Background(), &pb.EventRequest{
		Path:    path,
		Action:  pb.Action_ADDED,
		Content: buff,
	})
	if err != nil {
		return fmt.Errorf("failed to send event: %s", err)
	}
	return nil
}

// OnFileDeleted sends remove action details to remote events server.
func (g *ActionHandler) OnFileDeleted(path string) error {
	_, err := g.Client.Event(context.Background(), &pb.EventRequest{
		Path:   path,
		Action: pb.Action_REMOVED,
	})
	if err != nil {
		return fmt.Errorf("failed to send event: %s", err)
	}
	return nil
}
