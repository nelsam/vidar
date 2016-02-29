// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package commands

import (
	"log"
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
}

func NewFSLocator(driver gxui.Driver, theme *basic.Theme) *FSLocator {
	f := new(FSLocator)
	f.Init(driver, theme)
	return f
}

func (f *FSLocator) Init(driver gxui.Driver, theme *basic.Theme) {
	f.LinearLayout.Init(f, theme)
	f.theme = theme

	f.SetDirection(gxui.LeftToRight)
	f.dir = newDirLabel(theme)
	f.AddChild(f.dir)
	f.file = newFileBox(driver, theme)
	f.AddChild(f.file)
}

func (f *FSLocator) loadEditorDir(control gxui.Control) {
	startingPath := findStart(control)

	f.dir.SetText(startingPath)
	f.file.SetText("")
}

func (f *FSLocator) Path() string {
	return filepath.Join(f.dir.Text(), f.file.Text())
}

func (f *FSLocator) SetPath(filePath string) {
	dir, file := filepath.Split(filePath)
	f.dir.SetText(dir)
	f.file.SetText(file)
}

func (f *FSLocator) KeyPress(event gxui.KeyboardEvent) bool {
	if event.Modifier == 0 {
		switch event.Key {
		case gxui.KeyTab, gxui.KeySlash:
			fullPath := f.Path()
			matches, err := filepath.Glob(fullPath + "*")
			if err != nil {
				log.Printf("Error globbing path: %s", err)
				return true
			}
			if len(matches) == 1 {
				fullPath = matches[0]
				f.file.setFile(filepath.Base(fullPath))
			}
			if finfo, err := os.Stat(fullPath); err == nil && finfo.IsDir() {
				f.dir.SetText(fullPath)
				f.file.SetText("")
			}
			return true
		case gxui.KeyBackspace:
			if len(f.file.Text()) == 0 {
				newDir := filepath.Dir(f.dir.Text())
				if newDir == f.dir.Text() {
					newDir = "/"
				}
				f.dir.SetText(newDir)
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

type dirLabel struct {
	mixins.Label
}

func newDirLabel(theme *basic.Theme) *dirLabel {
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

func newFileBox(driver gxui.Driver, theme *basic.Theme) *fileBox {
	file := new(fileBox)
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

type currentFiler interface {
	CurrentFile() string
}

func findStart(control gxui.Control) string {
	startingPath := findCurrentFile(control)
	if startingPath == "" {
		return os.Getenv("HOME")
	}
	return filepath.Dir(startingPath)
}

func findCurrentFile(control gxui.Control) string {
	switch src := control.(type) {
	case currentFiler:
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
