package main

import (
	"io/ioutil"
	"os"
	"unicode"

	"github.com/codegangsta/cli"
	"github.com/google/gxui"
	"github.com/google/gxui/drivers/gl"
	"github.com/google/gxui/math"
	"github.com/google/gxui/themes/basic"
	"github.com/google/gxui/themes/dark"
)

type identifierType int

const (
	invalid identifierType = iota
	builtin
	operator
	keyword
	function
	method
	typeDef
	typeName
	constructor
)

var (
	builtins = []string{
		"append",
		"cap",
		"close",
		"complex",
		"copy",
		"delete",
		"imag",
		"len",
		"make",
		"new",
		"panic",
		"print",
		"println",
		"real",
		"recover",
	}

	keywords = []string{
		"break",
		"default",
		"func",
		"interface",
		"select",
		"case",
		"defer",
		"go",
		"map",
		"struct",
		"chan",
		"else",
		"goto",
		"package",
		"switch",
		"const",
		"fallthrough",
		"if",
		"range",
		"type",
		"continue",
		"for",
		"import",
		"return",
		"var",
	}

	background   = gxui.Gray10
	builtinColor = gxui.Color{
		R: 0.9,
		G: 0.3,
		B: 0,
		A: 1.0,
	}
	keywordColor = gxui.Color{
		R: 0,
		G: 0.6,
		B: 0.8,
		A: 1.0,
	}

	app *cli.App
	ctx *cli.Context
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
	b, err := ioutil.ReadFile(ctx.String("file"))
	if err != nil {
		panic(err)
	}
	editor.SetText(string(b))
	editor.SetDesiredWidth(math.MaxSize.W)

	editor.SetSyntaxLayers(layers(editor.Text()))
	editor.SetTabWidth(8)
	editor.OnTextChanged(func(changes []gxui.TextBoxEdit) {
		// TODO: only update layers that changed.
		editor.SetSyntaxLayers(layers(editor.Text()))
	})
	window.AddChild(editor)

	window.OnClose(driver.Terminate)
	window.SetPadding(math.Spacing{L: 10, T: 10, R: 10, B: 10})
}

func layers(text string) gxui.CodeSyntaxLayers {
	var (
		start int
	)
	syntaxes := make(gxui.CodeSyntaxLayers, 0, 100)
	for i, r := range text {
		if isSeparator(r) {
			word := text[start:i]
			syntax := gxui.CreateCodeSyntaxLayer()
			syntax.Add(start, len(word))
			start = i + 1
			if len(word) == 0 {
				continue
			}
			if isBuiltin(word) {
				syntax.SetColor(builtinColor)
			} else if isKeyword(word) {
				syntax.SetColor(keywordColor)
			}
			syntaxes = append(syntaxes, syntax)
		}
	}
	return syntaxes
}

func isSeparator(r rune) bool {
	switch r {
	case '_', '"', '\'', '`':
		return false
	}
	return !(unicode.IsLetter(r) || unicode.IsNumber(r))
}

func inSlice(identifier string, slice []string) bool {
	for _, value := range slice {
		if identifier == value {
			return true
		}
	}
	return false
}

func isBuiltin(identifier string) bool {
	return inSlice(identifier, builtins)
}

func isKeyword(identifier string) bool {
	return inSlice(identifier, keywords)
}
