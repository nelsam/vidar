// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nelsam/gxui"
)

const stdinPathPattern = "<standard input>:"

type GoImports struct {
	statusKeeper
}

func NewGoImports(theme gxui.Theme) *GoImports {
	return &GoImports{statusKeeper: statusKeeper{theme: theme}}
}

func (gi *GoImports) Name() string {
	return "goimports"
}

func (gi *GoImports) Menu() string {
	return "Edit"
}

func (gi *GoImports) Exec(on interface{}) (executed, consume bool) {
	finder, ok := on.(ProjectFinder)
	if !ok {
		return false, false
	}
	editor := finder.CurrentEditor()
	if editor == nil {
		return true, true
	}
	proj := finder.Project()
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
			gi.err = err.Error()
			return true, true
		}
		stdinPatternStart := strings.Index(msg, stdinPathPattern)
		if stdinPatternStart > 0 {
			pathEnd := stdinPatternStart + len(stdinPathPattern)
			msg = msg[pathEnd:]
		}
		gi.err = fmt.Sprintf("goimports error: %s", msg)
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
