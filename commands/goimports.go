package commands

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"path"

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
	cmd := &exec.Cmd{
		Path:  "goimports",
		Stdin: bytes.NewBufferString(editor.Text()),
	}
	if proj.Gopath != "" {
		gopathGoimports := path.Join(proj.Gopath, "bin", "goimports")
		if _, err := os.Stat(gopathGoimports); err == nil {
			cmd.Path = gopathGoimports
		}
		envPath := os.Getenv("PATH")
		cmd.Env = []string{
			"GOPATH=" + proj.Gopath,
			"PATH=" + envPath + ":" + path.Join(proj.Gopath, "/bin"),
		}
	}
	formatted, err := cmd.Output()
	if err != nil {
		// TODO: report this error to the user via the UI
		log.Printf("Received error from goimports: %s", err)
		return true, true
	}
	editor.SetText(string(formatted))
	return true, true
}
