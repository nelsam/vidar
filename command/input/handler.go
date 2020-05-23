// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package input

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/text"
	"github.com/nelsam/vidar/editor"
)

type Binder interface {
	Bindable(name string) bind.Bindable
	Execute(bind.Bindable)
}

type textChangeHook interface {
	init(text.Editor, []rune)
	textChanged(text.Editor, []text.Edit) error
}

// Handler is vidar's default input handler.  It takes the runes typed
// and writes them to the editor's text box, allowing plugins to bind to text
// change events.
//
// It can be overridden by installing a plugin that implements
// commander.Handler.
//
// TODO: Remove the need for so much of the editor's methods.  Ideally, we should
// be able to do what we need to do with just text.Editor.  This may require
// expanding text.Editor some, but we shouldn't need editor.Controller for so
// much of what we're doing.  Basic editing should be handleable by the editor.
type Handler struct {
	driver gxui.Driver
	binder Binder

	hooks      []textChangeHook
	applied    []AppliedChangeHook
	cancellers []Canceler
	confirmers []Confirmer
}

func New(d gxui.Driver, b Binder) *Handler {
	return &Handler{driver: d, binder: b}
}

func (e *Handler) Name() string {
	return "input-handler"
}

func (e *Handler) New() text.Handler {
	return New(e.driver, e.binder)
}

func (e *Handler) Bind(b bind.Bindable) (text.Handler, error) {
	newH := New(e.driver, e.binder)
	newH.hooks = append(newH.hooks, e.hooks...)
	newH.applied = append(newH.applied, e.applied...)
	newH.cancellers = append(newH.cancellers, e.cancellers...)
	newH.confirmers = append(newH.confirmers, e.confirmers...)

	didBind := false
	if c, isCanceler := b.(Canceler); isCanceler {
		newH.cancellers = append(newH.cancellers, c)
		didBind = true
	}

	if c, isConfirmer := b.(Confirmer); isConfirmer {
		newH.confirmers = append(newH.confirmers, c)
		didBind = true
	}

	if a, ok := b.(AppliedChangeHook); ok {
		didBind = true
		newH.applied = append(newH.applied, a)
	}
	switch src := b.(type) {
	case ChangeHook:
		didBind = true
		r := &hookReader{hook: src, driver: e.driver}
		r.start()
		newH.hooks = append(newH.hooks, r)
	case ContextChangeHook:
		didBind = true
		newH.hooks = append(newH.hooks, &ctxHookReader{hook: src, driver: newH.driver})
	}

	if !didBind {
		return nil, fmt.Errorf("text.Handler: type %T did not implement any known hook type", b)
	}

	return newH, nil
}

func (e *Handler) Init(newEditor text.Editor, contents []rune) {
	for _, h := range e.hooks {
		h.init(newEditor, contents)
	}
}

func (h *Handler) Apply(e text.Editor, edits ...text.Edit) {
	editor := e.(*editor.CodeEditor)
	c := editor.Controller()
	text := c.TextRunes()
	delta := 0
	sort.Slice(edits, func(i, j int) bool {
		return edits[i].At < edits[j].At
	})
	for i, edit := range edits {
		// It's a pretty common problem that the edit.Old and/or edit.New is
		// a slice of the actual data being edited, so copy it to preserve the
		// edit.
		edit.Old = clone(edit.Old)
		edit.New = clone(edit.New)
		edit.At += delta
		edits[i] = edit

		oldE := edit.At + len(edit.Old)
		newE := edit.At + len(edit.New)
		if oldE != newE {
			if newE > oldE {
				text = append(text, make([]rune, newE-oldE)...)
			}
			if newE < len(text) {
				copy(text[newE:], text[oldE:])
			}
			if oldE > newE {
				text = text[:len(text)-(oldE-newE)]
			}
		}
		copy(text[edit.At:newE], edit.New)
		delta += newE - oldE
	}
	c.Deselect(false)
	h.driver.CallSync(func() {
		c.SetTextRunesNoEvent(text)
		h.textEdited(e, edits)
	})
}

func (e *Handler) HandleEvent(focused text.Editor, ev gxui.KeyboardEvent) {
	if ev.Modifier&^gxui.ModShift != 0 {
		// This Handler, at least for now, doesn't handle key bindings.
		return
	}
	editor := focused.(*editor.CodeEditor)
	ctrl := editor.Controller()
	switch ev.Key {
	case gxui.KeyEnter:
		for _, c := range e.confirmers {
			if c.Confirm(focused) {
				return
			}
		}
		var edits []text.Edit
		for _, s := range editor.Controller().SelectionSlice() {
			if s.Start() < 0 {
				continue
			}
			edits = append(edits, text.Edit{
				At:  s.Start(),
				Old: ctrl.TextRunes()[s.Start():s.End()],
				New: []rune{'\n'},
			})
		}
		e.Apply(focused, edits...)
	case gxui.KeyEscape:
		for _, c := range e.cancellers {
			if c.Cancel(focused) {
				return
			}
		}
		for _, h := range e.hooks {
			// if esc is pressed without any cancellers waiting, cancel *all*
			// working text change hooks that can be cancelled.
			if canceler, ok := h.(Canceler); ok {
				_ = canceler.Cancel(focused)
			}
		}
	case gxui.KeyBackspace, gxui.KeyDelete:
		var edits []text.Edit
		for _, s := range editor.Controller().SelectionSlice() {
			if s.Start() < 0 {
				continue
			}
			edit := text.Edit{
				At:  s.Start(),
				Old: ctrl.TextRunes()[s.Start():s.End()],
			}
			if s.Start() == s.End() {
				if ev.Key == gxui.KeyBackspace {
					if edit.At == 0 {
						continue
					}
					edit.At--
				}
				edit.Old = ctrl.TextRunes()[edit.At : edit.At+1]
			}
			edits = append(edits, edit)
		}
		e.Apply(focused, edits...)
	}
}

func (e *Handler) HandleInput(focused text.Editor, ev gxui.KeyStrokeEvent) {
	if ev.Modifier&^gxui.ModShift != 0 {
		return
	}
	editor := focused.(*editor.CodeEditor)
	ctrl := editor.Controller()
	var edits []text.Edit
	for _, s := range editor.Controller().SelectionSlice() {
		edits = append(edits, text.Edit{
			At:  s.Start(),
			Old: ctrl.TextRunes()[s.Start():s.End()],
			New: []rune{ev.Character},
		})
	}
	e.Apply(focused, edits...)
}

func (e *Handler) textEdited(focused text.Editor, edits []text.Edit) {
	for _, a := range e.applied {
		a.Applied(focused, edits)
	}
	for _, h := range e.hooks {
		if err := h.textChanged(focused, edits); err != nil {
			log.Printf("Hook %v failed: %s", h, err)
		}
	}
}

func clone(t []rune) []rune {
	newT := make([]rune, len(t))
	copy(newT, t)
	return newT
}

func contextDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
