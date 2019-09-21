// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// +build !darwin,!windows

package fsw

import (
	"io"
	"log"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type watcher struct {
	*fsnotify.Watcher

	tracking map[string]struct{}
	mu       sync.Mutex
}

func New() (Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &watcher{Watcher: fsw, tracking: make(map[string]struct{})}, nil
}

func (w *watcher) Add(path string) error {
	if err := w.Watcher.Add(path); err != nil {
		return err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.tracking[path] = struct{}{}
	return nil
}

func (w *watcher) Remove(path string) error {
	if err := w.Watcher.Remove(path); err != nil {
		return err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.tracking, path)
	return nil
}

func (w *watcher) Next() (Event, error) {
	select {
	case e, ok := <-w.Events:
		if !ok {
			return Event{}, io.EOF
		}
		return Event{Path: e.Name, Op: Op(e.Op)}, nil
	case err, ok := <-w.Errors:
		if !ok {
			return Event{}, io.EOF
		}
		return Event{}, err
	}
}

func (w *watcher) RemoveAll() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	for p := range w.tracking {
		if err := w.Watcher.Remove(p); err != nil {
			log.Printf("WARNING: error removing tracking path %s: %s", p, err)
			continue
		}
		delete(w.tracking, p)
	}
	return nil
}
