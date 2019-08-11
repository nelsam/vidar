// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// Package fsw wraps filesystem watchers to ensure the best
// experience on the operating system that vidar was built for.
// For example, we use fsnotify on linux, but a polling watcher
// on darwin where the open file limits conflict with fsnotify's
// watcher.
package fsw

type Op uint32

const (
	Create Op = 1 << iota
	Write
	Remove
	Rename
	Chmod
)

type Event struct {
	Path string
	Op   Op
}

type Watcher interface {
	// Watch creates a new EventHandler, adding path to its watched
	// paths.
	Watch(path string) (EventHandler, error)
}

// EventHandler is a type that handles filesystem events, reporting
// them back to the caller and allowing the watched paths to be
// altered.
type EventHandler interface {
	Add(name string) error
	Remove(name string) error
	Next() (Event, error)
	Close() error
}
