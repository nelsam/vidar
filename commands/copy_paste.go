package commands

import "github.com/nelsam/gxui"

type Copy struct {
	driver gxui.Driver
}

func NewCopy(driver gxui.Driver) *Copy {
	return &Copy{
		driver: driver,
	}
}

func (c *Copy) Start(gxui.Control) gxui.Control {
	return nil
}

func (c *Copy) Name() string {
	return "copy-selection"
}

func (c *Copy) Next() gxui.Focusable {
	return nil
}

func (c *Copy) Exec(target interface{}) (executed, consume bool) {
	finder, ok := target.(EditorFinder)
	if !ok {
		return false, false
	}
	editor := finder.CurrentEditor()
	if editor.Controller().SelectionCount() != 1 {
		panic("Trying to copy without exactly 1 selection")
	}
	selection := editor.Controller().SelectionText(0)
	c.driver.SetClipboard(selection)
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
	selection := editor.Controller().FirstSelection()
	newRunes, _ := editor.Controller().ReplaceAt(editor.Runes(), selection.Start(), selection.End(), []rune(""))
	editor.SetText(string(newRunes))
	return
}

type Paste struct {
	driver gxui.Driver
}

func NewPaste(driver gxui.Driver) *Paste {
	return &Paste{
		driver: driver,
	}
}

func (p *Paste) Name() string {
	return "paste"
}

func (p *Paste) Start(gxui.Control) gxui.Control {
	return nil
}

func (p *Paste) Next() gxui.Focusable {
	return nil
}

func (p *Paste) Exec(target interface{}) (executed, consume bool) {
	finder, ok := target.(EditorFinder)
	if !ok {
		return false, false
	}
	contents, err := p.driver.GetClipboard()
	if err != nil {
		panic(err)
	}
	editor := finder.CurrentEditor()
	editor.Controller().Replace(func(sel gxui.TextSelection) string {
		return contents
	})
	return true, true
}
