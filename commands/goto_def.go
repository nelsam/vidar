// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"bytes"
	"log"
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
}

func NewGotoDef() *GotoDef {
	return &GotoDef{}
}

func (gi *GotoDef) Start(gxui.Control) gxui.Control {
	return nil
}

func (gi *GotoDef) Name() string {
	return "goto-definition"
}

func (gi *GotoDef) Next() gxui.Focusable {
	return nil
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
		cmd.Env[0] += ":" + filepath.Join(proj.Gopath, "bin")
		cmd.Env = append(cmd.Env, "GOPATH="+proj.Gopath)
	}
	output, err := cmd.Output()
	if err != nil {
		// TODO: report this error to the user via the UI
		log.Printf("Received error from godef: %s", err)
		log.Print("Error output:")
		log.Print(errBuffer.String())
		return true, true
	}
	path, line, column := parseGodef(output)
	opener.OpenLine(path, line, column)
	return true, true
}

func parseGodef(output []byte) (path string, line, column int) {
	values := bytes.Split(bytes.TrimSpace(output), []byte{':'})
	if len(values) != 3 {
		log.Printf("godef output %s not understood", string(output))
		return "", 0, 0
	}
	pathBytes, lineBytes, colBytes := values[0], values[1], values[2]
	line, err := strconv.Atoi(string(lineBytes))
	if err != nil {
		log.Printf("godef output: %s is not a line number", string(lineBytes))
		return "", 0, 0
	}
	col, err := strconv.Atoi(string(colBytes))
	if err != nil {
		log.Printf("godef output: %s is not a column number", string(colBytes))
		return "", 0, 0
	}
	return string(pathBytes), oneToZeroBased(line), oneToZeroBased(col)
}
