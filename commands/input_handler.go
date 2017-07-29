// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/input"
	"github.com/nelsam/vidar/editor"
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

type textChangeHook interface {
	init(input.Editor, []rune)
	textChanged(input.Editor, []input.Edit) error
}

// InputHandler is vidar's default input handler.  It takes the runes typed
// and writes them to the editor's text box, allowing plugins to bind to text
// change events.
//
// It can be overridden by installing a plugin that implements
// commander.InputHandler.
type InputHandler struct {
	driver gxui.Driver
	hooks  []textChangeHook
}

func NewInputHandler(d gxui.Driver) *InputHandler {
	return &InputHandler{driver: d}
}

func (e *InputHandler) Name() string {
	return "input-handler"
}

func (e *InputHandler) New() input.Handler {
	return NewInputHandler(e.driver)
}

func (e *InputHandler) Bind(h bind.CommandHook) error {
	switch src := h.(type) {
	case ChangeHook:
		h := &hookReader{hook: src, driver: e.driver}
		h.start()
		e.hooks = append(e.hooks, h)
	case ContextChangeHook:
		e.hooks = append(e.hooks, &ctxHookReader{hook: src, driver: e.driver})
	default:
		return fmt.Errorf("expected ChangeHook or ContextChangeHook; got %T", h)
	}
	return nil
}

func (e *InputHandler) Init(newEditor input.Editor, contents []rune) {
	for _, h := range e.hooks {
		h.init(newEditor, contents)
	}
}

func (h *InputHandler) Apply(e input.Editor, edits ...input.Edit) {
	c := e.(*editor.CodeEditor).Controller()
	text := c.TextRunes()
	delta := 0
	sort.Slice(edits, func(i, j int) bool {
		return edits[i].At < edits[j].At
	})
	scrollTo := 0
	for _, edit := range edits {
		oldE := edit.At + delta + len(edit.Old)
		newE := edit.At + delta + len(edit.New)
		if oldE != newE {
			if newE > oldE {
				text = append(text, make([]rune, newE-oldE)...)
			}
			copy(text[newE:], text[oldE:])
			if oldE > newE {
				text = text[:len(text)-(oldE-newE)]
			}
		}
		copy(text[edit.At+delta:newE], edit.New)
		scrollTo = edit.At + delta
		delta += newE - oldE
	}
	h.driver.Call(func() {
		c.SetTextRunes(text)
		e.(*editor.CodeEditor).ScrollToRune(scrollTo)
		h.textEdited(e, edits)
	})
}

func (e *InputHandler) HandleInput(focused input.Editor, ev gxui.KeyStrokeEvent) {
	if ev.Modifier&^gxui.ModShift != 0 {
		return
	}
	editor := focused.(*editor.CodeEditor)
	var edits []input.Edit
	for _, c := range editor.Controller().ReplaceAllRunes([]rune{ev.Character}) {
		edits = append(edits, input.Edit{
			At:  c.At,
			Old: c.Old,
			New: c.New,
		})
	}
	editor.InputEventHandler.KeyStroke(ev)
	e.textEdited(focused, edits)
}

func (e *InputHandler) textEdited(focused input.Editor, edits []input.Edit) {
	for _, h := range e.hooks {
		if err := h.textChanged(focused, edits); err != nil {
			log.Printf("Hook %v failed: %s", h, err)
		}
	}
}

func contextDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
