// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/google/gxui"
	"github.com/google/gxui/drivers/gl"
	"github.com/google/gxui/math"
	"github.com/google/gxui/themes/basic"
	"github.com/google/gxui/themes/dark"
)

type context int

const (
	standard context = iota
	function
	method
	quoteString
	backtickString
	character
	comment
	blockComment
	variable
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
	functionColor = gxui.Color{
		R: 0.3,
		G: 0.6,
		B: 0,
		A: 1.0,
	}
	typeColor = gxui.Color{
		R: 0.3,
		G: 0.6,
		B: 0.3,
		A: 1.0,
	}
	stringColor = gxui.Color{
		R: 0,
		G: 0.8,
		B: 0,
		A: 1.0,
	}
	commentColor = gxui.Gray60

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

	filename := ctx.String("file")
	editor.SetSyntaxLayers(layers(filename, editor.Text()))
	editor.SetTabWidth(8)
	suggester := &codeSuggestionProvider{path: filename, editor: editor}
	editor.SetSuggestionProvider(suggester)
	editor.OnTextChanged(func(changes []gxui.TextBoxEdit) {
		// TODO: only update layers that changed.
		editor.SetSyntaxLayers(layers(filename, editor.Text()))
	})
	window.AddChild(editor)

	window.OnClose(driver.Terminate)
	window.SetPadding(math.Spacing{L: 10, T: 10, R: 10, B: 10})
}

func layers(filename, text string) gxui.CodeSyntaxLayers {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, text, 0)
	if err != nil {
		panic(err)
	}
	layers := make(gxui.CodeSyntaxLayers, 0, 100)
	if f.Doc != nil {
		layer := gxui.CreateCodeSyntaxLayer()
		layer.Add(int(f.Doc.Pos()), len(f.Doc.Text()))
		layer.SetColor(commentColor)
		layers = append(layers, layer)
	}
	for _, comment := range f.Comments {
		layer := gxui.CreateCodeSyntaxLayer()
		layer.Add(int(comment.Pos()), len(comment.Text()))
		layer.SetColor(commentColor)
		layers = append(layers, layer)
	}
	for _, decl := range f.Decls {
		var newLayers gxui.CodeSyntaxLayers
		switch src := decl.(type) {
		case *ast.GenDecl:
			newLayers = handleGenDecl(src)
		case *ast.FuncDecl:
			newLayers = handleFuncDecl(src)
		default:
			panic("Unexpected declaration type")
		}
		layers = append(layers, newLayers...)
	}
	for _, something := range f.Unresolved {
		log.Printf("Found unresolved declaration %s at position %d", something.String(), something.Pos())
	}
	return layers
}

func handleFuncDecl(decl *ast.FuncDecl) gxui.CodeSyntaxLayers {
	layerLen := 3 // doc, func keyword, and function name
	if decl.Recv != nil {
		layerLen += decl.Recv.NumFields()
	}
	if decl.Type.Params != nil {
		layerLen += decl.Type.Params.NumFields()
	}
	if decl.Type.Results != nil {
		layerLen += decl.Type.Results.NumFields()
	}
	layers := make(gxui.CodeSyntaxLayers, 0, layerLen)
	if decl.Doc != nil {
		doc := gxui.CreateCodeSyntaxLayer()
		doc.Add(int(decl.Doc.Pos()), int(decl.Doc.End()-decl.Doc.Pos()))
		doc.SetColor(commentColor)
		layers = append(layers, doc)
	}
	funcKeyword := gxui.CreateCodeSyntaxLayer()
	funcKeyword.Add(int(decl.Pos()-1), len("func"))
	funcKeyword.SetColor(keywordColor)
	layers = append(layers, funcKeyword)
	if decl.Recv != nil {
		for _, block := range decl.Recv.List {
			recvType := gxui.CreateCodeSyntaxLayer()
			recvType.Add(int(block.Type.Pos()-1), int(block.Type.End()-block.Type.Pos()))
			recvType.SetColor(typeColor)
			layers = append(layers, recvType)
		}
	}
	funcName := gxui.CreateCodeSyntaxLayer()
	funcName.Add(int(decl.Name.Pos()-1), len(decl.Name.String()))
	funcName.SetColor(functionColor)
	layers = append(layers, funcName)
	if decl.Type.Params != nil {
		for _, block := range decl.Type.Params.List {
			recvType := gxui.CreateCodeSyntaxLayer()
			recvType.Add(int(block.Type.Pos()-1), int(block.Type.End()-block.Type.Pos()))
			recvType.SetColor(typeColor)
			layers = append(layers, recvType)
		}
	}
	if decl.Type.Results != nil {
		for _, block := range decl.Type.Results.List {
			recvType := gxui.CreateCodeSyntaxLayer()
			recvType.Add(int(block.Type.Pos()-1), int(block.Type.End()-block.Type.Pos()))
			recvType.SetColor(typeColor)
			layers = append(layers, recvType)
		}
	}
	return layers
}

func handleGenDecl(decl *ast.GenDecl) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(decl.Specs)+2)
	if decl.Doc != nil {
		doc := gxui.CreateCodeSyntaxLayer()
		doc.Add(int(decl.Doc.Pos()), int(decl.Doc.End()-decl.Doc.Pos()))
		doc.SetColor(commentColor)
		layers = append(layers, doc)
	}
	tok := gxui.CreateCodeSyntaxLayer()
	tok.Add(int(decl.TokPos-1), len(decl.Tok.String()))
	tok.SetColor(keywordColor)
	layers = append(layers, tok)
	for _, spec := range decl.Specs {
		var newLayers gxui.CodeSyntaxLayers
		switch src := spec.(type) {
		case *ast.ValueSpec:
			newLayers = handleValue(src)
		case *ast.ImportSpec:
			newLayers = handleImport(src)
		case *ast.TypeSpec:
			newLayers = handleType(src)
		default:
			panic("Unknown type")
		}
		layers = append(layers, newLayers...)
	}
	return layers
}

func handleValue(val *ast.ValueSpec) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(val.Names)+len(val.Values)+3)
	if val.Doc != nil {
		doc := gxui.CreateCodeSyntaxLayer()
		doc.Add(int(val.Doc.Pos()-1), int(val.Doc.End()-val.Doc.Pos()))
		doc.SetColor(commentColor)
		layers = append(layers, doc)
	}
	if val.Type != nil {
		typ := gxui.CreateCodeSyntaxLayer()
		typ.Add(int(val.Type.Pos()-1), int(val.Type.End()-val.Type.Pos()))
		typ.SetColor(typeColor)
		layers = append(layers, typ)
	}
	if val.Comment != nil {
		doc := gxui.CreateCodeSyntaxLayer()
		doc.Add(int(val.Comment.Pos()-1), int(val.Comment.End()-val.Comment.Pos()))
		doc.SetColor(commentColor)
		layers = append(layers, doc)
	}
	return layers
}

func handleImport(*ast.ImportSpec) gxui.CodeSyntaxLayers {
	return nil
}

func handleType(*ast.TypeSpec) gxui.CodeSyntaxLayers {
	return nil
}
