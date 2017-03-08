// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// Package godef contains logc for interacting with the godef
// command line tool.  It can be imported directly or used as a
// plugn.
package godef

import (
	"bytes"
	"fmt"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/editor"
	"github.com/nelsam/vidar/plugin/status"
	"github.com/nelsam/vidar/settings"
)

type OpenProject interface {
	Project() settings.Project
	CurrentEditor() *editor.CodeEditor
}

type Commander interface {
	Command(name string) bind.Command
}

type Opener interface {
	SetLocation(filepath string, position token.Position)
	Exec(interface{}) (executed, consume bool)
}

type Godef struct {
	status.General

	proj   OpenProject
	cmdr   Commander
	opener Opener

	callables []interface{}
}

func New(theme gxui.Theme) *Godef {
	g := &Godef{}
	g.Theme = theme
	return g
}

func (g *Godef) Name() string {
	return "goto-definition"
}

func (g *Godef) Menu() string {
	return "Edit"
}

func (g *Godef) Start(gxui.Control) gxui.Control {
	g.proj = nil
	g.cmdr = nil
	g.callables = nil
	g.opener = nil
	return nil
}

func (g *Godef) Exec(on interface{}) (executed, consume bool) {
	// TODO: Refactor this!  Seriously, vidar's design is forcing
	// this to be more complicated than it needs to be.
	backtrack := false
	switch src := on.(type) {
	case OpenProject:
		g.proj = src
		if g.cmdr != nil {
			backtrack = true
			g.opener = g.setup()
		}
	case Commander:
		g.cmdr = src
		if g.proj != nil {
			backtrack = true
			g.opener = g.setup()
		}
	}
	if g.opener == nil {
		g.callables = append(g.callables, on)
		return false, false
	}

	if backtrack {
		for _, c := range g.callables {
			subExecuted, subConsume := g.opener.Exec(c)
			executed = executed || subExecuted
			if subConsume {
				return executed, subConsume
			}
		}
	}

	return g.opener.Exec(on)
}

func (g *Godef) setup() Opener {
	editor := g.proj.CurrentEditor()
	if editor == nil {
		return nil
	}
	proj := g.proj.Project()
	lastCaret := editor.Controller().LastCaret()
	cmd := exec.Command("godef", "-f", editor.Filepath(), "-o", strconv.Itoa(lastCaret), "-i")
	cmd.Stdin = bytes.NewBufferString(editor.Text())
	errBuffer := &bytes.Buffer{}
	cmd.Stderr = errBuffer
	cmd.Env = []string{"PATH=" + os.Getenv("PATH")}
	if proj.Gopath != "" {
		cmd.Env[0] += string(os.PathListSeparator) + filepath.Join(proj.Gopath, "bin")
		cmd.Env = append(cmd.Env, "GOPATH="+proj.Gopath)
	}
	output, err := cmd.Output()
	if err != nil {
		g.Err = fmt.Sprintf("godef error: %s", string(output))
		return nil
	}
	path, line, column, err := parseGodef(output)
	if err != nil {
		g.Err = err.Error()
		return nil
	}
	open, ok := g.cmdr.Command("open-file").(Opener)
	if !ok {
		g.Err = "no open-file command found"
		return nil
	}
	open.SetLocation(path, token.Position{
		Line:   line,
		Column: column,
	})
	return open
}

func parseGodef(output []byte) (path string, line, column int, err error) {
	values := bytes.Split(bytes.TrimSpace(output), []byte{':'})
	if len(values) != 3 {
		return "", 0, 0, fmt.Errorf("godef output %s not understood", string(output))
	}
	pathBytes, lineBytes, colBytes := values[0], values[1], values[2]
	line, err = strconv.Atoi(string(lineBytes))
	if err != nil {
		return "", 0, 0, fmt.Errorf("godef output: %s is not a line number", string(lineBytes))
	}
	col, err := strconv.Atoi(string(colBytes))
	if err != nil {
		return "", 0, 0, fmt.Errorf("godef output: %s is not a column number", string(colBytes))
	}
	return string(pathBytes), line - 1, col - 1, nil
}
