// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package command

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/input"
	"github.com/nelsam/vidar/plugin/status"
)

type Editor interface {
	input.Editor
	Controller() *gxui.TextBoxController
}

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

func (c *Copy) Defaults() []fmt.Stringer {
	return []fmt.Stringer{gxui.KeyboardEvent{
		Modifier: gxui.ModControl,
		Key:      gxui.KeyC,
	}}
}

func (c *Copy) Exec(target interface{}) bind.Status {
	editor, ok := target.(Editor)
	if !ok {
		return bind.Waiting
	}

	selections := editor.Controller().SelectionSlice()
	var buffer bytes.Buffer
	for i := 0; i < len(selections); i++ {
		buffer.WriteString(editor.Controller().SelectionText(i))
	}

	c.driver.SetClipboard(buffer.String())
	return bind.Done
}

type Cut struct {
	Copy

	editor  Editor
	applier Applier
}

func NewCut(driver gxui.Driver) *Cut {
	copy := NewCopy(driver)
	return &Cut{Copy: *copy}
}

func (c *Cut) Name() string {
	return "cut-selection"
}

func (c *Cut) Defaults() []fmt.Stringer {
	return []fmt.Stringer{gxui.KeyboardEvent{
		Modifier: gxui.ModControl,
		Key:      gxui.KeyX,
	}}
}

func (c *Cut) Reset() {
	c.editor = nil
	c.applier = nil
}

func (c *Cut) Store(target interface{}) bind.Status {
	switch src := target.(type) {
	case Editor:
		c.editor = src
	case Applier:
		c.applier = src
	}
	if c.editor != nil && c.applier != nil {
		return bind.Done
	}
	return bind.Waiting
}

func (c *Cut) Exec() error {
	status := c.Copy.Exec(c.editor)
	if status&bind.Executed == 0 {
		return errors.New("copy command failed to exec on Editor")
	}
	c.removeSelections()
	return nil
}

func (c *Cut) removeSelections() {
	text := c.editor.Controller().TextRunes()
	var edits []input.Edit
	for _, s := range c.editor.Controller().SelectionSlice() {
		old := text[s.Start():s.End()]
		edits = append(edits, input.Edit{
			At:  s.Start(),
			Old: old,
		})
	}
	c.applier.Apply(c.editor, edits...)
}

type Paste struct {
	status.General

	driver  gxui.Driver
	editor  Editor
	applier Applier
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

func (p *Paste) Defaults() []fmt.Stringer {
	return []fmt.Stringer{gxui.KeyboardEvent{
		Modifier: gxui.ModControl,
		Key:      gxui.KeyV,
	}}
}

func (p *Paste) Reset() {
	p.editor = nil
	p.applier = nil
}

func (p *Paste) Store(target interface{}) bind.Status {
	switch src := target.(type) {
	case Editor:
		p.editor = src
	case Applier:
		p.applier = src
	}
	if p.editor != nil && p.applier != nil {
		return bind.Done
	}
	return bind.Waiting
}

func (p *Paste) Exec() error {
	p.replaceSelections()
	return nil
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
	for _, s := range p.editor.Controller().SelectionSlice() {
		old := text[s.Start():s.End()]
		edits = append(edits, input.Edit{
			At:  s.Start(),
			Old: old,
			New: replacement,
		})
	}
	p.applier.Apply(p.editor, edits...)
}
