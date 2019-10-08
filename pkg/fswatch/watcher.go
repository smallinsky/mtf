package fswatch

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
)

func New(dir string, eventHanlder EventHandler) *Watcher {
	return &Watcher{
		Dir:          dir,
		EventHandler: eventHanlder,
		stop:         make(chan struct{}),
	}
}

type EventHandler interface {
	OnFileCreated(path string) error
	OnFileDeleted(path string) error
}

// Watcher watches the Dir and sends notification about file system changes events to EventHandler.
type Watcher struct {
	Dir          string
	EventHandler EventHandler
	stop         chan struct{}
}

// Run starts dir notification watcher and forwards events to EventHandler.
func (w *Watcher) Run() error {
	if w.EventHandler == nil {
		return fmt.Errorf("EventHandler can't be nil")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %v", err)
	}
	defer watcher.Close()
	watcher.Add(w.Dir)

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Create == fsnotify.Create {
				if err := w.EventHandler.OnFileCreated(event.Name); err != nil {
					return fmt.Errorf("failed to call OnFileCreated: %v", err)
				}
			}
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				if err := w.EventHandler.OnFileDeleted(event.Name); err != nil {
					return fmt.Errorf("failed to call OnFileDeleted: %v", err)
				}
			}
		case err := <-watcher.Errors:
			return fmt.Errorf("watcher got err: %v", err)
		case <-w.stop:
			log.Printf("[INFO] stopping watcher")
			return nil
		}
	}
}

// Stop the watcher.
func (w *Watcher) Stop() {
	close(w.stop)
}
