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
	"github.com/nelsam/vidar/commander/input"
	"github.com/nelsam/vidar/commands/cursor"
	"github.com/nelsam/vidar/settings"
)

type Projecter interface {
	Project() settings.Project
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
	return []string{"cursor-movement", "input-handler"}
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
		log.Printf("cancelled or %d == 0", n)
		return
	}

	bounds := l.editor.Size().Rect().Contract(l.editor.Padding())
	line := l.editor.Line(l.editor.LineIndex(pos))
	lineOffset := gxui.ChildToParent(math.ZeroPoint, line, l.editor)
	target := line.PositionAt(pos).Add(lineOffset)
	cs := l.DesiredSize(math.ZeroSize, bounds.Size())

	g.driver.Call(func() {
		if ctxCancelled(ctx) {
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

func (g *GoCode) Moving(ie input.Editor, d cursor.Direction, cursors []int) cursor.Direction {
	e := ie.(Editor)
	g.mu.RLock()
	defer g.mu.RUnlock()
	l, ok := g.lists[e]
	if !ok {
		return d
	}
	if !l.Attached() {
		if cancel, ok := g.cancels[e]; ok {
			cancel()
		}
		return d
	}
	if len(cursors) != 1 {
		g.stop(e)
		return d
	}
	switch d {
	case cursor.Up:
		l.SelectPrevious()
	case cursor.Down:
		l.SelectNext()
	default:
		return d
	}
	return cursor.NoDirection
}

func (g *GoCode) Moved(ie input.Editor, d cursor.Direction, cursors []int) {
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
	if len(cursors) != 1 {
		return
	}
	pos := cursors[0]

	if cancel, ok := g.cancels[e]; ok {
		cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	g.cancels[e] = cancel
	e.RemoveChild(l)
	go g.show(ctx, l, pos)
}

func (g *GoCode) Cancel(ie input.Editor) bool {
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
		e.RemoveChild(l)
		delete(g.lists, e)
		return true
	}
	return false
}

func (g *GoCode) Confirm(ie input.Editor) bool {
	e := ie.(Editor)
	g.mu.Lock()
	defer g.mu.Unlock()
	l, ok := g.lists[e]
	if !ok {
		return false
	}
	l.apply()
	e.RemoveChild(l)
	delete(g.lists, e)
	return true
}
