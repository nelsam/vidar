// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package gocode

import (
	"context"
	"log"
	"sync"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/command/caret"
	"github.com/nelsam/vidar/commander/text"
	"github.com/nelsam/vidar/setting"
)

type Projecter interface {
	Project() setting.Project
}

func New(theme *basic.Theme, driver gxui.Driver) (*Completions, *GoCode) {
	g := GoCode{
		driver:  driver,
		lists:   make(map[Editor]*suggestionList),
		cancels: make(map[Editor]func()),
	}
	c := Completions{
		gocode: &g,
	}
	c.Theme = theme
	return &c, &g
}

func ctxCancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

type GoCode struct {
	driver gxui.Driver

	mu      sync.RWMutex
	lists   map[Editor]*suggestionList
	cancels map[Editor]func()
}

func (g *GoCode) Name() string {
	return "gocode-updates"
}

func (g *GoCode) OpNames() []string {
	return []string{"caret-movement", "input-handler"}
}

func (g *GoCode) cancel(e Editor) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if cancel, ok := g.cancels[e]; ok {
		cancel()
		delete(g.cancels, e)
	}
}

func (g *GoCode) show(ctx context.Context, l *suggestionList, pos int) {
	n := l.show(ctx, pos)
	if n == 0 || ctxCancelled(ctx) {
		// TODO: Add this as a UI message.
		log.Printf("gocode: found no results (or context cancelled)")
		return
	}

	bounds := l.editor.Size().Rect().Contract(l.editor.Padding())
	line := l.editor.Line(l.editor.LineIndex(pos))
	lineOffset := gxui.ChildToParent(math.ZeroPoint, line, l.editor)
	target := line.PositionAt(pos).Add(lineOffset)
	cs := l.DesiredSize(math.ZeroSize, bounds.Size())

	g.driver.Call(func() {
		if ctxCancelled(ctx) {
			// TODO: Add this as a UI message.
			log.Printf("cancelled")
			return
		}
		l.SetSize(cs)
		c := l.editor.AddChild(l)
		c.Layout(cs.Rect().Offset(target).Intersect(bounds))
		l.Redraw()
		l.editor.Redraw()
	})
}

func (g *GoCode) set(e Editor, l *suggestionList, pos int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if cancel, ok := g.cancels[e]; ok {
		cancel()
	}
	g.lists[e] = l
	ctx, cancel := context.WithCancel(context.Background())
	g.cancels[e] = cancel
	go g.show(ctx, l, pos)
}

func (g *GoCode) Moving(ie text.Editor, d caret.Direction, m caret.Mod, carets []int) (caret.Direction, caret.Mod, []int) {
	e := ie.(Editor)
	g.mu.RLock()
	defer g.mu.RUnlock()
	l, ok := g.lists[e]
	if !ok {
		return d, m, carets
	}
	if !l.Attached() {
		if cancel, ok := g.cancels[e]; ok {
			cancel()
		}
		return d, m, carets
	}
	if len(e.Carets()) > 1 || m != caret.NoMod || l.adapter.Len() == 0 {
		g.stop(e)
		return d, m, carets
	}
	switch d {
	case caret.Up:
		l.SelectPrevious()
	case caret.Down:
		l.SelectNext()
	case caret.NoDirection:
		g.stop(e)
		return d, m, carets
	default:
		return d, m, carets
	}
	return caret.NoDirection, caret.NoMod, nil
}

func (g *GoCode) Moved(ie text.Editor, carets []int) {
	e := ie.(Editor)
	g.mu.RLock()
	defer g.mu.RUnlock()
	l, ok := g.lists[e]
	if !ok {
		return
	}
	if !l.Attached() {
		return
	}
	if len(carets) != 1 {
		return
	}
	pos := carets[0]

	if cancel, ok := g.cancels[e]; ok {
		cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	g.cancels[e] = cancel
	e.RemoveChild(l)
	go g.show(ctx, l, pos)
}

func (g *GoCode) Cancel(ie text.Editor) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.stop(ie.(Editor))
}

func (g *GoCode) stop(e Editor) bool {
	if cancel, ok := g.cancels[e]; ok {
		cancel()
		delete(g.cancels, e)
	}
	if l, ok := g.lists[e]; ok {
		if e.Children().Find(l) != nil {
			e.RemoveChild(l)
		}
		delete(g.lists, e)
		return true
	}
	return false
}

func (g *GoCode) Confirm(ie text.Editor) bool {
	e := ie.(Editor)
	g.mu.Lock()
	defer g.mu.Unlock()
	l, ok := g.lists[e]
	if !ok {
		return false
	}
	g.stop(e)
	if l.adapter.Len() == 0 {
		return false
	}
	l.apply()
	return true
}
