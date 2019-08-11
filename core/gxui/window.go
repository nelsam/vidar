// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package gxui

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"runtime/debug"
	"sync/atomic"
	"unsafe"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/vidar/asset"
	"github.com/nelsam/vidar/input"
	"github.com/nelsam/vidar/ui"
)

// A Controllable is a type that has a gxui.Control.
type Controllable interface {
	Control() gxui.Control
}

func control(v interface{}) (gxui.Control, error) {
	if c, ok := v.(Controllable); ok {
		return c.Control(), nil
	}
	return nil, fmt.Errorf("gxui: cannot use %v (type %T) as type Controllable", v, v)
}

// icon returns the image.Image to be used as vidar's icon.
func icon() image.Image {
	r := bytes.NewReader(asset.MustAsset("icon.png"))
	i, _, err := image.Decode(r)
	if err != nil {
		panic(fmt.Errorf("could not decode png: %s", err))
	}
	return i
}

// Window is a wrapper for gxui.Window, implementing ui.Window
type Window struct {
	// Keep all fields that will be used for atomic operations
	// first, for byte-alignment reasons.
	handlerAddr unsafe.Pointer
	focusedAddr unsafe.Pointer

	window gxui.Window
	child  Controllable
}

func newWindow(t gxui.Theme, size ui.Size, name string) *Window {
	w := &Window{
		window: t.CreateWindow(int(size.Width), int(size.Height), name),
	}
	w.window.SetIcon(icon())

	// TODO: Check the system's DPI settings for this value
	w.window.SetScale(1)

	w.window.SetPadding(math.Spacing{L: 10, T: 10, R: 10, B: 10})
	return w
}

func (w *Window) Raw() gxui.Window {
	return w.window
}

func (w *Window) SetChild(child interface{}) error {
	c, ok := child.(Controllable)
	if !ok {
		return fmt.Errorf("Type %T is not Controllable and cannot be used in *Window.SetChild", child)
	}
	w.window.RemoveAll()
	w.window.AddChild(c.Control())
	w.child = c
	return nil
}

func (w *Window) Size() ui.Size {
	s := w.window.Size()
	return ui.Size{Width: uint64(s.W), Height: uint64(s.H)}
}

func (w *Window) HandleInput(h input.Handler) {
	atomic.StorePointer(&w.handlerAddr, unsafe.Pointer(&h))
}

func (w *Window) handler() input.Handler {
	p := atomic.LoadPointer(&w.handlerAddr)
	if p == nil {
		return nil
	}
	return *((*input.Handler)(p))
}

func (w *Window) storeFocused(e input.Editor) {
	atomic.StorePointer(&w.focusedAddr, unsafe.Pointer(&e))
}

func (w *Window) focused() input.Editor {
	p := atomic.LoadPointer(&w.focusedAddr)
	if p == nil {
		return nil
	}
	return *((*input.Editor)(p))
}

func (w *Window) SetFocus(f gxui.Focusable) bool {
	// TODO: this won't work.  The input.Editor is not going to be the
	// thing passed to SetFocus.
	//
	// Ideas:
	// - We could look through all the input.Editors in the window
	//   and see if their underlying gxui.CodeEditor matches?
	// - We could add focus to our exported `ui` contract, force our
	//   internal code to understand focus and pass it along to the
	//   window?
	// Notes:
	// - This isn't a critical core feature - performance doesn't
	//   matter much.
	// - We should focus on what makes it easiest to write new UI
	//   wrappers.  The core internal logic is only going to happen
	//   once.
	if e, ok := f.(input.Editor); ok {
		w.storeFocused(e)
	}
	return w.window.SetFocus(f)
}

// KeyPress handles key bindings for w.
func (w *Window) KeyPress(event gxui.KeyboardEvent) (consume bool) {
	defer func() {
		if r := recover(); r != nil {
			// TODO: display this in the UI
			log.Printf("ERR: panic while handling key event: %v", r)
			log.Printf("Stack trace:\n%s", debug.Stack())
		}
	}()
	focused := w.focused()
	if focused == nil {
		return false
	}
	w.handler().HandleEvent(focused, event)
	return true
}

func (w *Window) KeyStroke(event gxui.KeyStrokeEvent) (consume bool) {
	defer func() {
		if r := recover(); r != nil {
			// TODO: display this in the UI
			log.Printf("ERR: panic while handling key stroke: %v", r)
			log.Printf("Stack trace:\n%s", debug.Stack())
		}
	}()
	focused := w.focused()
	if focused == nil {
		return false
	}
	w.handler().HandleInput(focused, event)
	return true
}
