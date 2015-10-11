package commander

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
)

type FSLocator struct {
	mixins.LinearLayout

	theme *basic.Theme
	dir   *dirLabel
	file  *fileBox

	callback func(filepath string)
}

func NewFSLocator(driver gxui.Driver, theme *basic.Theme, startingPath string, callback func(filepath string)) *FSLocator {
	dirname, filename := filepath.Split(startingPath)
	locator := &FSLocator{
		theme:    theme,
		dir:      newDirLabel(theme, dirname),
		file:     newFileBox(driver, theme, filename),
		callback: callback,
	}
	locator.LinearLayout.Init(locator, theme)
	locator.SetDirection(gxui.LeftToRight)

	locator.dir.SetText(dirname)
	locator.AddChild(locator.dir)

	locator.AddChild(locator.file)
	return locator
}

func (f *FSLocator) KeyPress(event gxui.KeyboardEvent) bool {
	switch event.Key {
	case gxui.KeyTab, gxui.KeySlash, gxui.KeyEnter:
		fullPath := filepath.Join(f.dir.Text(), f.file.Text())
		matches, err := filepath.Glob(fullPath + "*")
		if err != nil {
			panic(err)
		}
		if len(matches) == 1 {
			fullPath = matches[0]
			f.file.setFile(filepath.Base(fullPath))
		}
		if finfo, err := os.Stat(fullPath); err == nil && finfo.IsDir() {
			f.dir.SetText(fullPath)
			f.file.SetText("")
			return true
		}
		if event.Key == gxui.KeyEnter {
			f.callback(fullPath)
			return true
		}
	case gxui.KeyBackspace:
		if len(f.file.Text()) == 0 {
			newDir := filepath.Dir(f.dir.Text())
			f.dir.SetText(newDir)
			return true
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
	if event.Character == '/' {
		return false
	}
	return f.file.KeyStroke(event)
}

func (f *FSLocator) KeyRepeat(event gxui.KeyboardEvent) {
	f.file.KeyRepeat(event)
}

func (f *FSLocator) Paint(c gxui.Canvas) {
	f.LinearLayout.Paint(c)

	if f.file.HasFocus() {
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

type dirLabel struct {
	mixins.Label
}

func newDirLabel(theme *basic.Theme, startingDir string) *dirLabel {
	label := new(dirLabel)
	label.Label.Init(label, theme, theme.DefaultMonospaceFont(), theme.LabelStyle.FontColor)
	label.SetMargin(math.Spacing{L: 3, T: 3, R: 3, B: 3})
	return label
}

func (l *dirLabel) SetText(dir string) {
	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}
	l.Label.SetText(dir)
}

func (l *dirLabel) Text() string {
	return strings.TrimSuffix(l.Label.Text(), "/")
}

type fileBox struct {
	mixins.TextBox
}

func newFileBox(driver gxui.Driver, theme *basic.Theme, startingFile string) *fileBox {
	file := new(fileBox)
	file.TextBox.Init(file, driver, theme, theme.DefaultMonospaceFont())
	file.SetTextColor(theme.TextBoxDefaultStyle.FontColor)
	file.SetMargin(math.Spacing{L: 3, T: 3, R: 3, B: 3})
	file.SetPadding(math.Spacing{L: 3, T: 3, R: 3, B: 3})
	file.SetBackgroundBrush(theme.TextBoxDefaultStyle.Brush)
	file.SetDesiredWidth(math.MaxSize.W)
	file.SetMultiline(false)
	file.SetText(startingFile)
	file.Controller().SetCaret(len(startingFile))
	return file
}

func (f *fileBox) setFile(file string) {
	f.SetText(file)
	f.Controller().SetCaret(len(file))
}
