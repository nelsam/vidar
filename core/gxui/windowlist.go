// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package gxui

import "sync"

type windowList struct {
	list []*Window
	mu   sync.Mutex
}

func (l *windowList) add(w *Window) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.list = append(l.list, w)
}

func (l *windowList) remove(w *Window) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i, ew := range l.list {
		if ew != w {
			continue
		}
		l.list = append(l.list[:i], l.list[i+1:]...)
		break
	}
}

func (l *windowList) len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.list)
}
