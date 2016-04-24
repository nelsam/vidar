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

	"github.com/nelsam/gxui"
)

type GoImports struct {
}

func NewGoImports() *GoImports {
	return &GoImports{}
}

func (gi *GoImports) Start(gxui.Control) gxui.Control {
	return nil
}

func (gi *GoImports) Name() string {
	return "goimports"
}

func (gi *GoImports) Next() gxui.Focusable {
	return nil
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
		cmd.Env[0] += ":" + filepath.Join(proj.Gopath, "bin")
		cmd.Env = append(cmd.Env, "GOPATH="+proj.Gopath)
	}
	formatted, err := cmd.Output()
	if err != nil {
		// TODO: report this error to the user via the UI
		log.Printf("Received error from goimports: %s", err)
		log.Print("Error output:")
		log.Print(errBuffer.String())
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
