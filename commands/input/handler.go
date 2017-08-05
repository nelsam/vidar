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
	"github.com/nelsam/vidar/commander/input"
	"github.com/nelsam/vidar/editor"
)

type textChangeHook interface {
	init(input.Editor, []rune)
	textChanged(input.Editor, []input.Edit) error
}

// Handler is vidar's default input handler.  It takes the runes typed
// and writes them to the editor's text box, allowing plugins to bind to text
// change events.
//
// It can be overridden by installing a plugin that implements
// commander.Handler.
//
// TODO: Remove the need for so much of the editor's methods.  Ideally, we should
// be able to do what we need to do with just input.Editor.  This may require
// expanding input.Editor some, but we shouldn't need editor.Controller for so
// much of what we're doing.  Basic editing should be handleable by the editor.
type Handler struct {
	driver gxui.Driver
	hooks  []textChangeHook
}

func New(d gxui.Driver) *Handler {
	return &Handler{driver: d}
}

func (e *Handler) Name() string {
	return "input-handler"
}

func (e *Handler) New() input.Handler {
	return New(e.driver)
}

func (e *Handler) Bind(h bind.CommandHook) error {
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

func (e *Handler) Init(newEditor input.Editor, contents []rune) {
	for _, h := range e.hooks {
		h.init(newEditor, contents)
	}
}

func (h *Handler) Apply(e input.Editor, edits ...input.Edit) {
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
		edits[i] = edit

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
		delta += newE - oldE
	}
	c.Deselect(false)
	h.driver.Call(func() {
		c.SetTextRunes(text)
		h.textEdited(e, edits)
		if editor.IsSuggestionListShowing() {
			editor.SortSuggestionList()
		}
	})
}

func (e *Handler) HandleEvent(focused input.Editor, ev gxui.KeyboardEvent) {
	if ev.Modifier&^gxui.ModShift != 0 {
		// This Handler, at least for now, doesn't handle key bindings.
		return
	}
	editor := focused.(*editor.CodeEditor)
	ctrl := editor.Controller()
	switch ev.Key {
	case gxui.KeyBackspace, gxui.KeyDelete:
		var edits []input.Edit
		for _, s := range editor.Controller().Selections() {
			if s.Start() < 0 {
				continue
			}
			edit := input.Edit{
				At:  s.Start(),
				Old: ctrl.TextRunes()[s.Start():s.End()],
			}
			if s.Start() == s.End() {
				if ev.Key == gxui.KeyBackspace {
					edit.At--
				}
				edit.Old = ctrl.TextRunes()[edit.At : edit.At+1]
			}
			edits = append(edits, edit)
		}
		e.Apply(focused, edits...)
	}
}

func (e *Handler) HandleInput(focused input.Editor, ev gxui.KeyStrokeEvent) {
	if ev.Modifier&^gxui.ModShift != 0 {
		return
	}
	editor := focused.(*editor.CodeEditor)
	ctrl := editor.Controller()
	var edits []input.Edit
	for _, s := range editor.Controller().Selections() {
		edits = append(edits, input.Edit{
			At:  s.Start(),
			Old: ctrl.TextRunes()[s.Start():s.End()],
			New: []rune{ev.Character},
		})
	}
	e.Apply(focused, edits...)
}

func (e *Handler) textEdited(focused input.Editor, edits []input.Edit) {
	for _, h := range e.hooks {
		if err := h.textChanged(focused, edits); err != nil {
			log.Printf("Hook %v failed: %s", h, err)
		}
	}
	var gEdits []gxui.TextBoxEdit
	for _, e := range edits {
		gEdits = append(gEdits, gxui.TextBoxEdit{
			At:    e.At,
			Delta: len(e.New) - len(e.Old),
			Old:   e.Old,
			New:   e.New,
		})
	}
	focused.(*editor.CodeEditor).Controller().TextEdited(gEdits)
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
