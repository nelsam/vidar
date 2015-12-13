// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package main

import (
	"os"
	"path"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/drivers/gl"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/gxui/themes/dark"
	"github.com/nelsam/vidar/commander"
	"github.com/spf13/cobra"
)

var (
	background = gxui.Gray10

	workingDir string
	cmd        *cobra.Command
	files      []string
)

func init() {
	cmd = &cobra.Command{
		Use:   "vidar [files...]",
		Short: "An experimental Go editor",
		Long:  "An editor for Go code, still in its infancy.  Basic editing of Go code is mostly complete, but there's still a high potential for data loss.",
		Run: func(cmd *cobra.Command, args []string) {
			files = args
			gl.StartDriver(uiMain)
		},
	}
}

func main() {
	cmd.Execute()
}

func uiMain(driver gxui.Driver) {
	theme := dark.CreateTheme(driver).(*basic.Theme)
	theme.WindowBackground = background

	// TODO: figure out a better way to get this resolution
	window := theme.CreateWindow(1600, 800, "Vidar - GXUI Go Editor")
	commander := commander.New(driver, theme, theme.DefaultMonospaceFont())

	window.OnKeyDown(func(event gxui.KeyboardEvent) {
		if (event.Modifier.Control() || event.Modifier.Super()) && event.Key == gxui.KeyQ {
			os.Exit(0)
		}
		commander.KeyDown(event)
	})
	window.OnKeyUp(func(event gxui.KeyboardEvent) {
		commander.KeyPress(event)
	})

	// TODO: Check the system's DPI settings for this value
	window.SetScale(1)

	window.AddChild(commander)
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		filepath := path.Join(workingDir, file)
		commander.Controller().Editor().Open(filepath)
	}

	window.OnClose(driver.Terminate)
	window.SetPadding(math.Spacing{L: 10, T: 10, R: 10, B: 10})
}
