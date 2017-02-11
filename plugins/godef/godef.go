// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// Package godef contains logic for interacting with the godef
// command line tool.  It can be imported directly or used as a
// plugin.
package godef

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander"
	"github.com/nelsam/vidar/editor"
	"github.com/nelsam/vidar/settings"
)

type LineOpener interface {
	Project() settings.Project
	OpenLine(path string, line, column int)
	CurrentEditor() *editor.CodeEditor
}

type Godef struct {
	commander.GenericStatuser
}

func New(theme gxui.Theme) *Godef {
	g := &Godef{}
	g.Theme = theme
	return g
}

func (gi *Godef) Name() string {
	return "goto-definition"
}

func (gi *Godef) Menu() string {
	return "Edit"
}

func (gi *Godef) Exec(on interface{}) (executed, consume bool) {
	opener, ok := on.(LineOpener)
	if !ok {
		return false, false
	}
	editor := opener.CurrentEditor()
	if editor == nil {
		return true, true
	}
	proj := opener.Project()
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
		gi.Err = fmt.Sprintf("godef error: %s", string(output))
		return true, true
	}
	path, line, column, err := parseGodef(output)
	if err != nil {
		gi.Err = err.Error()
		return true, true
	}
	opener.OpenLine(path, line, column)
	return true, true
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
