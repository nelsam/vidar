// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package goimports

import (
	"bytes"
	"fmt"
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
	return "goimports-on-save"
}

func (o OnSave) OpName() string {
	return "save-current-file"
}

func (o OnSave) BeforeSave(proj setting.Project, path, text string) (newText string, err error) {
	return goimports(path, text, proj.Environ())
}

type GoImports struct {
	status.General

	editor    input.Editor
	projecter Projecter
	applier   Applier
}

func New(theme gxui.Theme) *GoImports {
	g := &GoImports{}
	g.Theme = theme
	return g
}

func (gi *GoImports) Name() string {
	return "goimports"
}

func (gi *GoImports) Menu() string {
	return "Golang"
}

func (gi *GoImports) Defaults() []fmt.Stringer {
	return []fmt.Stringer{gxui.KeyboardEvent{
		Modifier: gxui.ModControl | gxui.ModShift,
		Key:      gxui.KeyF,
	}}
}

func (gi *GoImports) Reset() {
	gi.editor = nil
	gi.projecter = nil
	gi.applier = nil
}

func (gi *GoImports) Store(target interface{}) bind.Status {
	switch src := target.(type) {
	case input.Editor:
		gi.editor = src
	case Projecter:
		gi.projecter = src
	case Applier:
		gi.applier = src
	}
	if gi.editor != nil && gi.projecter != nil && gi.applier != nil {
		return bind.Done
	}
	return bind.Waiting
}

func (gi *GoImports) Exec() error {
	proj := gi.projecter.Project()
	text := gi.editor.Text()
	formatted, err := goimports(gi.editor.Filepath(), text, proj.Environ())
	if err != nil {
		gi.Err = err.Error()
		return err
	}
	gi.applier.Apply(gi.editor, input.Edit{
		At:  0,
		Old: []rune(text),
		New: []rune(formatted),
	})
	return nil
}

func goimports(path, text string, env []string) (newText string, err error) {
	cmd := exec.Command("goimports")
	cmd.Stdin = bytes.NewBufferString(text)
	errBuffer := &bytes.Buffer{}
	cmd.Stderr = errBuffer
	cmd.Env = env
	cmd.Dir = filepath.Dir(path)
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
		return "", fmt.Errorf("goimports: %s", msg)
	}
	return string(formatted), nil
}
