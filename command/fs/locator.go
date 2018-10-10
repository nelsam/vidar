// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package fs

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/scoring"
	"github.com/nelsam/vidar/setting"
)

const minInputChars = 10

var (
	dirColor = gxui.Color{
		R: 0.1,
		G: 0.3,
		B: 0.8,
		A: 1,
	}
	completionBG = gxui.Color{
		R: 1,
		G: 1,
		B: 1,
		A: 0.05,
	}
	completionPadding = math.Size{
		H: 7,
		W: 7,
	}
)

// FileGetter is used to get the currently open file.
//
// TODO: replace this with hooks on opening a file.
type FileGetter interface {
	CurrentFile() string
}

// Projecter is used to get the currently open project.
//
// TODO: replace this with hooks on opening a project.
type Projecter interface {
	Project() setting.Project
}

// Locator is a type of UI element which prompts the user for a file path.  It has
// completion features to help with locating existing files and folders.
type Locator struct {
	mixins.LinearLayout

	lock sync.RWMutex

	theme       *basic.Theme
	driver      gxui.Driver
	dir         *dirLabel
	file        *fileBox
	completions []gxui.Label
	files       []string
}

func NewLocator(driver gxui.Driver, theme *basic.Theme) *Locator {
	f := &Locator{}
	f.Init(driver, theme)
	return f
}

func (f *Locator) Init(driver gxui.Driver, theme *basic.Theme) {
	f.LinearLayout.Init(f, theme)
	f.theme = theme
	f.driver = driver

	f.SetDirection(gxui.LeftToRight)
	f.dir = newDirLabel(driver, theme)
	f.AddChild(f.dir)
	f.file = newFileBox(driver, theme, f)
	f.AddChild(f.file)
	f.loadDirContents()
}

func (f *Locator) LoadDir(control gxui.Control) {
	startingPath := findStart(control)

	f.driver.Call(func() {
		defer f.loadDirContents()
		f.dir.SetText(startingPath)
		f.file.SetText("")
	})
}

func (f *Locator) Path() string {
	return filepath.Join(f.dir.Text(), f.file.Text())
}

func (f *Locator) SetPath(filePath string) {
	defer f.loadDirContents()
	dir, file := filepath.Split(filePath)

	f.dir.SetText(dir)
	f.file.SetText(file)
}

func (f *Locator) KeyPress(event gxui.KeyboardEvent) bool {
	return f.file.KeyPress(event)
}

func (f *Locator) KeyDown(event gxui.KeyboardEvent) {
	f.file.KeyDown(event)
}

func (f *Locator) KeyUp(event gxui.KeyboardEvent) {
	f.file.KeyUp(event)
}

func (f *Locator) KeyStroke(event gxui.KeyStrokeEvent) bool {
	return f.file.KeyStroke(event)
}

func (f *Locator) KeyRepeat(event gxui.KeyboardEvent) {
	f.file.KeyRepeat(event)
}

func (f *Locator) Paint(c gxui.Canvas) {
	f.LinearLayout.Paint(c)

	if f.HasFocus() {
		r := f.Size().Rect()
		s := f.theme.FocusedStyle
		c.DrawRoundedRect(r, 3, 3, 3, 3, s.Pen, s.Brush)
	}
}

func (f *Locator) IsFocusable() bool {
	return f.file.IsFocusable()
}

func (f *Locator) HasFocus() bool {
	return f.file.HasFocus()
}

func (f *Locator) GainedFocus() {
	f.file.GainedFocus()
}

func (f *Locator) LostFocus() {
	f.file.LostFocus()
}

func (f *Locator) OnGainedFocus(callback func()) gxui.EventSubscription {
	return f.file.OnGainedFocus(callback)
}

func (f *Locator) OnLostFocus(callback func()) gxui.EventSubscription {
	return f.file.OnLostFocus(callback)
}

func (f *Locator) updateCompletions() {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.clearCompletions(f.completions)

	f.completions = nil
	newCompletions := scoring.Sort(f.files, f.file.Text())

	for _, comp := range newCompletions {
		color := f.theme.LabelStyle.FontColor
		if strings.HasSuffix(comp, "/") {
			color = dirColor
		}
		l := newCompletionLabel(f.driver, f.theme, color)
		l.text = comp
		f.completions = append(f.completions, l)
	}

	f.addCompletions(f.completions)
}

func (f *Locator) clearCompletions(completions []gxui.Label) {
	cloned := make([]gxui.Label, 0, len(completions))
	cloned = append(cloned, completions...)
	f.driver.Call(func() {
		for _, l := range cloned {
			f.RemoveChild(l)
		}
	})
}

func (f *Locator) addCompletions(completions []gxui.Label) {
	cloned := make([]gxui.Label, 0, len(completions))
	cloned = append(cloned, completions...)
	f.driver.Call(func() {
		for _, l := range cloned {
			f.AddChild(l)
			l.SetHorizontalAlignment(gxui.AlignCenter)
		}
	})
}

func (f *Locator) loadDirContents() {
	f.lock.Lock()
	defer func() {
		f.lock.Unlock()
		f.updateCompletions()
	}()
	f.files = nil
	dir := f.dir.Text()
	if dir == "" {
		// Should be Windows-only.  The drive hasn't been chosen yet.
		return
	}
	contents, err := ioutil.ReadDir(dir)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		log.Printf("Unexpected error trying to read directory %s: %s", f.dir.Text(), err)
		return
	}
	for _, finfo := range contents {
		name := finfo.Name()
		if finfo.IsDir() {
			name += "/"
		}
		f.files = append(f.files, name)
	}
}

type dirLabel struct {
	mixins.Label

	driver gxui.Driver
}

func newDirLabel(driver gxui.Driver, theme *basic.Theme) *dirLabel {
	label := &dirLabel{driver: driver}
	label.Label.Init(label, theme, theme.DefaultMonospaceFont(), theme.LabelStyle.FontColor)
	label.SetMargin(math.Spacing{L: 3, T: 3, R: 3, B: 3})
	return label
}

func (l *dirLabel) SetText(dir string) {
	if len(dir) == 0 {
		return
	}
	if dir[len(dir)-1] != filepath.Separator {
		dir += string(filepath.Separator)
	}
	l.Label.SetText(dir)
}

func (l *dirLabel) Text() string {
	text := l.Label.Text()
	if root, ok := fsroot(text); ok {
		return root
	}
	if text == "" {
		log.Printf("This is odd.  We have an empty root that isn't considered a drive root.")
		return ""
	}
	if text[len(text)-1] == filepath.Separator {
		text = text[:len(text)-1]
	}
	return text
}

type fileBox struct {
	mixins.TextBox

	locator *Locator
	font    gxui.Font
}

func newFileBox(driver gxui.Driver, theme *basic.Theme, l *Locator) *fileBox {
	file := &fileBox{
		locator: l,
		font:    theme.DefaultMonospaceFont(),
	}
	file.TextBox.Init(file, driver, theme, theme.DefaultMonospaceFont())
	file.SetTextColor(theme.TextBoxDefaultStyle.FontColor)
	file.SetMargin(math.Spacing{L: 3, T: 3, R: 3, B: 3})
	file.SetPadding(math.Spacing{L: 3, T: 3, R: 3, B: 3})
	file.SetBackgroundBrush(theme.TextBoxDefaultStyle.Brush)
	file.SetDesiredWidth(math.MaxSize.W)
	file.SetMultiline(false)
	return file
}

func (f *fileBox) KeyPress(event gxui.KeyboardEvent) bool {
	l := f.locator
	if event.Modifier == 0 {
		l.lock.RLock()
		defer l.lock.RUnlock()

		switch event.Key {
		case gxui.KeyEscape:
			if len(l.completions) > 0 {
				l.clearCompletions(l.completions)
				l.completions = nil
				return true
			}
		case gxui.KeySlash:
			fullPath := l.Path()
			if len(l.completions) > 0 {
				fullPath = filepath.Join(l.dir.Text(), l.completions[0].Text())
			}
			l.dir.SetText(fullPath)
			l.file.setFile("")
			go l.loadDirContents()
			return true
		case gxui.KeyEnter:
			if len(l.completions) == 0 {
				return false
			}
			fullPath := filepath.Join(l.dir.Text(), l.completions[0].Text())
			finfo, err := os.Stat(fullPath)
			if os.IsNotExist(err) {
				return false
			}
			if err != nil {
				return false
			}
			if !finfo.IsDir() {
				l.file.setFile(finfo.Name())
				return false
			}
			l.dir.SetText(fullPath)
			l.file.setFile("")
			go l.loadDirContents()
			return true
		case gxui.KeyRight:
			l.clearCompletions(l.completions)
			for i := 1; i < len(l.completions); i++ {
				l.completions[i-1], l.completions[i] = l.completions[i], l.completions[i-1]
			}
			l.addCompletions(l.completions)
			return true
		case gxui.KeyLeft:
			l.clearCompletions(l.completions)
			for i := len(l.completions) - 1; i > 0; i-- {
				l.completions[i-1], l.completions[i] = l.completions[i], l.completions[i-1]
			}
			l.addCompletions(l.completions)
			return true
		case gxui.KeyBackspace:
			if len(l.file.Text()) == 0 {
				newDir := filepath.Dir(l.dir.Text())
				if newDir == l.dir.Text() {
					newDir = systemRoot
				}
				l.dir.SetText(newDir)
				go l.loadDirContents()
				return true
			}
		}
	}
	return f.TextBox.KeyPress(event)
}

func (f *fileBox) KeyStroke(event gxui.KeyStrokeEvent) bool {
	defer f.locator.updateCompletions()
	if event.Character == filepath.Separator {
		return false
	}
	return f.TextBox.KeyStroke(event)
}

func (f *fileBox) setFile(file string) {
	f.SetText(file)
	f.Controller().SetCaret(len(file))
}

func (f *fileBox) DesiredSize(min, max math.Size) math.Size {
	s := f.TextBox.DesiredSize(min, max)
	chars := len(f.Text())
	if chars < minInputChars {
		chars = minInputChars
	}
	width := chars * f.font.GlyphMaxSize().W
	if width > max.W {
		width = max.W
	}
	if width < min.W {
		width = min.W
	}
	s.W = width
	return s
}

type completionLabel struct {
	mixins.Label

	padding math.Size

	// text should be used for setting the label's text
	// outside of the UI thread.  The next time Paint is
	// called, SetText will be passed this value.
	text string
}

func newCompletionLabel(driver gxui.Driver, theme gxui.Theme, color gxui.Color) *completionLabel {
	l := &completionLabel{}
	l.Init(l, theme, theme.DefaultMonospaceFont(), color)
	l.SetMargin(math.Spacing{
		T: 3,
		R: 3,
	})
	l.padding = completionPadding
	return l
}

func (l *completionLabel) DesiredSize(min, max math.Size) math.Size {
	size := l.Label.DesiredSize(min, max)
	size.W += l.padding.W
	size.H += l.padding.H
	return size
}

func (l *completionLabel) SetText(text string) {
	l.text = text
	l.Label.SetText(text)
}

func (l *completionLabel) Paint(c gxui.Canvas) {
	if l.Text() != l.text {
		l.SetText(l.text)
	}
	l.Label.Paint(c)
	r := l.Size().Rect()
	c.DrawRoundedRect(r, 3, 3, 3, 3, gxui.TransparentPen, gxui.CreateBrush(completionBG))
}

// Elementer is used to find child elements of an element.
type Elementer interface {
	Elements() []interface{}
}

func findStart(control gxui.Control) string {
	if startingPath := findCurrentFile(control); startingPath != "" {
		return filepath.Dir(startingPath)
	}
	if project, ok := findProject(control); ok {
		return project.Path
	}
	return setting.DefaultProject.Path
}

func findCurrentFile(control gxui.Control) string {
	switch src := control.(type) {
	case FileGetter:
		return src.CurrentFile()
	case gxui.Parent:
		for _, child := range src.Children() {
			if file := findCurrentFile(child.Control); file != "" {
				return file
			}
		}
	}
	return ""
}

func findProject(e interface{}) (setting.Project, bool) {
	switch src := e.(type) {
	case Projecter:
		return src.Project(), true
	case Elementer:
		for _, elem := range src.Elements() {
			if proj, ok := findProject(elem); ok {
				return proj, true
			}
		}
	}
	return setting.Project{}, false
}
