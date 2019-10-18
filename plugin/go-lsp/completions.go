// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"unicode"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/command/caret"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/input"
	"github.com/nelsam/vidar/plugin/status"
	"github.com/nelsam/vidar/setting"
)

type Applier interface {
	Apply(input.Editor, ...input.Edit)
}

type Complete struct {
	status.General

	e input.Editor
}

func (c *Complete) Name() string {
	return "show-suggestions"
}

func (c *Complete) Menu() string {
	return "Golang"
}

func (c *Complete) Defaults() []fmt.Stringer {
	return []fmt.Stringer{gxui.KeyboardEvent{
		Modifier: gxui.ModControl,
		Key:      gxui.KeySpace,
	}}
}

func (c *Complete) Reset() {
	c.General = status.General{}
}

func (c *Complete) Store(v interface{}) bind.Status {

	return bind.Waiting
}

func (c *Complete) Exec() error {}

// gocode

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

	mu      sync.Mutex
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
			// TODO: Add a UI message.
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

func (g *GoCode) Moving(ie input.Editor, d caret.Direction, m caret.Mod, carets []int) (caret.Direction, caret.Mod, []int) {
	e := ie.(Editor)
	g.mu.Lock()
	defer g.mu.Unlock()
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
		return caret.NoDirection, caret.NoMod, nil
	case caret.Down:
		l.SelectNext()
		return caret.NoDirection, caret.NoMod, nil
	case caret.NoDirection:
		g.stop(e)
	}
	return d, m, carets
}

func (g *GoCode) Moved(ie input.Editor, carets []int) {
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
		if e.Children().Find(l) != nil {
			e.RemoveChild(l)
		}
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
	g.stop(e)
	if l.adapter.Len() == 0 {
		return false
	}
	l.apply()
	return true
}

// list

type Editor interface {
	input.Editor
	gxui.Focusable
	gxui.Parent

	Carets() []int
	Size() math.Size
	Padding() math.Spacing
	LineIndex(caret int) int
	Line(idx int) mixins.TextBoxLine
	AddChild(gxui.Control) *gxui.Child
	RemoveChild(gxui.Control)
}

type suggestionList struct {
	mixins.List
	driver  gxui.Driver
	adapter *suggestions.Adapter
	font    gxui.Font
	project setting.Project
	editor  Editor
	ctrl    TextController
	applier Applier
	gocode  *GoCode
}

func newSuggestionList(driver gxui.Driver, theme *basic.Theme, proj setting.Project, editor Editor, ctrl TextController, applier Applier, gocode *GoCode) *suggestionList {
	s := &suggestionList{
		driver:  driver,
		adapter: &suggestions.Adapter{},
		font:    theme.DefaultMonospaceFont(),
		project: proj,
		editor:  editor,
		ctrl:    ctrl,
		applier: applier,
		gocode:  gocode,
	}

	s.Init(s, theme)
	s.OnGainedFocus(func() { gxui.Focus(editor) })
	s.OnClick(func(gxui.MouseEvent) {
		s.driver.CallSync(func() {
			s.gocode.Confirm(s.editor)
		})
	})

	s.SetPadding(math.CreateSpacing(2))
	s.SetBackgroundBrush(theme.CodeSuggestionListStyle.Brush)
	s.SetBorderPen(theme.CodeSuggestionListStyle.Pen)

	s.SetAdapter(s.adapter)
	return s
}

func (s *suggestionList) show(ctx context.Context, pos int) int {
	runes := s.ctrl.TextRunes()
	if pos >= len(runes) {
		log.Printf("Warning: suggestion list sees a pos of %d while the rune length of the editor is %d", pos, len(runes))
		return 0
	}

	// load suggestions from lsp

	s.driver.CallSync(func() {
		// TODO: This doesn't always create a large enough box; it needs to be fixed.
		longest := s.adapter.Sort(runes[start:pos])
		if s.adapter.Len() == 0 {
			return
		}
		s.Select(s.adapter.ItemAt(0))

		size := s.font.GlyphMaxSize()
		size.W *= longest
		s.adapter.SetSize(size)
	})
	return s.adapter.Len()
}

func (s *suggestionList) apply() {
	suggestion := s.Selected().(suggestions.Suggestion)
	start := s.adapter.Pos()
	carets := s.ctrl.Carets()
	if len(carets) != 1 {
		log.Printf("Cannot apply completion to more than one caret; got %d", len(carets))
		return
	}
	end := carets[0]
	runes := s.ctrl.TextRunes()

	if start > end {
		return
	}
	go s.applier.Apply(s.editor, input.Edit{
		At:  start,
		Old: runes[start:end],
		New: []rune(suggestion.Name),
	})
}

func wordPart(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsNumber(r)
}
