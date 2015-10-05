// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/drivers/gl"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/gxui/themes/dark"
	"github.com/nelsam/gxui_playground/editors"
	"github.com/nelsam/gxui_playground/suggestions"
	"github.com/nelsam/gxui_playground/syntax"
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

func setFile(editor *editors.Editor, filepath string) {
	editor.Filepath = filepath
	editor.LastModified = time.Time{}

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
	editor.LastModified = finfo.ModTime()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	f.Close()
	editor.SetText(string(b))
}

func uiMain(driver gxui.Driver) {
	theme := dark.CreateTheme(driver).(*basic.Theme)
	theme.WindowBackground = background

	window := theme.CreateWindow(1600, 800, "GXUI Test Editor")
	cmdBox := theme.CreateTextBox()
	editor := editors.New(driver, theme, theme.DefaultMonospaceFont()).(*editors.Editor)
	suggester := suggestions.NewGoCodeProvider(editor).(*suggestions.GoCodeProvider)

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
			suggester.Path = path.Join(workingDir, editor.Filepath)
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
		suggester.Path = path.Join(workingDir, editor.Filepath)
	}
	editor.SetDesiredWidth(math.MaxSize.W)

	newLayers, err := syntax.Layers(editor.Filepath, editor.Text())
	editor.SetSyntaxLayers(newLayers)
	// TODO: display the error in some pane of the editor
	_ = err

	editor.SetTabWidth(8)
	editor.SetSuggestionProvider(suggester)
	editor.OnTextChanged(func(changes []gxui.TextBoxEdit) {
		editor.HasChanges = true
		// TODO: only update layers that changed.
		newLayers, err := syntax.Layers(editor.Filepath, editor.Text())
		editor.SetSyntaxLayers(newLayers)
		// TODO: display the error in some pane of the editor
		_ = err
	})
	editor.OnKeyPress(func(event gxui.KeyboardEvent) {
		if event.Modifier.Control() || event.Modifier.Super() {
			switch event.Key {
			case gxui.KeyS:
				if !editor.LastModified.IsZero() {
					finfo, err := os.Stat(editor.Filepath)
					if err != nil {
						panic(err)
					}
					if finfo.ModTime().After(editor.LastModified) {
						panic("Cannot save file: written since last open")
					}
				}
				f, err := os.Create(editor.Filepath)
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
				editor.LastModified = finfo.ModTime()
				f.Close()
				editor.HasChanges = false
			case gxui.KeyO:
				if editor.HasChanges {
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
