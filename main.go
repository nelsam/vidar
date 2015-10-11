// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package main

import (
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/drivers/gl"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/gxui/themes/dark"
	"github.com/nelsam/gxui_playground/commander"
	"github.com/nelsam/gxui_playground/editors"
)

var (
	background = gxui.Gray10

	workingDir string
	app        *cli.App
	ctx        *cli.Context
)

func init() {
	app = cli.NewApp()
	app.Name = "Experimental Go Editor"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "file",
			Usage: "The file to open on startup",
		},
	}
	app.Action = func(context *cli.Context) {
		ctx = context
		gl.StartDriver(uiMain)
	}
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	workingDir = dir
}

func main() {
	app.Run(os.Args)
}

func setFile(editor editors.Editor, filepath string) {
	editor.Open(filepath)
}

func uiMain(driver gxui.Driver) {
	filepath := ctx.String("file")
	if filepath != "" {
		workingDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		filepath = path.Join(workingDir, filepath)
	}

	theme := dark.CreateTheme(driver).(*basic.Theme)
	theme.WindowBackground = background

	window := theme.CreateWindow(1600, 800, "GXUI Test Editor")
	cmdBox := commander.New(driver, theme)
	editor := editors.New(driver, theme, theme.DefaultMonospaceFont(), cmdBox.PromptOpenFile)

	window.OnKeyDown(func(event gxui.KeyboardEvent) {
		if (event.Modifier.Control() || event.Modifier.Super()) && event.Key == gxui.KeyQ {
			os.Exit(0)
		}
	})
	// TODO: Check the system's DPI settings for this value
	window.SetScale(1)

	cmdBox.CurrentFile(filepath)

	layout := theme.CreateLinearLayout()
	layout.SetDirection(gxui.BottomToTop)
	layout.SetHorizontalAlignment(gxui.AlignLeft)
	layout.AddChild(cmdBox)
	layout.AddChild(editor)

	window.AddChild(layout)
	editor.Open(filepath)

	window.OnClose(driver.Terminate)
	window.SetPadding(math.Spacing{L: 10, T: 10, R: 10, B: 10})
}
