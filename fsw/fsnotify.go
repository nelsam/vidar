// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// +build !darwin,!windows

package fsw

import (
	"io"

	"github.com/fsnotify/fsnotify"
)

type fsnotifyWatcher struct{}

func New() Watcher {
	return &fsnotifyWatcher{}
}

func (w fsnotifyWatcher) Watch(path string) (EventHandler, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return fsnotifyEventHandler{Watcher: fsw}, nil
}

type fsnotifyEventHandler struct {
	*fsnotify.Watcher
}

func (h fsnotifyEventHandler) Next() (Event, error) {
	select {
	case e, ok := <-h.Events:
		if !ok {
			return Event{}, io.EOF
		}
		return Event{Path: e.Name, Op: Op(e.Op)}, nil
	case err, ok := <-h.Errors:
		if !ok {
			return Event{}, io.EOF
		}
		return Event{}, err
	}
}
