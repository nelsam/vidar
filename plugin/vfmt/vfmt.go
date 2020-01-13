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
	"github.com/nelsam/vidar/commander/input"
	"github.com/nelsam/vidar/plugin/status"
	"github.com/nelsam/vidar/setting"
)

const stdinPathPattern = "<standard input>:"

type Projecter interface {
	Project() setting.Project
}

type Applier interface {
	Apply(input.Editor, ...input.Edit)
}

type OnSave struct {
}

func (o OnSave) Name() string {
	return "vfmt-on-save"
}

func (o OnSave) OpName() string {
	return "save-current-file"
}

func (o OnSave) BeforeSave(proj setting.Project, path, text string) (newText string, err error) {
	return vfmt(text)
}

type Vfmt struct {
	status.General

	editor  input.Editor
	applier Applier
}

func New(theme gxui.Theme) *Vfmt {
	g := &Vfmt{}
	g.Theme = theme
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
	case input.Editor:
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
	text := f.editor.Text()
	formatted, err := vfmt(text)
	if err != nil {
		f.Err = err.Error()
		return err
	}
	f.applier.Apply(f.editor, input.Edit{
		At:  0,
		Old: []rune(text),
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
