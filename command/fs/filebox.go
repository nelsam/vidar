// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package fs

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/setting"
)

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
	if event.Modifier != 0 {
		return l.file.KeyPress(event)
	}
	l.lock.RLock()
	defer l.lock.RUnlock()

	switch event.Key {
	case gxui.KeyEscape:
		if len(l.completions) > 0 {
			l.clearCompletions(l.completions)
			l.completions = nil
			return true
		}
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
		if len(l.file.Text()) != 0 {
			break
		}
		oldDir := strings.TrimSuffix(l.dir.Text(), string(filepath.Separator))
		newDir := filepath.Dir(oldDir)
		if oldDir == filepath.VolumeName(oldDir) {
			newDir = systemRoot
		}
		l.dir.SetText(newDir)
		go l.loadDirContents()
		return true
	}
	return f.TextBox.KeyPress(event)
}

func (f *fileBox) KeyStroke(event gxui.KeyStrokeEvent) bool {
	defer f.locator.updateCompletions()
	fullPath := f.locator.Path()
	if !pathSeparator(fullPath, event.Character) {
		return f.TextBox.KeyStroke(event)
	}
	if len(f.locator.completions) > 0 {
		fullPath = filepath.Join(f.locator.dir.Text(), f.locator.completions[0].Text())
	}
	f.locator.dir.SetText(fullPath)
	f.setFile("")
	go f.locator.loadDirContents()
	return false
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
