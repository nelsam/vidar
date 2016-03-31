package commands

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/editor"
	"github.com/nelsam/vidar/settings"
)

type ProjectFinder interface {
	CurrentEditor() *editor.CodeEditor
	Project() settings.Project
}

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
	proj := finder.Project()
	cmd := exec.Command("goimports", "-srcdir", filepath.Dir(editor.Filepath()))
	cmd.Stdin = bytes.NewBufferString(editor.Text())
	errBuffer := &bytes.Buffer{}
	cmd.Stderr = errBuffer
	cmd.Env = []string{"PATH=" + os.Getenv("PATH")}
	if proj.Gopath != "" {
		cmd.Env[0] += ":" + path.Join(proj.Gopath, "/bin")
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
	editor.SetText(string(formatted))
	return true, true
}
