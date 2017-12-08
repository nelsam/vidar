// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// +build darwin

package fsw

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

const pollDuration = 200 * time.Millisecond

type poll map[string]os.FileInfo

func (p poll) addChildren(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	c, err := f.Readdir(0)
	if err != nil {
		return err
	}
	for _, info := range c {
		p[filepath.Join(name, info.Name())] = info
	}
	return nil
}

type poller struct {
	closed uint32
	mu     sync.Mutex
	ticker *time.Ticker
	last   map[string]poll
	events chan Event
	errs   chan error
}

func New() (Watcher, error) {
	p := &poller{
		ticker: time.NewTicker(pollDuration),
		last:   make(map[string]poll),
		events: make(chan Event),
		errs:   make(chan error),
	}
	go p.run()
	return p, nil
}

func (p *poller) run() {
	defer close(p.events)
	for range p.ticker.C {
		if closed := atomic.LoadUint32(&p.closed); closed != 0 {
			return
		}
		p.triggerEvents()
	}
}

func (p *poller) triggerEvents() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for name, last := range p.last {
		i, err := os.Stat(name)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			p.errs <- err
			continue
		}
		next := make(poll)
		next[name] = i
		if i.IsDir() {
			if err := next.addChildren(name); err != nil {
				p.errs <- err
				continue
			}
		}
		p.triggerDiff(last, next)
		p.last[name] = next
	}
}

func (p *poller) triggerDiff(prev, next poll) {
	creates := make(map[string]os.FileInfo)
	for name, nexti := range next {
		previ, ok := prev[name]
		if !ok {
			creates[name] = nexti
			continue
		}
		delete(prev, name)
		if previ.ModTime() != nexti.ModTime() {
			p.events <- Event{Op: Write, Path: name}
		}
		if previ.Mode() != nexti.Mode() {
			p.events <- Event{Op: Chmod, Path: name}
		}
	}

	for name, previ := range prev {
		ev := Event{Op: Remove, Path: name}
		for _, nexti := range creates {
			if os.SameFile(previ, nexti) {
				ev.Op = Rename
				break
			}
		}
		p.events <- ev
	}

	for name := range creates {
		p.events <- Event{Op: Create, Path: name}
	}
}

func (p *poller) Add(name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	init := make(poll)
	i, err := os.Stat(name)
	if err != nil {
		return err
	}
	init[name] = i
	if i.IsDir() {
		if err := init.addChildren(name); err != nil {
			return err
		}
	}
	p.last[name] = init
	return nil
}

func (p *poller) Remove(name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.last, name)
	return nil
}

func (p *poller) Close() error {
	p.ticker.Stop()
	atomic.StoreUint32(&p.closed, 1)
	return nil
}

func (p *poller) Next() (Event, error) {
	select {
	case e, ok := <-p.events:
		if !ok {
			return Event{}, io.EOF
		}
		return e, nil
	case err := <-p.errs:
		return Event{}, err
	}
}
