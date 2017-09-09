// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package gocode

import (
	"context"
	"log"
	"unicode"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander/input"
	"github.com/nelsam/vidar/settings"
	"github.com/nelsam/vidar/suggestions"
)

type Editor interface {
	input.Editor
	gxui.Parent

	Size() math.Size
	Padding() math.Spacing
	LineIndex(caret int) int
	Line(idx int) mixins.TextBoxLine
	Filepath() string
	AddChild(gxui.Control) *gxui.Child
	RemoveChild(gxui.Control)
}

type Applier interface {
	Apply(input.Editor, ...input.Edit)
}

type suggestionList struct {
	mixins.List
	driver  gxui.Driver
	adapter *suggestions.Adapter
	font    gxui.Font
	project settings.Project
	editor  Editor
	ctrl    TextController
	applier Applier
}

func newSuggestionList(driver gxui.Driver, theme *basic.Theme, proj settings.Project, editor Editor, ctrl TextController, applier Applier) *suggestionList {
	s := &suggestionList{
		driver:  driver,
		adapter: &suggestions.Adapter{},
		font:    theme.DefaultMonospaceFont(),
		project: proj,
		editor:  editor,
		ctrl:    ctrl,
		applier: applier,
	}

	s.Init(s, theme)
	s.OnGainedFocus(s.Redraw)
	s.OnLostFocus(s.Redraw)
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

	start := pos
	for start > 0 && wordPart(runes[start-1]) {
		start--
	}
	if s.adapter.Pos() != start {
		if ctxCancelled(ctx) {
			return 0
		}
		suggestions := s.parseSuggestions(runes, start)
		s.adapter.Set(start, suggestions...)
	}
	if ctxCancelled(ctx) {
		return 0
	}

	s.driver.CallSync(func() {
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

func (s *suggestionList) parseSuggestions(runes []rune, start int) []suggestions.Suggestion {
	suggestions, err := suggestions.For(s.project.Environ(), s.editor.Filepath(), string(runes), start)
	if err != nil {
		log.Printf("Failed to load suggestions: %s", err)
		return nil
	}
	return suggestions
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

	s.applier.Apply(s.editor, input.Edit{
		At:  start,
		Old: s.ctrl.TextRunes()[start:end],
		New: []rune(suggestion.Name),
	})
}

func wordPart(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsNumber(r)
}
