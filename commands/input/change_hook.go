// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package input

import (
	"log"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/input"
)

// ChangeHook is a hook that triggers events on text changing.
// Each ChangeHook gets its own goroutine that will be used to
// process events.  When there is a lull in the number of events
// coming through (i.e. there is a pause in typing), Apply will
// be called from the UI goroutine.
//
// For plugins that need to process all events at once, or apply
// to the entire file instead of one edit at a time, implement
// ContextChangeHook.
type ChangeHook interface {
	// Init is called when a file is opened, to initialize the
	// hook.  The full text of the editor will be passed in.
	Init(input.Editor, []rune)

	// TextChanged is called for every text edit.
	//
	// TextChanged may be called multiple times prior to Apply if
	// more edits happen before Apply is called.
	//
	// Hooks that have to start over from scratch when new updates
	// are made should implement ContextChangeHook, instead.
	TextChanged(input.Editor, input.Edit)

	// Apply is called when there is a break in text changes, to
	// apply the hook's event.  Unlike TextChanged, Apply is
	// called in the main UI thread.
	Apply(input.Editor) error
}

type editNode struct {
	edit   input.Edit
	editor input.Editor
	next   unsafe.Pointer
}

func (e *editNode) nextNode() *editNode {
	p := atomic.LoadPointer(&e.next)
	if p == nil {
		return nil
	}
	return (*editNode)(p)
}

type hookReader struct {
	driver gxui.Driver
	cond   *sync.Cond
	next   unsafe.Pointer
	last   unsafe.Pointer
	hook   ChangeHook
}

func (r *hookReader) start() {
	if r.cond != nil {
		return
	}
	r.cond = sync.NewCond(&sync.Mutex{})
	go r.run()
	r.cond.Signal()
}

func (r *hookReader) run() {
	for {
		r.processEdits()
	}
}

func (r *hookReader) processEdits() error {
	r.cond.Wait()
	nextPtr := atomic.LoadPointer(&r.next)
	if nextPtr == nil {
		return nil
	}
	editors := make(map[input.Editor]struct{})
	for nextPtr != nil {
		next := (*editNode)(nextPtr)
		r.hook.TextChanged(next.editor, next.edit)
		editors[next.editor] = struct{}{}
		nextPtr = atomic.LoadPointer(&next.next)
	}
	atomic.StorePointer(&r.next, nil)
	for e := range editors {
		r.driver.Call(func() {
			if err := r.hook.Apply(e); err != nil {
				log.Printf("Error applying changes to editor %v: %s", e, err)
			}
		})
	}
	return nil
}

func (r *hookReader) init(e input.Editor, text []rune) {
	r.hook.Init(e, text)
	r.hook.Apply(e)
}

func (r *hookReader) textChanged(e input.Editor, changes []input.Edit) error {
	if len(changes) == 0 {
		return nil
	}
	lPtr := atomic.LoadPointer(&r.last)
	if lPtr == nil {
		lPtr = unsafe.Pointer(&editNode{})
	}
	l := (*editNode)(lPtr)
	for _, edit := range changes {
		n := &editNode{
			editor: e,
			edit:   edit,
		}
		atomic.StorePointer(&l.next, unsafe.Pointer(n))
		if atomic.CompareAndSwapPointer(&r.next, nil, unsafe.Pointer(n)) {
			r.cond.Signal()
		}
		l = n
	}
	atomic.StorePointer(&r.last, unsafe.Pointer(l))
	return nil
}
