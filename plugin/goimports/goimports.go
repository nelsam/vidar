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
	"github.com/nelsam/vidar/editor"
	"github.com/nelsam/vidar/plugin/status"
	"github.com/nelsam/vidar/settings"
)

const stdinPathPattern = "<standard input>:"

type OpenProject interface {
	CurrentEditor() *editor.CodeEditor
	Project() settings.Project
}

type OnSave struct {
}

func (o OnSave) Name() string {
	return "goimports-on-save"
}

func (o OnSave) CommandName() string {
	return "save-current-file"
}

func (o OnSave) BeforeSave(proj settings.Project, path, text string) (newText string, err error) {
	return goimports(proj.Gopath, path, text)
}

type GoImports struct {
	status.General
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
	return "Edit"
}

func (gi *GoImports) Exec(on interface{}) (executed, consume bool) {
	current, ok := on.(OpenProject)
	if !ok {
		return false, false
	}
	editor := current.CurrentEditor()
	if editor == nil {
		return true, true
	}
	proj := current.Project()
	text := editor.Text()
	formatted, err := goimports(proj.Gopath, editor.Filepath(), text)
	if err != nil {
		gi.Err = err.Error()
		return true, true
	}
	edits := []gxui.TextBoxEdit{
		{
			At:    0,
			Delta: len(formatted) - len(text),
			Old:   []rune(text),
			New:   []rune(formatted),
		},
	}
	editor.Controller().SetTextEdits([]rune(formatted), edits)
	return true, true
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
