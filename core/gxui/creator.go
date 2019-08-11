// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// Package gxui implements a wrapper between github.com/nelsam/gxui
// and github.com/nelsam/vidar/ui.  Some day, we'll allow plugins
// to override this - but for now, this just decouples the majority
// of our code from direct implementations of gxui.
package gxui

import (
	"errors"
	"fmt"
	"sync"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/drivers/gl"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/gxui/themes/dark"
	"github.com/nelsam/vidar/theme"
	"github.com/nelsam/vidar/ui"
)

var background = gxui.Gray10

// Creator wraps up all the gxui element creation logic in a
// ui.Creator way.
type Creator struct {
	driver gxui.Driver
	theme  *basic.Theme

	windows *windowList
	wg      *sync.WaitGroup
}

func (c Creator) Name() string {
	return "gxui"
}

func (c Creator) Theme() theme.Theme {
	return theme.Default
}

// Raw is a temporary function to return the raw gxui elements
// to the caller.  It will be removed once vidar's core logic is
// no longer tightly coupled to gxui mixins.
func (c Creator) Raw() (gxui.Driver, gxui.Theme) {
	return c.driver, c.theme
}

func (c Creator) Start() error {
	if c.wg != nil {
		return errors.New("gxui: Creator.Init() called for the second time!")
	}
	c.wg = &sync.WaitGroup{}

	go gl.StartDriver(c.uiMain)
	return nil
}

func (c Creator) uiMain(d gxui.Driver) {
	c.driver = d
	c.theme = dark.CreateTheme(d).(*basic.Theme)

	font := font(d)
	if font == nil {
		font = c.theme.DefaultMonospaceFont()
	}
	c.theme.SetDefaultMonospaceFont(font)
	c.theme.SetDefaultFont(font)
	c.theme.WindowBackground = background
}

func (c Creator) Wait() {
	c.wg.Wait()
}

func (c Creator) Quit() {
	c.driver.Terminate()
}

func (c Creator) Runner() ui.Runner {
	return Runner{driver: c.driver}
}

func (c Creator) Window(size ui.Size, name string) (ui.Window, error) {
	w := newWindow(c.theme, size, name)
	w.window.OnClose(func() {
		c.wg.Done()
	})
	c.wg.Add(1)
	return w, nil
}

func (c Creator) LinearLayout(start ui.Direction) (ui.Layout, error) {
	l := c.theme.CreateLinearLayout()
	var d gxui.Direction
	switch start {
	case ui.Left:
		d = gxui.LeftToRight
	case ui.Up:
		d = gxui.TopToBottom
	case ui.Right:
		d = gxui.RightToLeft
	case ui.Down:
		d = gxui.BottomToTop
	default:
		return nil, fmt.Errorf("Direction %v is not a valid linear layout start", start)
	}
	l.SetDirection(d)
	return LinearLayout{layout: l}, nil
}

func (c Creator) ManualLayout() (ui.Layout, error) {
	return newManualLayout(c)
}

func (c Creator) TabbedLayout() (ui.Layout, error) {
	h := c.theme.CreatePanelHolder()
	return TabbedLayout{h}, nil
}

func (c Creator) SplitLayout() (ui.Layout, error) {
	return nil, fmt.Errorf("gxui.SplitLayout: Not implemented")
}

func (c Creator) Label(v []rune) (ui.TextDisplay, error) {
	l := c.theme.CreateLabel()
	l.SetText(string(v))
	return Label{label: l}, nil
}

func (c Creator) Editor() (ui.TextEditor, error) {
	ed := &codeEditor{}
	ed.CodeEditor.Init(ed, c.driver, c.theme, c.theme.DefaultMonospaceFont())
	ed.CodeEditor.SetScrollBarEnabled(true)
	ed.CodeEditor.SetScrollRound(true)
	ed.CodeEditor.SetDesiredWidth(math.MaxSize.W)
	ed.CodeEditor.SetTextColor(c.theme.TextBoxDefaultStyle.FontColor)
	ed.CodeEditor.SetMargin(math.Spacing{L: 3, T: 3, R: 3, B: 3})
	ed.CodeEditor.SetPadding(math.Spacing{L: 3, T: 3, R: 3, B: 3})
	ed.CodeEditor.SetBorderPen(gxui.TransparentPen)
	return Editor{ed: ed}, nil
}
