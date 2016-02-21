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
	"github.com/nelsam/vidar/commands"
	"github.com/nelsam/vidar/controller"
	"github.com/nelsam/vidar/editor"
	"github.com/nelsam/vidar/navigator"
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
	controller := controller.New(driver, theme)

	nav := navigator.New(driver, theme, controller)
	controller.SetNavigator(nav)

	editor := editor.New(driver, theme, theme.DefaultMonospaceFont())
	controller.SetEditor(editor)

	projects := &navigator.Projects{}
	projects.Init(driver, theme)
	nav.Add(projects)

	projTree := &navigator.ProjectTree{}
	projTree.Init(driver, theme)
	nav.Add(projTree)

	commander := commander.New(theme, controller)

	// TODO: Store these in a config file or something
	openFile := commands.NewFileOpener(driver, theme)
	ctrlO := gxui.KeyboardEvent{
		Key:      gxui.KeyO,
		Modifier: gxui.ModControl,
	}
	supO := gxui.KeyboardEvent{
		Key:      gxui.KeyO,
		Modifier: gxui.ModSuper,
	}
	commander.Map(openFile, "File", ctrlO, supO)

	addProject := commands.NewProjectAdder(driver, theme)
	ctrlShiftN := gxui.KeyboardEvent{
		Key:      gxui.KeyN,
		Modifier: gxui.ModControl | gxui.ModShift,
	}
	supShiftN := gxui.KeyboardEvent{
		Key:      gxui.KeyN,
		Modifier: gxui.ModSuper | gxui.ModShift,
	}
	commander.Map(addProject, "File", ctrlShiftN, supShiftN)

	openProj := commands.NewProjectOpener(driver, theme)
	ctrlShiftO := gxui.KeyboardEvent{
		Key:      gxui.KeyO,
		Modifier: gxui.ModControl | gxui.ModShift,
	}
	cmdShiftO := gxui.KeyboardEvent{
		Key:      gxui.KeyO,
		Modifier: gxui.ModSuper | gxui.ModShift,
	}
	commander.Map(openProj, "File", ctrlShiftO, cmdShiftO)

	find := commands.NewFinder(driver, theme)
	ctrlF := gxui.KeyboardEvent{
		Key:      gxui.KeyF,
		Modifier: gxui.ModControl,
	}
	supF := gxui.KeyboardEvent{
		Key:      gxui.KeyF,
		Modifier: gxui.ModSuper,
	}
	commander.Map(find, "Edit", ctrlF, supF)

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
		commander.Controller().Editor().Open(filepath, 0)
	}

	window.OnClose(driver.Terminate)
	window.SetPadding(math.Spacing{L: 10, T: 10, R: 10, B: 10})
}
