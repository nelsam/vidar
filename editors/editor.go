// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package editors

import (
	"fmt"
	"time"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/gxui_playground/suggestions"
)

// Editor is an implementation of gxui.CodeEditor, based on
// gxui/mixins.CodeEditor.
type Editor struct {
	mixins.CodeEditor
	adapter     *suggestions.Adapter
	suggestions gxui.List
	theme       *basic.Theme
	outer       mixins.CodeEditorOuter

	LastModified time.Time
	HasChanges   bool
	Filepath     string
}

func New(driver gxui.Driver, theme *basic.Theme, font gxui.Font) gxui.CodeEditor {
	e := new(Editor)
	e.Init(e, driver, theme, font)
	return e
}

func (e *Editor) Init(outer mixins.CodeEditorOuter, driver gxui.Driver, theme *basic.Theme, font gxui.Font) {
	e.outer = outer
	e.theme = theme

	e.adapter = &suggestions.Adapter{}
	e.suggestions = e.outer.CreateSuggestionList()
	e.suggestions.SetAdapter(e.adapter)

	e.CodeEditor.Init(outer, driver, theme, font)

	e.SetTextColor(theme.TextBoxDefaultStyle.FontColor)
	e.SetMargin(math.Spacing{L: 3, T: 3, R: 3, B: 3})
	e.SetPadding(math.Spacing{L: 3, T: 3, R: 3, B: 3})
	e.SetBorderPen(gxui.TransparentPen)
}

// mixins.CodeEditor overrides
func (e *Editor) Paint(c gxui.Canvas) {
	e.CodeEditor.Paint(c)

	if e.HasFocus() {
		r := e.Size().Rect()
		c.DrawRoundedRect(r, 3, 3, 3, 3, e.theme.FocusedStyle.Pen, e.theme.FocusedStyle.Brush)
	}
}

func (e *Editor) CreateSuggestionList() gxui.List {
	l := e.theme.CreateList()
	l.SetBackgroundBrush(e.theme.CodeSuggestionListStyle.Brush)
	l.SetBorderPen(e.theme.CodeSuggestionListStyle.Pen)
	return l
}

func (e *Editor) SetSuggestionProvider(provider gxui.CodeSuggestionProvider) {
	if e.SuggestionProvider() != provider {
		e.CodeEditor.SetSuggestionProvider(provider)
		if e.IsSuggestionListShowing() {
			e.updateSuggestionList()
		}
	}
}

func (e *Editor) IsSuggestionListShowing() bool {
	return e.outer.Children().Find(e.suggestions) != nil
}

func (e *Editor) SortSuggestionList() {
	caret := e.Controller().LastCaret()
	partial := e.WordAt(caret)
	e.adapter.Sort(partial)
}

func (e *Editor) ShowSuggestionList() {
	if e.SuggestionProvider() == nil || e.IsSuggestionListShowing() {
		return
	}
	e.updateSuggestionList()
}

func (e *Editor) updateSuggestionList() {
	caret := e.Controller().LastCaret()

	suggestions := e.SuggestionProvider().SuggestionsAt(caret)
	if len(suggestions) == 0 {
		// TODO: if len(suggestions) == 1, show the completion in-line
		// instead of in a completion box.
		e.HideSuggestionList()
		return
	}
	longest := 0
	for _, suggestion := range suggestions {
		suggestionText := suggestion.(fmt.Stringer).String()
		if len(suggestionText) > longest {
			longest = len(suggestionText)
		}
	}
	size := e.Font().GlyphMaxSize()
	size.W *= longest
	e.adapter.SetSize(size)

	e.adapter.SetSuggestions(suggestions)
	e.SortSuggestionList()
	child := e.AddChild(e.suggestions)

	// Position the suggestion list below the last caret
	lineIdx := e.LineIndex(caret)
	// TODO: What if the last caret is not visible?
	bounds := e.Size().Rect().Contract(e.Padding())
	line := e.Line(lineIdx)
	lineOffset := gxui.ChildToParent(math.ZeroPoint, line, e.outer)
	target := line.PositionAt(caret).Add(lineOffset)
	cs := e.suggestions.DesiredSize(math.ZeroSize, bounds.Size())
	e.suggestions.Select(e.suggestions.Adapter().ItemAt(0))
	e.suggestions.SetSize(cs)
	child.Layout(cs.Rect().Offset(target).Intersect(bounds))
}

func (e *Editor) HideSuggestionList() {
	if e.IsSuggestionListShowing() {
		e.RemoveChild(e.suggestions)
	}
}

func (e *Editor) KeyPress(event gxui.KeyboardEvent) bool {
	switch event.Key {
	case gxui.KeyTab:
		// TODO: implement emacs-style indent.
	case gxui.KeySpace:
		if event.Modifier.Control() {
			e.ShowSuggestionList()
			return true
		}
	case gxui.KeyUp:
		fallthrough
	case gxui.KeyDown:
		if e.IsSuggestionListShowing() {
			return e.suggestions.KeyPress(event)
		}
	case gxui.KeyLeft:
		e.HideSuggestionList()
	case gxui.KeyRight:
		e.HideSuggestionList()
	case gxui.KeyEnter:
		controller := e.Controller()
		if e.IsSuggestionListShowing() {
			text := e.adapter.Suggestion(e.suggestions.Selected()).Code()
			start, end := controller.WordAt(controller.LastCaret())
			controller.SetSelection(gxui.CreateTextSelection(start, end, false))
			controller.ReplaceAll(text)
			controller.Deselect(false)
			e.HideSuggestionList()
		} else {
			// TODO: implement electric braces.  See
			// http://www.emacswiki.org/emacs/AutoPairs under
			// "Electric-RET".
			e.Controller().ReplaceWithNewlineKeepIndent()
		}
		return true
	case gxui.KeyEscape:
		if e.IsSuggestionListShowing() {
			e.HideSuggestionList()
			return true
		}
	}
	return e.CodeEditor.KeyPress(event)
}

func (e *Editor) KeyStroke(event gxui.KeyStrokeEvent) (consume bool) {
	consume = e.CodeEditor.KeyStroke(event)
	if e.IsSuggestionListShowing() {
		e.SortSuggestionList()
	}
	return
}

func (e *Editor) CreateLine(theme gxui.Theme, index int) (mixins.TextBoxLine, gxui.Control) {
	lineNumber := theme.CreateLabel()
	lineNumber.SetText(fmt.Sprintf("%4d", index+1))

	line := &mixins.CodeEditorLine{}
	line.Init(line, theme, &e.CodeEditor, index)

	layout := theme.CreateLinearLayout()
	layout.SetDirection(gxui.LeftToRight)
	layout.AddChild(lineNumber)
	layout.AddChild(line)

	return line, layout
}
