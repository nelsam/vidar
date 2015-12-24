// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"log"
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
)

type EditorFinder interface {
	CurrentEditor() gxui.CodeEditor
}

type Find struct {
	driver  gxui.Driver
	theme   *basic.Theme
	editor  gxui.CodeEditor
	display gxui.Label
	pattern *findBox
}

func NewFinder(driver gxui.Driver, theme *basic.Theme) *Find {
	finder := &Find{}
	finder.Init(driver, theme)
	return finder
}

func (f *Find) Init(driver gxui.Driver, theme *basic.Theme) {
	f.driver = driver
	f.theme = theme
}

func (f *Find) Start(control gxui.Control) gxui.Control {
	f.editor = findEditor(control)
	f.display = f.theme.CreateLabel()
	f.display.SetText("find text")
	f.pattern = newFindBox(f.driver, f.theme, f.editor)
	f.pattern.OnTextChanged(func([]gxui.TextBoxEdit) {
		log.Printf("Text changed")
		f.editor.Controller().ClearSelections()
		needle := f.pattern.Text()
		if len(needle) == 0 {
			return
		}
		haystack := f.editor.Text()
		moveCursor := true
		log.Printf("Starting to search for selections")
		start := 0
		log.Printf("Length of text: %d", len(haystack))
		for next := strings.Index(haystack, needle); next != -1; next = strings.Index(haystack[start:], needle) {
			start += next
			log.Printf("Next selection at %d", start)
			selection := gxui.CreateTextSelection(start, start+len(needle), moveCursor)
			moveCursor = false
			f.editor.Controller().AddSelection(selection)
			start++
		}
	})
	log.Printf("Display and text set up")
	return f.display
}

func (f *Find) Name() string {
	return "find"
}

func (f *Find) Next() gxui.Focusable {
	return f.pattern
}

func (f *Find) Exec(target interface{}) (executed, consume bool) {
	return false, false
}

type findBox struct {
	mixins.TextBox
	editor gxui.CodeEditor
}

func newFindBox(driver gxui.Driver, theme *basic.Theme, editor gxui.CodeEditor) *findBox {
	box := &findBox{}
	box.Init(driver, theme, editor)
	return box
}

func (b *findBox) Init(driver gxui.Driver, theme *basic.Theme, editor gxui.CodeEditor) {
	b.TextBox.Init(b, driver, theme, theme.DefaultMonospaceFont())
	b.editor = editor
}

func (b *findBox) Complete(event gxui.KeyboardEvent) bool {
	log.Printf("Checking if event %+v is enter", event)
	if event.Key == gxui.KeyEnter {
		log.Printf("Key is enter")
		selections := b.editor.Controller().Selections()
		caret := b.editor.Controller().FirstCaret()
		var next gxui.TextSelection
		for _, next = range selections {
			if next.Start() > caret {
				break
			}
		}
		log.Printf("Found next text selection at %+v, moving to start", next)
		b.editor.Controller().SetCaret(next.Start())
	}
	return false
}

func findEditor(control gxui.Control) gxui.CodeEditor {
	switch src := control.(type) {
	case EditorFinder:
		return src.CurrentEditor()
	case gxui.Parent:
		for _, child := range src.Children() {
			if editor := findEditor(child.Control); editor != nil {
				return editor
			}
		}
	}
	return nil
}
