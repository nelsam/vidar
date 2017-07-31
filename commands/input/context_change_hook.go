// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package input

import (
	"context"
	"log"
	"sync"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/input"
)

// ContextChangeHook is similar to a ChangeHook, but takes a
// context.Context that will be cancelled if new changes show
// up before Apply is called.
type ContextChangeHook interface {
	// Init is called when a file is opened, to initialize the
	// hook.  The full text of the editor will be passed in.
	Init(input.Editor, []rune)

	// TextChanged is called in a new goroutine whenever any text
	// is changed in the editor.  Any changes to the UI should be
	// saved for Apply, since most of those calls must be called
	// in the UI goroutine.
	//
	// If TextChanged is currently running and new edits come
	// through, the context.Context will be cancelled and
	// TextChanged will be called again with the new edits appended
	// to those from the previous call.
	TextChanged(context.Context, input.Editor, []input.Edit)

	// Apply is called when there is a break in text changes, to
	// apply the hook's event.  Unlike TextChanged, Apply is
	// called in the main UI thread.
	Apply(input.Editor) error
}

type ctxHookReader struct {
	driver gxui.Driver
	cancel func()
	edits  []input.Edit
	mu     sync.Mutex
	hook   ContextChangeHook
}

func (r *ctxHookReader) init(e input.Editor, text []rune) {
	r.hook.Init(e, text)
	r.hook.Apply(e)
}

func (r *ctxHookReader) textChanged(e input.Editor, changes []input.Edit) error {
	if len(changes) == 0 {
		return nil
	}

	if r.cancel != nil {
		r.cancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel

	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range changes {
		r.edits = append(r.edits, c)
	}
	go func() {
		if contextDone(ctx) {
			return
		}

		r.mu.Lock()
		defer r.mu.Unlock()
		r.hook.TextChanged(ctx, e, r.edits)

		if contextDone(ctx) {
			return
		}

		r.edits = nil
		r.driver.Call(func() {
			r.mu.Lock()
			defer r.mu.Unlock()
			if contextDone(ctx) {
				return
			}
			if err := r.hook.Apply(e); err != nil {
				log.Printf("Error applying hook %v to editor %v: %s", r.hook, e, err)
			}
		})
	}()
	return nil
}
