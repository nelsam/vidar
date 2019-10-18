// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/command/lsp"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/plugin/command"
	"github.com/nelsam/vidar/setting"
)

type GolangHook struct {
	proj setting.Project
}

func (h GolangHook) Name() string {
	return "golang-hook"
}

func (h GolangHook) OpNames() []string {
	return []string{"focus-location", "project-change"}
}

func (h *GolangHook) SetProject(proj setting.Project) {
	h.proj = proj
}

func (h GolangHook) FileBindables(path string) []bind.Bindable {
	if !strings.HasSuffix(path, ".go") {
		return nil
	}
	var env []string
	if h.proj.Gopath != "" {
		env = append(env,
			"GOPATH="+h.proj.Gopath,
			"PATH="+h.proj.Gopath+"/bin:"+os.Getenv("PATH"),
		)
	}
	conn, err := lsp.New(lsp.WithCommand(exec.Command("go-langserver", "-trace"), env...))
	if err := cmd.Start(); err != nil {
		log.Printf("Error: could not start gols: %s", err)
		return nil
	}
	conn, err := lsp.New(lsp.WithConn(&golsconn))
	if err != nil {
		log.Printf("Error: could not create connection to gols: %s", err)
		return nil
	}
	rsp, err := conn.Initialize(map[string]interface{}{
		"gocodeCompletionEnabled": true,
		"formatTool":              "goimports",
		"lintTool":                "golint",
	})
	if err != nil {
		log.Printf("could not initialize go lsp: %s", err)
		return nil
	}
	log.Printf("got response from gols:")
	pretty(rsp)
	return []bind.Bindable{
		&lspsyntax{conn: conn},
	}
}

// Bindables is the main entry point to the command.
func Bindables(cmdr command.Commander, driver gxui.Driver, theme gxui.Theme) []bind.Bindable {
	return []bind.Bindable{
		&GoLSPClient{},
	}
}

func pretty(v interface{}) {
	fmt.Println(prettyv("", reflect.ValueOf(v)))
}

func prettyv(prefix string, v reflect.Value) string {
	if !v.IsValid() {
		return prefix + "<nil>\n"
	}
	switch v.Kind() {
	case reflect.Interface:
		return prettyv(prefix, v.Elem())
	case reflect.Struct:
		out := prefix + v.Type().Name() + " {\n"
		subprefix := prefix + "    "
		for i := 0; i < v.NumField(); i++ {
			out += subprefix + v.Type().Field(i).Name + ": "
			out += prettyv(subprefix, v.Field(i))[len(subprefix):]
		}
		out += prefix + "}\n"
		return out
	default:
		return fmt.Sprintf("%s%#v\n", prefix, v.Interface())
	}
}
