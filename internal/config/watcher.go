package config

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	watcher *fsnotify.Watcher
}

func NewWatcher(filepath string) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = watcher.Add(filepath)
	if err != nil {
		return nil, err
	}

	return &Watcher{watcher: watcher}, nil
}

func (w *Watcher) Close() {
	if err := w.watcher.Close(); err != nil {
		log.Println("error closing watcher:", err)
	}
}

func (w *Watcher) DoRun(fn func()) {
	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Println("modified file:", event.Name)
					fn()
				}
			case err := <-w.watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()
}
