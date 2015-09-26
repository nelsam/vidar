// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package main

import (
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/google/gxui"
	"github.com/google/gxui/drivers/gl"
	"github.com/google/gxui/math"
	"github.com/google/gxui/themes/basic"
	"github.com/google/gxui/themes/dark"
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
}

func main() {
	app.Run(os.Args)
}

func uiMain(driver gxui.Driver) {
	theme := dark.CreateTheme(driver)
	theme.(*basic.Theme).WindowBackground = background
	window := theme.CreateWindow(1600, 800, "GXUI Test Editor")

	// TODO: Check the system's DPI settings for this value
	window.SetScale(1)

	editor := theme.CreateCodeEditor()

	lastMod := time.Time{}
	text := `// Scratch
// This buffer is for jotting down quick notes, but is not saved to disk.
// Use at your own risk!`
	filename := ctx.String("file")
	if filename != "" {
		text = ""
		f, err := os.Open(filename)
		if err != nil && !os.IsNotExist(err) {
			panic(err)
		}
		if err == nil {
			finfo, err := f.Stat()
			if err != nil {
				panic(err)
			}
			lastMod = finfo.ModTime()
			b, err := ioutil.ReadAll(f)
			if err != nil {
				panic(err)
			}
			f.Close()
			text = string(b)
		}
	}
	editor.SetText(text)
	editor.SetDesiredWidth(math.MaxSize.W)

	newLayers, err := layers(filename, editor.Text())
	editor.SetSyntaxLayers(newLayers)
	// TODO: display the error in some pane of the editor
	_ = err

	editor.SetTabWidth(8)
	suggester := &codeSuggestionProvider{path: filename, editor: editor}
	editor.SetSuggestionProvider(suggester)
	editor.OnTextChanged(func(changes []gxui.TextBoxEdit) {
		// TODO: only update layers that changed.
		newLayers, err := layers(filename, editor.Text())
		editor.SetSyntaxLayers(newLayers)
		// TODO: display the error in some pane of the editor
		_ = err
	})
	editor.OnKeyPress(func(event gxui.KeyboardEvent) {
		if event.Modifier.Control() || event.Modifier.Super() {
			switch event.Key {
			case gxui.KeyQ:
				os.Exit(0)
			case gxui.KeyS:
				if !lastMod.IsZero() {
					finfo, err := os.Stat(filename)
					if err != nil {
						panic(err)
					}
					if finfo.ModTime().After(lastMod) {
						panic("Cannot save file: written since last open")
					}
				}
				f, err := os.Create(filename)
				if err != nil {
					panic(err)
				}
				if !strings.HasSuffix(editor.Text(), "\n") {
					editor.SetText(editor.Text() + "\n")
				}
				if _, err := f.WriteString(editor.Text()); err != nil {
					panic(err)
				}
				finfo, err := f.Stat()
				if err != nil {
					panic(err)
				}
				lastMod = finfo.ModTime()
				f.Close()
			}
		}
	})
	window.AddChild(editor)
	gxui.SetFocus(editor)

	window.OnClose(driver.Terminate)
	window.SetPadding(math.Spacing{L: 10, T: 10, R: 10, B: 10})
}
