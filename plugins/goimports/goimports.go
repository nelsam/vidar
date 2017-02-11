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
	"github.com/nelsam/vidar/commander"
	"github.com/nelsam/vidar/editor"
	"github.com/nelsam/vidar/settings"
)

const stdinPathPattern = "<standard input>:"

type OpenProject interface {
	CurrentEditor() *editor.CodeEditor
	Project() settings.Project
}

type GoImports struct {
	commander.GenericStatuser
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
	cmd := exec.Command("goimports", "-srcdir", filepath.Dir(editor.Filepath()))
	cmd.Stdin = bytes.NewBufferString(text)
	errBuffer := &bytes.Buffer{}
	cmd.Stderr = errBuffer
	cmd.Env = []string{"PATH=" + os.Getenv("PATH")}
	if proj.Gopath != "" {
		cmd.Env[0] += string(os.PathListSeparator) + filepath.Join(proj.Gopath, "bin")
		cmd.Env = append(cmd.Env, "GOPATH="+proj.Gopath)
	}
	formatted, err := cmd.Output()
	if err != nil {
		msg := errBuffer.String()
		if msg == "" {
			gi.Err = err.Error()
			return true, true
		}
		stdinPatternStart := strings.Index(msg, stdinPathPattern)
		if stdinPatternStart > 0 {
			pathEnd := stdinPatternStart + len(stdinPathPattern)
			msg = msg[pathEnd:]
		}
		gi.Err = fmt.Sprintf("goimports error: %s", msg)
		return true, true
	}
	edits := []gxui.TextBoxEdit{
		{
			At:    0,
			Delta: len(formatted) - len(text),
			Old:   []rune(text),
			New:   []rune(string(formatted)),
		},
	}
	editor.Controller().SetTextEdits([]rune(string(formatted)), edits)
	return true, true
}
