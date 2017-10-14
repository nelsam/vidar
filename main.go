// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package main

import (
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/drivers/gl"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/gxui/themes/dark"
	"github.com/nelsam/vidar/commander"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commands"
	"github.com/nelsam/vidar/commands/input"
	"github.com/nelsam/vidar/controller"
	"github.com/nelsam/vidar/editor"
	"github.com/nelsam/vidar/navigator"
	"github.com/nelsam/vidar/plugin"
	"github.com/nelsam/vidar/settings"
	"github.com/nelsam/vidar/theme"
	"github.com/spf13/cobra"
	"github.com/tmc/fonts"
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
		Long: "An editor for Go code, still in its infancy.  " +
			"Basic editing of Go code is mostly complete, but " +
			"panics still happen and can result in the loss of " +
			"unsaved work.",
		Run: func(cmd *cobra.Command, args []string) {
			files = args
			gl.StartDriver(uiMain)
		},
	}
}

func main() {
	cmd.Execute()
}

func font(driver gxui.Driver) gxui.Font {
	desiredFonts := settings.DesiredFonts()
	if len(desiredFonts) == 0 {
		return nil
	}
	var (
		font       settings.Font
		fontReader io.Reader
		err        error
	)
	for _, font = range desiredFonts {
		fontReader, err = fonts.Load(font.Name)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil
	}
	if closer, ok := fontReader.(io.Closer); ok {
		defer closer.Close()
	}
	fontBytes, err := ioutil.ReadAll(fontReader)
	if err != nil {
		log.Printf("Failed to read font file: %s", err)
		return nil
	}
	gFont, err := driver.CreateFont(fontBytes, font.Size)
	if err != nil {
		log.Printf("Could not parse font: %s", err)
		return nil
	}
	return gFont
}

func uiMain(driver gxui.Driver) {
	gTheme := dark.CreateTheme(driver).(*basic.Theme)
	font := font(driver)
	if font == nil {
		font = gTheme.DefaultMonospaceFont()
	}
	gTheme.SetDefaultMonospaceFont(font)
	gTheme.SetDefaultFont(font)
	gTheme.WindowBackground = background

	// TODO: figure out a better way to get this resolution
	window := gTheme.CreateWindow(1600, 800, "Vidar - GXUI Go Editor")
	controller := controller.New(driver, gTheme)

	// Bindings should be added immediately after creating the commander,
	// since other types rely on the bindings having been bound.
	cmdr := commander.New(driver, gTheme, controller)
	bindings := []bind.Bindable{input.New(driver, cmdr)}
	for _, c := range commands.Commands(cmdr, driver, gTheme) {
		bindings = append(bindings, c)
	}
	for _, h := range commands.Hooks(cmdr, driver, gTheme) {
		bindings = append(bindings, h)
	}
	bindings = append(bindings, plugin.Bindables(cmdr, driver, gTheme)...)
	cmdr.Push(bindings...)

	nav := navigator.New(driver, gTheme)
	controller.SetNavigator(nav)

	editor := editor.New(driver, window, gTheme, theme.Default, gTheme.DefaultMonospaceFont())
	controller.SetEditor(editor)

	projTree := navigator.NewProjectTree(cmdr, driver, window, gTheme)
	projects := navigator.NewProjectsPane(cmdr, driver, gTheme, projTree.Frame())

	nav.Add(projects)
	nav.Add(projTree)

	nav.Resize(window.Size().H)
	window.OnResize(func() {
		nav.Resize(window.Size().H)
	})

	// TODO: Check the system's DPI settings for this value
	window.SetScale(1)

	window.AddChild(cmdr)

	window.OnKeyDown(func(event gxui.KeyboardEvent) {
		if (event.Modifier.Control() || event.Modifier.Super()) && event.Key == gxui.KeyQ {
			os.Exit(0)
		}
		if event.Modifier == 0 && event.Key == gxui.KeyF11 {
			window.SetFullscreen(!window.Fullscreen())
		}
		if window.Focus() == nil {
			cmdr.KeyDown(event)
		}
	})
	window.OnKeyUp(func(event gxui.KeyboardEvent) {
		if window.Focus() == nil {
			cmdr.KeyPress(event)
		}
	})

	opener := cmdr.Bindable("open-file").(*commands.FileOpener)
	for _, file := range files {
		filepath, err := filepath.Abs(file)
		if err != nil {
			log.Printf("Failed to get path: %s", err)
		}
		opener.Start(nil)
		opener.SetLocation(filepath, token.Position{})
		cmdr.Execute(opener)
	}

	window.OnClose(driver.Terminate)
	window.SetPadding(math.Spacing{L: 10, T: 10, R: 10, B: 10})
}
