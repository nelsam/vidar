// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

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
	"github.com/nelsam/vidar/settings"
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

type FileGetter interface {
	CurrentFile() string
}

type ProjectGetter interface {
	CurrentProject() settings.Project
}

type FSLocator struct {
	mixins.LinearLayout

	lock sync.RWMutex

	theme       *basic.Theme
	driver      gxui.Driver
	dir         *dirLabel
	file        *fileBox
	completions []gxui.Label
	files       []string
}

func NewFSLocator(driver gxui.Driver, theme *basic.Theme) *FSLocator {
	f := &FSLocator{}
	f.Init(driver, theme)
	return f
}

func (f *FSLocator) Init(driver gxui.Driver, theme *basic.Theme) {
	f.LinearLayout.Init(f, theme)
	f.theme = theme
	f.driver = driver

	f.SetDirection(gxui.LeftToRight)
	f.dir = newDirLabel(driver, theme)
	f.AddChild(f.dir)
	f.file = newFileBox(driver, theme)
	f.AddChild(f.file)
	f.loadDirContents()
}

func (f *FSLocator) loadEditorDir(control gxui.Control) {
	startingPath := findStart(control)

	f.driver.Call(func() {
		defer f.loadDirContents()
		f.dir.SetText(startingPath)
		f.file.SetText("")
	})
}

func (f *FSLocator) Path() string {
	return filepath.Join(f.dir.Text(), f.file.Text())
}

func (f *FSLocator) SetPath(filePath string) {
	defer f.loadDirContents()
	dir, file := filepath.Split(filePath)

	f.dir.SetText(dir)
	f.file.SetText(file)
}

func (f *FSLocator) KeyPress(event gxui.KeyboardEvent) bool {
	if event.Modifier == 0 {
		f.lock.RLock()
		defer f.lock.RUnlock()

		switch event.Key {
		case gxui.KeyEscape:
			if len(f.completions) > 0 {
				f.clearCompletions(f.completions)
				f.completions = nil
				return true
			}
		case gxui.KeySlash:
			fullPath := f.Path()
			if len(f.completions) > 0 {
				fullPath = filepath.Join(f.dir.Text(), f.completions[0].Text())
			}
			f.dir.SetText(fullPath)
			f.file.setFile("")
			go f.loadDirContents()
			return true
		case gxui.KeyEnter:
			if len(f.completions) == 0 {
				return false
			}
			fullPath := filepath.Join(f.dir.Text(), f.completions[0].Text())
			finfo, err := os.Stat(fullPath)
			if os.IsNotExist(err) {
				return false
			}
			if err != nil {
				return false
			}
			if !finfo.IsDir() {
				f.file.setFile(finfo.Name())
				return false
			}
			f.dir.SetText(fullPath)
			f.file.setFile("")
			go f.loadDirContents()
			return true
		case gxui.KeyRight:
			f.clearCompletions(f.completions)
			for i := 1; i < len(f.completions); i++ {
				f.completions[i-1], f.completions[i] = f.completions[i], f.completions[i-1]
			}
			f.addCompletions(f.completions)
			return true
		case gxui.KeyLeft:
			f.clearCompletions(f.completions)
			for i := len(f.completions) - 1; i > 0; i-- {
				f.completions[i-1], f.completions[i] = f.completions[i], f.completions[i-1]
			}
			f.addCompletions(f.completions)
			return true
		case gxui.KeyBackspace:
			if len(f.file.Text()) == 0 {
				newDir := filepath.Dir(f.dir.Text())
				if newDir == f.dir.Text() {
					newDir = systemRoot
				}
				f.dir.SetText(newDir)
				go f.loadDirContents()
				return true
			}
		}
	}
	return f.file.KeyPress(event)
}

func (f *FSLocator) KeyDown(event gxui.KeyboardEvent) {
	f.file.KeyDown(event)
}

func (f *FSLocator) KeyUp(event gxui.KeyboardEvent) {
	f.file.KeyUp(event)
}

func (f *FSLocator) KeyStroke(event gxui.KeyStrokeEvent) bool {
	defer f.updateCompletions()
	if event.Character == filepath.Separator {
		return false
	}
	return f.file.KeyStroke(event)
}

func (f *FSLocator) KeyRepeat(event gxui.KeyboardEvent) {
	f.file.KeyRepeat(event)
}

func (f *FSLocator) Paint(c gxui.Canvas) {
	f.LinearLayout.Paint(c)

	if f.HasFocus() {
		r := f.Size().Rect()
		s := f.theme.FocusedStyle
		c.DrawRoundedRect(r, 3, 3, 3, 3, s.Pen, s.Brush)
	}
}

func (f *FSLocator) IsFocusable() bool {
	return f.file.IsFocusable()
}

func (f *FSLocator) HasFocus() bool {
	return f.file.HasFocus()
}

func (f *FSLocator) GainedFocus() {
	f.file.GainedFocus()
}

func (f *FSLocator) LostFocus() {
	f.file.LostFocus()
}

func (f *FSLocator) OnGainedFocus(callback func()) gxui.EventSubscription {
	return f.file.OnGainedFocus(callback)
}

func (f *FSLocator) OnLostFocus(callback func()) gxui.EventSubscription {
	return f.file.OnLostFocus(callback)
}

func (f *FSLocator) updateCompletions() {
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

func (f *FSLocator) clearCompletions(completions []gxui.Label) {
	cloned := make([]gxui.Label, 0, len(completions))
	cloned = append(cloned, completions...)
	f.driver.Call(func() {
		for _, l := range cloned {
			f.RemoveChild(l)
		}
	})
}

func (f *FSLocator) addCompletions(completions []gxui.Label) {
	cloned := make([]gxui.Label, 0, len(completions))
	cloned = append(cloned, completions...)
	f.driver.Call(func() {
		for _, l := range cloned {
			f.AddChild(l)
			l.SetHorizontalAlignment(gxui.AlignCenter)
		}
	})
}

func (f *FSLocator) loadDirContents() {
	defer f.updateCompletions()

	dir := f.dir.Text()
	if dir == "" {
		log.Printf("This is odd: we have an empty directory")
		return
	}
	f.files = nil
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
	if len(text) == 0 || text == "/" {
		return "/"
	}
	if text[len(text)-1] == filepath.Separator {
		text = text[:len(text)-1]
	}
	return text
}

type fileBox struct {
	mixins.TextBox

	font gxui.Font
}

func newFileBox(driver gxui.Driver, theme *basic.Theme) *fileBox {
	file := &fileBox{
		font: theme.DefaultMonospaceFont(),
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

func findStart(control gxui.Control) string {
	if startingPath := findCurrentFile(control); startingPath != "" {
		return filepath.Dir(startingPath)
	}
	if project, ok := findProject(control); ok {
		return project.Path
	}
	return settings.DefaultProject.Path
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

func findProject(control gxui.Control) (settings.Project, bool) {
	switch src := control.(type) {
	case ProjectGetter:
		return src.CurrentProject(), true
	case gxui.Parent:
		for _, child := range src.Children() {
			if proj, ok := findProject(child.Control); ok {
				return proj, true
			}
		}
	}
	return settings.Project{}, false
}
