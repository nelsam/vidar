// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"bytes"
	"fmt"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander"
)

type Copy struct {
	driver gxui.Driver
}

func NewCopy(driver gxui.Driver) *Copy {
	return &Copy{
		driver: driver,
	}
}

func (c *Copy) Name() string {
	return "copy-selection"
}

func (c *Copy) Menu() string {
	return "Edit"
}

func (c *Copy) Exec(target interface{}) (executed, consume bool) {
	finder, ok := target.(EditorFinder)
	if !ok {
		return false, false
	}
	editor := finder.CurrentEditor()
	if editor == nil {
		return true, true
	}

	selections := editor.Controller().Selections()
	var buffer bytes.Buffer
	for i := 0; i < selections.Len(); i++ {
		buffer.WriteString(editor.Controller().SelectionText(i))
	}

	c.driver.SetClipboard(buffer.String())
	return true, true
}

type Cut struct {
	Copy
}

func NewCut(driver gxui.Driver) *Cut {
	copy := NewCopy(driver)
	return &Cut{Copy: *copy}
}

func (c *Cut) Name() string {
	return "cut-selection"
}

func (c *Cut) Exec(target interface{}) (executed, consume bool) {
	executed, consume = c.Copy.Exec(target)
	if !executed {
		return executed, consume
	}
	editor := target.(EditorFinder).CurrentEditor()
	if editor == nil {
		return true, consume
	}
	selection := editor.Controller().FirstSelection()
	newRunes, edit := editor.Controller().ReplaceAt(editor.Runes(), selection.Start(), selection.End(), []rune(""))
	editor.Controller().SetTextEdits(newRunes, []gxui.TextBoxEdit{edit})
	editor.Controller().ClearSelections()
	return true, consume
}

type Paste struct {
	commander.GenericStatuser

	driver gxui.Driver
}

func NewPaste(driver gxui.Driver, theme gxui.Theme) *Paste {
	p := &Paste{}
	p.Theme = theme
	p.driver = driver
	return p
}

func (p *Paste) Name() string {
	return "paste"
}

func (p *Paste) Menu() string {
	return "Edit"
}

func (p *Paste) Exec(target interface{}) (executed, consume bool) {
	finder, ok := target.(EditorFinder)
	if !ok {
		return false, false
	}
	editor := finder.CurrentEditor()
	if editor == nil {
		return true, true
	}
	contents, err := p.driver.GetClipboard()
	if err != nil {
		p.Err = fmt.Sprintf("Error reading clipboard: %s", err)
		return true, false
	}
	editor.Controller().Replace(func(sel gxui.TextSelection) string {
		return contents
	})
	return true, true
}
