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
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/bind"
	"github.com/nelsam/vidar/command/focus"
	"github.com/nelsam/vidar/plugin/status"
	"github.com/nelsam/vidar/setting"
)

type Projecter interface {
	Project() setting.Project
}

type Commander interface {
	Execute(bind.Bindable)
}

type Opener interface {
	For(...focus.Opt) bind.Bindable
}

type Editor interface {
	Filepath() string
	Text() string
	LineStart(int) int
}

type CursorController interface {
	LastCaret() int
}

type Godef struct {
	status.General

	proj   Projecter
	cmdr   Commander
	opener Opener
	editor Editor
	ctrl   CursorController
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
	return "Golang"
}

func (g *Godef) Defaults() []fmt.Stringer {
	return []fmt.Stringer{gxui.KeyboardEvent{
		Modifier: gxui.ModControl | gxui.ModShift,
		Key:      gxui.KeyG,
	}}
}

func (g *Godef) Reset() {
	g.proj = nil
	g.cmdr = nil
	g.opener = nil
	g.ctrl = nil
	g.editor = nil
}

func (g *Godef) Store(target interface{}) bind.Status {
	switch src := target.(type) {
	case Projecter:
		g.proj = src
	case Commander:
		g.cmdr = src
	case Editor:
		g.editor = src
	case CursorController:
		g.ctrl = src
	case Opener:
		g.opener = src
	}
	if g.proj != nil && g.cmdr != nil && g.opener != nil && g.ctrl != nil && g.editor != nil {
		return bind.Done
	}
	return bind.Waiting
}

func (g *Godef) Exec() error {
	proj := g.proj.Project()
	lastCaret := g.ctrl.LastCaret()
	cmd := exec.Command("godef", "-f", g.editor.Filepath(), "-o", strconv.Itoa(lastCaret), "-i")
	cmd.Stdin = bytes.NewBufferString(g.editor.Text())
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
		return err
	}
	path, line, col, err := parseGodef(output)
	if err != nil {
		g.Err = err.Error()
		return err
	}
	g.cmdr.Execute(g.opener.For(focus.Path(path), focus.Line(line), focus.Column(col)))
	return nil
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
