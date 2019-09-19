// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// Package fsw wraps filesystem watchers to ensure the best
// experience on the operating system that vidar was built for.
// For example, fsnotify is used on most systems; but on darwin,
// where the open file limits are low and watchers use up open
// files, we instead use a polling watcher.
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
	Add(name string) error
	Remove(name string) error
	RemoveAll() error
	Close() error
	Next() (Event, error)
}
