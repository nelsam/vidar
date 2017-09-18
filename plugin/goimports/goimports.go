// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package goimports

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
	"github.com/nelsam/vidar/settings"
)

const stdinPathPattern = "<standard input>:"

type Projecter interface {
	Project() settings.Project
}

type Editor interface {
	input.Editor
	Filepath() string
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

func (o OnSave) BeforeSave(proj settings.Project, path, text string) (newText string, err error) {
	return goimports(proj.Gopath, path, text)
}

type GoImports struct {
	status.General

	editor    Editor
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

func (gi *GoImports) Reset() {
	gi.editor = nil
	gi.projecter = nil
	gi.applier = nil
}

func (gi *GoImports) Store(target interface{}) bind.Status {
	switch src := target.(type) {
	case Editor:
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
	formatted, err := goimports(proj.Gopath, gi.editor.Filepath(), text)
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

func goimports(gopath, path, text string) (newText string, err error) {
	cmd := exec.Command("goimports", "-srcdir", filepath.Dir(path))
	cmd.Stdin = bytes.NewBufferString(text)
	errBuffer := &bytes.Buffer{}
	cmd.Stderr = errBuffer
	cmd.Env = []string{"PATH=" + os.Getenv("PATH")}
	if gopath != "" {
		cmd.Env[0] += string(os.PathListSeparator) + filepath.Join(gopath, "bin")
		cmd.Env = append(cmd.Env, "GOPATH="+gopath)
	}
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
