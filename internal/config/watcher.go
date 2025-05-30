package config

import (
	"log/slog"

	"github.com/fsnotify/fsnotify"
)

// Watcher наблюдатель за изменениями в файле
type Watcher struct {
	watcher *fsnotify.Watcher
}

// NewWatcher создает новый наблюдатель для файла
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

// Close закрывает наблюдатель
func (w *Watcher) Close() {
	if err := w.watcher.Close(); err != nil {
		slog.Error("error closing watcher", "error", err)
	}
}

// DoRun запускает горутину, которая наблюдает за изменениями в файле и вызывает функцию fn при изменении файла
func (w *Watcher) DoRun(fn func()) {
	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					slog.Info("modified file", "file", event.Name)
					fn()
				}
			case err := <-w.watcher.Errors:
				slog.Error("error", "error", err)
			}
		}
	}()
}
