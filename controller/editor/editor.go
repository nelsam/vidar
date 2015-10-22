// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package editor

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/gxui_playground/suggestions"
	"github.com/nelsam/gxui_playground/syntax"
)

type Editor struct {
	mixins.PanelHolder

	editors map[string]*editor

	driver gxui.Driver
	theme  *basic.Theme
	font   gxui.Font
}

func New(driver gxui.Driver, theme *basic.Theme, font gxui.Font) *Editor {
	e := &Editor{
		editors: make(map[string]*editor),
		driver:  driver,
		theme:   theme,
		font:    font,
	}
	e.Init(e, theme)
	e.SetMargin(math.Spacing{L: 0, T: 2, R: 0, B: 0})
	return e
}

func (e *Editor) Open(file string) {
	if editor, ok := e.editors[file]; ok {
		e.Select(e.PanelIndex(editor))
		e.Focus()
		return
	}
	editor := new(editor)
	editor.Init(e.driver, e.theme, e.font, file)
	editor.SetTabWidth(8)
	suggester := suggestions.NewGoCodeProvider(editor).(*suggestions.GoCodeProvider)
	editor.SetSuggestionProvider(suggester)

	e.editors[file] = editor
	e.AddPanel(editor, file)
	e.Select(e.PanelIndex(editor))
	e.Focus()
}

func (e *Editor) Focus() {
	gxui.SetFocus(e.SelectedPanel().(gxui.Focusable))
}

func (e *Editor) Files() []string {
	files := make([]string, 0, len(e.editors))
	for file := range e.editors {
		files = append(files, file)
	}
	return files
}

func (e *Editor) CreatePanelTab() mixins.PanelTab {
	return basic.CreatePanelTab(e.theme)
}

func (e *Editor) KeyPress(event gxui.KeyboardEvent) bool {
	if event.Modifier.Control() || event.Modifier.Super() {
		switch event.Key {
		case gxui.KeyTab:
			panels := e.PanelCount()
			if panels < 2 {
				return true
			}
			current := e.PanelIndex(e.SelectedPanel())
			next := current + 1
			if event.Modifier.Shift() {
				next = current - 1
			}
			if next >= panels {
				next = 0
			}
			if next < 0 {
				next = panels - 1
			}
			e.Select(next)
			return true
		}
	}
	return e.PanelHolder.KeyPress(event)
}

func (e *Editor) CurrentFile() string {
	return e.SelectedPanel().(*editor).filepath
}

// editor is an implementation of gxui.CodeEditor, based on
// gxui/mixins.CodeEditor.
type editor struct {
	mixins.CodeEditor
	adapter     *suggestions.Adapter
	suggestions gxui.List
	theme       *basic.Theme

	lastModified time.Time
	hasChanges   bool
	filepath     string
}

func (e *editor) Init(driver gxui.Driver, theme *basic.Theme, font gxui.Font, file string) {
	e.theme = theme

	e.adapter = &suggestions.Adapter{}
	e.suggestions = e.CreateSuggestionList()
	e.suggestions.SetAdapter(e.adapter)

	e.CodeEditor.Init(e, driver, theme, font)
	e.SetDesiredWidth(math.MaxSize.W)

	e.OnTextChanged(func(changes []gxui.TextBoxEdit) {
		e.hasChanges = true
		// TODO: only update layers that changed.
		newLayers, err := syntax.Layers(e.filepath, e.Text())
		e.SetSyntaxLayers(newLayers)
		// TODO: display the error in some pane of the editor
		_ = err
	})
	e.filepath = file
	e.open()

	e.SetTextColor(theme.TextBoxDefaultStyle.FontColor)
	e.SetMargin(math.Spacing{L: 3, T: 3, R: 3, B: 3})
	e.SetPadding(math.Spacing{L: 3, T: 3, R: 3, B: 3})
	e.SetBorderPen(gxui.TransparentPen)
}

func (e *editor) open() {
	if e.filepath == "" {
		e.SetText(`// Scratch
// This buffer is for jotting down quick notes, but is not saved to disk.
// Use at your own risk!`)
		return
	}
	f, err := os.Open(e.filepath)
	if os.IsNotExist(err) {
		e.SetText("")
		return
	}
	if err != nil {
		panic(err)
	}
	defer f.Close()
	finfo, err := f.Stat()
	if err != nil {
		panic(err)
	}
	e.lastModified = finfo.ModTime()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	e.SetText(string(b))
}

// mixins.CodeEditor overrides
func (e *editor) Paint(c gxui.Canvas) {
	e.CodeEditor.Paint(c)

	if e.HasFocus() {
		r := e.Size().Rect()
		c.DrawRoundedRect(r, 3, 3, 3, 3, e.theme.FocusedStyle.Pen, e.theme.FocusedStyle.Brush)
	}
}

func (e *editor) CreateSuggestionList() gxui.List {
	l := e.theme.CreateList()
	l.SetBackgroundBrush(e.theme.CodeSuggestionListStyle.Brush)
	l.SetBorderPen(e.theme.CodeSuggestionListStyle.Pen)
	return l
}

func (e *editor) SetSuggestionProvider(provider gxui.CodeSuggestionProvider) {
	if e.SuggestionProvider() != provider {
		e.CodeEditor.SetSuggestionProvider(provider)
		if e.IsSuggestionListShowing() {
			e.updateSuggestionList()
		}
	}
}

func (e *editor) IsSuggestionListShowing() bool {
	return e.Children().Find(e.suggestions) != nil
}

func (e *editor) SortSuggestionList() {
	caret := e.Controller().LastCaret()
	partial := e.WordAt(caret)
	e.adapter.Sort(partial)
}

func (e *editor) ShowSuggestionList() {
	if e.SuggestionProvider() == nil || e.IsSuggestionListShowing() {
		return
	}
	e.updateSuggestionList()
}

func (e *editor) updateSuggestionList() {
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
	lineOffset := gxui.ChildToParent(math.ZeroPoint, line, e)
	target := line.PositionAt(caret).Add(lineOffset)
	cs := e.suggestions.DesiredSize(math.ZeroSize, bounds.Size())
	e.suggestions.Select(e.suggestions.Adapter().ItemAt(0))
	e.suggestions.SetSize(cs)
	child.Layout(cs.Rect().Offset(target).Intersect(bounds))
}

func (e *editor) HideSuggestionList() {
	if e.IsSuggestionListShowing() {
		e.RemoveChild(e.suggestions)
	}
}

func (e *editor) KeyPress(event gxui.KeyboardEvent) bool {
	if event.Modifier.Control() || event.Modifier.Super() {
		switch event.Key {
		case gxui.KeySpace:
			e.ShowSuggestionList()
			return true
		case gxui.KeyS:
			if !e.lastModified.IsZero() {
				finfo, err := os.Stat(e.filepath)
				if err != nil {
					panic(err)
				}
				if finfo.ModTime().After(e.lastModified) {
					panic("Cannot save file: written since last open")
				}
			}
			f, err := os.Create(e.filepath)
			if err != nil {
				panic(err)
			}
			if !strings.HasSuffix(e.Text(), "\n") {
				e.SetText(e.Text() + "\n")
			}
			if _, err := f.WriteString(e.Text()); err != nil {
				panic(err)
			}
			finfo, err := f.Stat()
			if err != nil {
				panic(err)
			}
			e.lastModified = finfo.ModTime()
			f.Close()
			e.hasChanges = false
			return true
		case gxui.KeyTab:
			return false
		}
	}
	switch event.Key {
	case gxui.KeyTab:
		// TODO: Gain knowledge about scope, so we know how much to indent.
		switch {
		case event.Modifier.Shift():
			e.Controller().UnindentSelection(e.TabWidth())
		default:
			e.Controller().IndentSelection(e.TabWidth())
		}
		return true
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

func (e *editor) KeyStroke(event gxui.KeyStrokeEvent) (consume bool) {
	consume = e.CodeEditor.KeyStroke(event)
	if e.IsSuggestionListShowing() {
		e.SortSuggestionList()
	}
	return
}

func (e *editor) CreateLine(theme gxui.Theme, index int) (mixins.TextBoxLine, gxui.Control) {
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
