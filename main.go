// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package main

import (
	"io/ioutil"
	"log"
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

func setFile(editor *statefulEditor, filepath string) {
	editor.filepath = filepath
	editor.lastModified = time.Time{}

	f, err := os.Open(filepath)
	if os.IsNotExist(err) {
		editor.SetText("")
		return
	}
	if err != nil {
		panic(err)
	}
	finfo, err := f.Stat()
	if err != nil {
		panic(err)
	}
	editor.lastModified = finfo.ModTime()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	f.Close()
	editor.SetText(string(b))
}

func uiMain(driver gxui.Driver) {
	theme := dark.CreateTheme(driver)
	theme.(*basic.Theme).WindowBackground = background

	window := theme.CreateWindow(1600, 800, "GXUI Test Editor")
	cmdBox := theme.CreateTextBox()
	editor := &statefulEditor{
		CodeEditor: theme.CreateCodeEditor(),
	}

	window.OnKeyDown(func(event gxui.KeyboardEvent) {
		if (event.Modifier.Control() || event.Modifier.Super()) && event.Key == gxui.KeyQ {
			os.Exit(0)
		}
	})
	// TODO: Check the system's DPI settings for this value
	window.SetScale(1)

	cmdBox.SetDesiredWidth(math.MaxSize.W)
	cmdBox.SetMultiline(false)
	cmdBox.OnKeyPress(func(event gxui.KeyboardEvent) {
		if event.Key == gxui.KeyEnter {
			setFile(editor, cmdBox.Text())
			cmdBox.SetText("")
			gxui.SetFocus(editor)
		}
	})

	editor.SetText(`// Scratch
// This buffer is for jotting down quick notes, but is not saved to disk.
// Use at your own risk!`)
	filepath := ctx.String("file")
	if filepath != "" {
		setFile(editor, filepath)
	}
	editor.SetDesiredWidth(math.MaxSize.W)

	newLayers, err := layers(editor.filepath, editor.Text())
	editor.SetSyntaxLayers(newLayers)
	// TODO: display the error in some pane of the editor
	_ = err

	editor.SetTabWidth(8)
	suggester := &codeSuggestionProvider{path: editor.filepath, editor: editor}
	editor.SetSuggestionProvider(suggester)
	editor.OnTextChanged(func(changes []gxui.TextBoxEdit) {
		editor.hasChanges = true
		// TODO: only update layers that changed.
		newLayers, err := layers(editor.filepath, editor.Text())
		editor.SetSyntaxLayers(newLayers)
		// TODO: display the error in some pane of the editor
		_ = err
	})
	editor.OnKeyPress(func(event gxui.KeyboardEvent) {
		if event.Modifier.Control() || event.Modifier.Super() {
			switch event.Key {
			case gxui.KeyS:
				if !editor.lastModified.IsZero() {
					finfo, err := os.Stat(editor.filepath)
					if err != nil {
						panic(err)
					}
					if finfo.ModTime().After(editor.lastModified) {
						panic("Cannot save file: written since last open")
					}
				}
				f, err := os.Create(editor.filepath)
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
				editor.lastModified = finfo.ModTime()
				f.Close()
				editor.hasChanges = false
			case gxui.KeyO:
				if editor.hasChanges {
					log.Printf("WARNING: Opening new file without saving changes")
				}
				gxui.SetFocus(cmdBox)
			}
		}
	})
	layout := theme.CreateLinearLayout()
	layout.SetDirection(gxui.BottomToTop)
	layout.SetHorizontalAlignment(gxui.AlignLeft)
	layout.AddChild(cmdBox)
	layout.AddChild(editor)

	window.AddChild(layout)
	gxui.SetFocus(editor)

	window.OnClose(driver.Terminate)
	window.SetPadding(math.Spacing{L: 10, T: 10, R: 10, B: 10})
}
