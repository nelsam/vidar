// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package vfmt

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/text"
	"github.com/nelsam/vidar/plugin/status"
	"github.com/nelsam/vidar/setting"
)

const stdinPathPattern = "<standard input>:"

type Applier interface {
	Apply(text.Editor, ...text.Edit)
}

type OnSave struct {
}

func (o OnSave) Name() string {
	return "vfmt-on-save"
}

func (o OnSave) OpName() string {
	return "save-current-file"
}

func (o OnSave) BeforeSave(_ setting.Project, _, text string) (newText string, err error) {
	return vfmt(text)
}

// LabelCreator is used to create labels to display to the user.
type LabelCreator interface {
	CreateLabel() gxui.Label
}

type Vfmt struct {
	status.General

	editor  text.Editor
	applier Applier
}

func New(c LabelCreator) *Vfmt {
	g := &Vfmt{}
	g.Theme = c
	return g
}

func (f *Vfmt) Name() string {
	return "vfmt"
}

func (f *Vfmt) Menu() string {
	return "Vlang"
}

func (f *Vfmt) Defaults() []fmt.Stringer {
	return []fmt.Stringer{gxui.KeyboardEvent{
		Modifier: gxui.ModControl | gxui.ModShift,
		Key:      gxui.KeyF,
	}}
}

func (f *Vfmt) Reset() {
	f.editor = nil
	f.applier = nil
}

func (f *Vfmt) Store(target interface{}) bind.Status {
	switch src := target.(type) {
	case text.Editor:
		f.editor = src
	case Applier:
		f.applier = src
	}
	if f.editor != nil && f.applier != nil {
		return bind.Done
	}
	return bind.Waiting
}

func (f *Vfmt) Exec() error {
	current := f.editor.Text()
	formatted, err := vfmt(current)
	if err != nil {
		f.Err = err.Error()
		return err
	}
	f.applier.Apply(f.editor, text.Edit{
		At:  0,
		Old: []rune(current),
		New: []rune(formatted),
	})
	return nil
}

func vfmt(text string) (newText string, err error) {
	// TODO: make `v fmt` accept stdin
	tmppath := filepath.Join(os.TempDir(), "vfmt-tmp.v")
	tmpf, err := os.Create(tmppath)
	if err != nil {
		return "", err
	}
	defer os.Remove(tmppath)
	tmpf.Write([]byte(text))
	tmpf.Close()

	cmd := exec.Command("v", "fmt", tmppath)
	errBuffer := &bytes.Buffer{}
	cmd.Stderr = errBuffer
	formatted, err := cmd.Output()
	if err != nil {
		msg := errBuffer.String()
		if msg == "" {
			return "", err
		}
		stdinPatternStart := strings.Index(msg, stdinPathPattern)
		if stdinPatternStart > 0 {
			pathEnd := stdinPatternStart + len(stdinPathPattern)
			msg = msg[pathEnd:]
		}
		return "", fmt.Errorf("vfmt: %s", msg)
	}
	return string(formatted), nil
}
