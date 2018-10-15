// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// +build !darwin,!windows

package fsw

import (
	"io"

	"github.com/fsnotify/fsnotify"
)

type watcher struct {
	*fsnotify.Watcher
}

func New() (Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &watcher{Watcher: fsw}, nil
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
