// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package editor

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander/input"
	"github.com/nelsam/vidar/suggestions"
	"github.com/nelsam/vidar/theme"
)

type CodeEditor struct {
	mixins.CodeEditor
	adapter          *suggestions.Adapter
	suggestions      gxui.List
	suggestionsChild *gxui.Child

	theme       *basic.Theme
	syntaxTheme theme.Theme
	driver      gxui.Driver
	history     *History

	lastModified time.Time
	hasChanges   bool
	filepath     string

	watcher *fsnotify.Watcher

	selections      gxui.TextSelectionList
	scrollPositions math.Point

	renamed  bool
	onRename func(newPath string)
}

func (e *CodeEditor) Init(driver gxui.Driver, theme *basic.Theme, syntaxTheme theme.Theme, font gxui.Font, file, headerText string) {
	e.theme = theme
	e.syntaxTheme = syntaxTheme
	e.driver = driver
	e.history = NewHistory()

	// TODO: move to plugins
	e.adapter = &suggestions.Adapter{}
	e.suggestions = e.CreateSuggestionList()
	e.suggestions.SetAdapter(e.adapter)

	e.CodeEditor.Init(e, driver, theme, font)
	e.CodeEditor.SetScrollBarEnabled(true)
	e.SetDesiredWidth(math.MaxSize.W)
	e.watcherSetup()

	// TODO: move to plugins
	e.OnTextChanged(func(changes []gxui.TextBoxEdit) {
		e.hasChanges = true
		e.history.Add(changes...)
	})
	e.filepath = file
	e.open(headerText)

	e.SetTextColor(theme.TextBoxDefaultStyle.FontColor)
	e.SetMargin(math.Spacing{L: 3, T: 3, R: 3, B: 3})
	e.SetPadding(math.Spacing{L: 3, T: 3, R: 3, B: 3})
	e.SetBorderPen(gxui.TransparentPen)
}

func (e *CodeEditor) OnRename(callback func(newPath string)) {
	e.onRename = callback
}

func (e *CodeEditor) open(headerText string) {
	go e.watch()
	e.load(headerText)
}

func (e *CodeEditor) watcherSetup() {
	var err error
	e.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Error creating new fsnotify watcher: %s", err)
	}
}

func (e *CodeEditor) startWatch() error {
	err := e.watcher.Add(e.filepath)
	if os.IsNotExist(err) {
		err = e.waitForFileCreate()
		if err != nil {
			return err
		}
		return e.startWatch()
	}
	return nil
}

func (e *CodeEditor) watch() {
	if e.watcher == nil {
		return
	}
	err := e.startWatch()
	if err != nil {
		log.Printf("Error trying to watch %s for changes: %s", e.filepath, err)
		return
	}
	defer e.watcher.Remove(e.filepath)
	fileDir := filepath.Dir(e.filepath)
	err = e.watcher.Add(fileDir)
	if err != nil {
		log.Printf("Error trying to watch %s for changes: %s", fileDir, err)
		return
	}
	defer e.watcher.Remove(fileDir)
	err = e.inotifyWait(func(event fsnotify.Event) bool {
		if event.Name != e.filepath {
			if e.renamed && event.Op == fsnotify.Create {
				e.filepath = event.Name
				e.onRename(e.filepath)
				e.open("")
				return true
			}
			return false
		}
		switch event.Op {
		case fsnotify.Write:
			e.load("")
		case fsnotify.Rename:
			e.renamed = true
		case fsnotify.Remove:
			e.open("")
			return true
		}
		return false
	})
	if err != nil {
		log.Printf("Failed to wait on events for %s: %s", e.filepath, err)
		return
	}
}

func (e *CodeEditor) waitForFileCreate() error {
	dir := filepath.Dir(e.filepath)
	if err := os.MkdirAll(dir, 0750|os.ModeDir); err != nil {
		return err
	}
	if err := e.watcher.Add(dir); err != nil {
		return err
	}
	defer e.watcher.Remove(dir)

	return e.inotifyWait(func(event fsnotify.Event) bool {
		return event.Name == e.filepath && event.Op&fsnotify.Create == fsnotify.Create
	})
}

func (e *CodeEditor) inotifyWait(eventFunc func(fsnotify.Event) (done bool)) error {
	for {
		select {
		case event := <-e.watcher.Events:
			if eventFunc(event) {
				return nil
			}
		case err := <-e.watcher.Errors:
			return err
		}
	}
}

func (e *CodeEditor) load(headerText string) {
	f, err := os.Open(e.filepath)
	if os.IsNotExist(err) {
		e.driver.Call(func() {
			e.SetText(headerText)
		})
		return
	}
	if err != nil {
		log.Printf("Error opening file %s: %s", e.filepath, err)
		return
	}
	defer f.Close()
	finfo, err := f.Stat()
	if err != nil {
		log.Printf("Error stating file %s: %s", e.filepath, err)
		return
	}
	e.lastModified = finfo.ModTime()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Printf("Error reading file %s: %s", e.filepath, err)
		return
	}
	newText := string(b)
	if !strings.HasPrefix(newText, headerText) {
		log.Printf("%s: header text does not match requested header text", e.filepath)
	}
	e.driver.Call(func() {
		if e.Text() == newText {
			return
		}
		e.SetText(newText)
		e.restorePositions()
	})
}

func (e *CodeEditor) HasChanges() bool {
	return e.hasChanges
}

func (e *CodeEditor) LastKnownMTime() time.Time {
	return e.lastModified
}

func (e *CodeEditor) Filepath() string {
	return e.filepath
}

func (e *CodeEditor) FlushedChanges() {
	e.hasChanges = false
	e.lastModified = time.Now()
}

func (e *CodeEditor) Elements() []interface{} {
	return []interface{}{
		e.history,
		e.adapter,
		e.Controller(),
	}
}

func (e *CodeEditor) SetSyntaxLayers(layers []input.SyntaxLayer) {
	defer e.syntaxTheme.Rainbow.Reset()
	sort.Slice(layers, func(i, j int) bool {
		return layers[i].Construct < layers[j].Construct
	})
	gLayers := make(gxui.CodeSyntaxLayers, 0, len(layers))
	for _, l := range layers {
		highlight, found := e.syntaxTheme.Constructs[l.Construct]
		if !found {
			highlight = e.syntaxTheme.Rainbow.Next()
		}
		gLayer := gxui.CreateCodeSyntaxLayer()
		gLayer.SetColor(gxui.Color(highlight.Foreground))
		gLayer.SetBackgroundColor(gxui.Color(highlight.Background))
		for _, s := range l.Spans {
			gLayer.Add(s.Start, s.End-s.Start)
		}
		gLayers = append(gLayers, gLayer)
	}
	e.CodeEditor.SetSyntaxLayers(gLayers)
}

func (e *CodeEditor) Paint(c gxui.Canvas) {
	e.CodeEditor.Paint(c)

	if e.HasFocus() {
		r := e.Size().Rect()
		c.DrawRoundedRect(r, 3, 3, 3, 3, e.theme.FocusedStyle.Pen, e.theme.FocusedStyle.Brush)
	}
}

func (e *CodeEditor) CreateSuggestionList() gxui.List {
	l := e.theme.CreateList()
	l.SetBackgroundBrush(e.theme.CodeSuggestionListStyle.Brush)
	l.SetBorderPen(e.theme.CodeSuggestionListStyle.Pen)
	return l
}

func (e *CodeEditor) SetSuggestionProvider(provider gxui.CodeSuggestionProvider) {
	if e.SuggestionProvider() != provider {
		e.CodeEditor.SetSuggestionProvider(provider)
		if e.IsSuggestionListShowing() {
			e.updateSuggestionList()
		}
	}
}

func (e *CodeEditor) IsSuggestionListShowing() bool {
	return e.suggestionsChild != nil
}

func (e *CodeEditor) SortSuggestionList() {
	e.updateSuggestionList()
}

func (e *CodeEditor) ShowSuggestionList() {
	if e.SuggestionProvider() == nil || e.IsSuggestionListShowing() {
		return
	}
	e.showSuggestionList()
	e.updateSuggestionList()
}

func (e *CodeEditor) showSuggestionList() {
	e.suggestionsChild = e.AddChild(e.suggestions)
}

func (e *CodeEditor) updateSuggestionList() {
	caret := e.Controller().LastCaret()
	lineIdx := e.LineIndex(caret)
	text := e.Controller().Line(lineIdx)

	// TODO: This only skips suggestions on line comments, not block comments
	// (/* ... */).  Since block comments seem to be pretty uncommon in Go,
	// I'm not going to worry about it just yet.  Ideally, this would be solved
	// by having the syntax highlighting logic provide some context details to
	// the editor, so it knows some information about the context surrounding
	// the caret.
	if comment := strings.Index(text, "//"); comment != -1 && comment <= caret {
		e.HideSuggestionList()
		return
	}

	start, _ := e.Controller().WordAt(caret)
	suggestions := e.SuggestionProvider().SuggestionsAt(start)

	// TODO: if len(suggestions) == 1, show the completion in-line
	// instead of in a completion box.
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

	partial := e.Controller().TextRange(start, caret)
	e.adapter.SetSuggestions(suggestions)
	e.adapter.Sort(partial)
	if e.adapter.Len() == 0 {
		e.HideSuggestionList()
		return
	}
	e.suggestions.Select(e.suggestions.Adapter().ItemAt(0))

	// Position the suggestion list below the last caret
	bounds := e.Size().Rect().Contract(e.Padding())
	line := e.Line(lineIdx)
	lineOffset := gxui.ChildToParent(math.ZeroPoint, line, e)
	target := line.PositionAt(caret).Add(lineOffset)
	cs := e.suggestions.DesiredSize(math.ZeroSize, bounds.Size())
	e.suggestions.SetSize(cs)
	e.suggestionsChild.Layout(cs.Rect().Offset(target).Intersect(bounds))

	e.suggestions.Redraw()
	e.Redraw()
}

func (e *CodeEditor) HideSuggestionList() {
	if e.IsSuggestionListShowing() {
		e.RemoveChild(e.suggestions)
		e.Redraw()
		e.suggestionsChild = nil
	}
}

func (e *CodeEditor) storePositions() {
	e.selections = e.Controller().Selections()
	e.scrollPositions = math.Point{
		X: e.HorizOffset(),
		Y: e.ScrollOffset(),
	}
}

func (e *CodeEditor) restorePositions() {
	e.Controller().SetSelections(e.selections)
	e.SetHorizOffset(e.scrollPositions.X)
	e.SetScrollOffset(e.scrollPositions.Y)
}

func (e *CodeEditor) KeyPress(event gxui.KeyboardEvent) bool {
	defer e.storePositions()
	if event.Modifier != 0 && event.Modifier != gxui.ModShift {
		// ctrl-, alt-, and cmd-/super- keybindings should be dealt
		// with by the commands mapped to key bindings.
		return false
	}
	switch event.Key {
	case gxui.KeyPeriod:
		e.ShowSuggestionList()
		return true
	case gxui.KeyBackspace:
		result := e.TextBox.KeyPress(event)
		if e.IsSuggestionListShowing() {
			e.SortSuggestionList()
		}
		return result
	case gxui.KeyPageUp, gxui.KeyPageDown, gxui.KeyDelete:
		// These are all bindings that the TextBox handles fine.
		return e.TextBox.KeyPress(event)
	case gxui.KeyTab:
		// TODO: Gain knowledge about scope, so we know how much to indent.
		switch {
		case event.Modifier.Shift():
			e.Controller().UnindentSelection()
		default:
			e.Controller().IndentSelection()
		}
		return true
	case gxui.KeyUp, gxui.KeyDown:
		if e.IsSuggestionListShowing() {
			return e.suggestions.KeyPress(event)
		}
		return false
	case gxui.KeyLeft, gxui.KeyRight:
		if e.IsSuggestionListShowing() {
			e.HideSuggestionList()
		}
		return false
	case gxui.KeyEnter:
		controller := e.Controller()
		if e.IsSuggestionListShowing() {
			text := e.adapter.Suggestion(e.suggestions.Selected()).Code()
			start, end := controller.WordAt(controller.LastCaret())
			controller.SetSelection(gxui.CreateTextSelection(start, end, false))
			controller.ReplaceAll(text)
			controller.Deselect(false)
			e.HideSuggestionList()
			return true
		}
		// TODO: implement electric braces.  See
		// http://www.emacswiki.org/emacs/AutoPairs under
		// "Electric-RET".
		e.Controller().ReplaceWithNewlineKeepIndent()
		e.ScrollToRune(e.Controller().FirstCaret())
		return true
	case gxui.KeyEscape:
		if e.IsSuggestionListShowing() {
			e.HideSuggestionList()
			return true
		}
		// TODO: Keep track of some sort of concept of a "focused" caret and
		// focus that.
		e.Controller().SetCaret(e.Controller().FirstCaret())
	}
	return false
}

func (e *CodeEditor) KeyStroke(event gxui.KeyStrokeEvent) (consume bool) {
	if e.IsSuggestionListShowing() {
		e.driver.Call(e.SortSuggestionList)
	}
	return false
}

func (e *CodeEditor) CreateLine(theme gxui.Theme, index int) (mixins.TextBoxLine, gxui.Control) {
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
