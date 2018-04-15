// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// Package goguru contains logic for interacting with the guru
// command line tool.  It can be imported directly or used as a
// plugin.
package goguru

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/command/focus"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/plugin/status"
	"github.com/nelsam/vidar/setting"
)

// zeroIndexOffset represents how far off of a zero-index guru's
// indexes are.
const zeroIndexOffset = 1

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

type ObjectPos struct {
	Path string
	Line int
	Col  int
}

func jsonStrVal(v []byte) []byte {
	if len(v) == 0 || v[0] != '"' {
		return nil
	}
	return v[1 : len(v)-1]
}

func (p *ObjectPos) UnmarshalJSON(v []byte) error {
	v = jsonStrVal(v)
	pathEnd := bytes.IndexRune(v, ':')
	lineStart := pathEnd + 1
	if lineStart <= 0 || lineStart >= len(v) {
		return fmt.Errorf(`cannot parse %s as "file:line:col"`, v)
	}
	lineEnd := lineStart + bytes.IndexRune(v[lineStart:], ':')
	colStart := lineEnd + 1
	if colStart <= 0 || colStart >= len(v) {
		return fmt.Errorf(`cannot parse %s as "file:line:col"`, v)
	}
	p.Path = string(v[:pathEnd])
	lineStr := string(v[lineStart:lineEnd])
	line, err := strconv.Atoi(lineStr)
	if err != nil {
		return fmt.Errorf(`failed to parse line number %s: %s`, lineStr, err)
	}
	p.Line = line - zeroIndexOffset
	colStr := string(v[colStart:])
	col, err := strconv.Atoi(colStr)
	if err != nil {
		return fmt.Errorf(`failed to parse column number %s: %s`, colStr, err)
	}
	p.Col = col - zeroIndexOffset
	return nil
}

// Def represents a result from `guru definition`.
type Def struct {
	Pos ObjectPos `json:"objpos"`
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
	// TODO: get all modified files to pass to guru
	if g.proj != nil && g.cmdr != nil && g.opener != nil && g.ctrl != nil && g.editor != nil {
		return bind.Done
	}
	return bind.Waiting
}

func (g *Godef) Exec() error {
	proj := g.proj.Project()
	lastCaret := g.ctrl.LastCaret()
	// TODO: allow querying with specific build tags.
	// TODO: pass in modified files.
	cmd := exec.Command("guru", "-json", "definition", g.editor.Filepath()+":#"+strconv.Itoa(lastCaret))
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
		g.Err = fmt.Sprintf("guru error: %s", output)
		return err
	}
	var def Def
	if err := json.Unmarshal(output, &def); err != nil {
		g.Err = fmt.Sprintf("failed to parse guru output %s: %s", output, err)
		return err
	}
	g.cmdr.Execute(g.opener.For(focus.Path(def.Pos.Path), focus.Line(def.Pos.Line), focus.Column(def.Pos.Col)))
	return nil
}
