// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"bytes"
	"fmt"
	"log"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/input"
	"github.com/nelsam/vidar/editor"
	"github.com/nelsam/vidar/plugin/status"
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

func (c *Copy) Exec(target interface{}) bind.Status {
	finder, ok := target.(EditorFinder)
	if !ok {
		return bind.Waiting
	}
	editor := finder.CurrentEditor()
	if editor == nil {
		return bind.Done
	}

	selections := editor.Controller().Selections()
	var buffer bytes.Buffer
	for i := 0; i < selections.Len(); i++ {
		buffer.WriteString(editor.Controller().SelectionText(i))
	}

	c.driver.SetClipboard(buffer.String())
	return bind.Done
}

type InputController interface {
	InputHandler() input.Handler
}

type Cut struct {
	Copy

	editor       *editor.CodeEditor
	inputHandler input.Handler
}

func NewCut(driver gxui.Driver) *Cut {
	copy := NewCopy(driver)
	return &Cut{Copy: *copy}
}

func (c *Cut) Name() string {
	return "cut-selection"
}

func (c *Cut) Start(gxui.Control) gxui.Control {
	c.inputHandler = nil
	c.editor = nil
	return nil
}

func (c *Cut) Exec(target interface{}) bind.Status {
	switch src := target.(type) {
	case EditorFinder:
		status := c.Copy.Exec(src)
		if status&bind.Executed == 0 {
			log.Printf("Error: copy command failed to exec on EditorFinder")
		}
		c.editor = src.CurrentEditor()
		if c.inputHandler != nil {
			c.removeSelections()
			return bind.Done
		}
	case InputController:
		c.inputHandler = src.InputHandler()
		if c.editor != nil {
			c.removeSelections()
			return bind.Done
		}
	}
	return bind.Waiting
}

func (c *Cut) removeSelections() {
	text := c.editor.Controller().TextRunes()
	var edits []input.Edit
	for _, s := range c.editor.Controller().Selections() {
		old := text[s.Start():s.End()]
		edits = append(edits, input.Edit{
			At:  s.Start(),
			Old: old,
		})
	}
	c.inputHandler.Apply(c.editor, edits...)
}

type Paste struct {
	status.General

	driver       gxui.Driver
	editor       *editor.CodeEditor
	inputHandler input.Handler
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

func (p *Paste) Start(gxui.Control) gxui.Control {
	p.editor = nil
	p.inputHandler = nil
	return nil
}

func (p *Paste) Exec(target interface{}) bind.Status {
	switch src := target.(type) {
	case EditorFinder:
		p.editor = src.CurrentEditor()
		if p.inputHandler != nil {
			p.replaceSelections()
			return bind.Done
		}
	case InputController:
		p.inputHandler = src.InputHandler()
		if p.editor != nil {
			p.replaceSelections()
			return bind.Done
		}
	}
	return bind.Waiting
}

func (p *Paste) replaceSelections() {
	text := p.editor.Controller().TextRunes()
	var edits []input.Edit
	contents, err := p.driver.GetClipboard()
	if err != nil {
		p.Err = fmt.Sprintf("Error reading clipboard: %s", err)
		return
	}
	replacement := []rune(contents)
	for _, s := range p.editor.Controller().Selections() {
		old := text[s.Start():s.End()]
		edits = append(edits, input.Edit{
			At:  s.Start(),
			Old: old,
			New: replacement,
		})
	}
	p.inputHandler.Apply(p.editor, edits...)
}
