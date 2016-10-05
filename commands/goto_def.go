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
	"strconv"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/editor"
	"github.com/nelsam/vidar/settings"
)

type LineOpener interface {
	Project() settings.Project
	OpenLine(path string, line, column int)
	CurrentEditor() *editor.CodeEditor
}

type GotoDef struct {
	statusKeeper
}

func NewGotoDef(theme gxui.Theme) *GotoDef {
	return &GotoDef{statusKeeper: statusKeeper{theme: theme}}
}

func (gi *GotoDef) Name() string {
	return "goto-definition"
}

func (gi *GotoDef) Menu() string {
	return "Edit"
}

func (gi *GotoDef) Exec(on interface{}) (executed, consume bool) {
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
		gi.err = fmt.Sprintf("godef error: %s", string(output))
		return true, true
	}
	path, line, column, err := parseGodef(output)
	if err != nil {
		gi.err = err.Error()
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
	return string(pathBytes), oneToZeroBased(line), oneToZeroBased(col), nil
}
