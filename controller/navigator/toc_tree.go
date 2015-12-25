package navigator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
)

type genericTreeNode struct {
	gxui.AdapterBase

	name     string
	path     string
	color    gxui.Color
	children []gxui.TreeNode
}

func (t genericTreeNode) Count() int {
	return len(t.children)
}

func (t genericTreeNode) ItemIndex(item gxui.AdapterItem) int {
	for i, child := range t.children {
		if strings.HasPrefix(item.(string), child.Item().(string)) {
			return i
		}
	}
	return -1
}

func (t genericTreeNode) Size(theme gxui.Theme) math.Size {
	return math.Size{
		W: 20 * theme.DefaultMonospaceFont().GlyphMaxSize().W,
		H: theme.DefaultMonospaceFont().GlyphMaxSize().H,
	}
}

func (t genericTreeNode) Item() gxui.AdapterItem {
	return t.path
}

func (t genericTreeNode) NodeAt(index int) gxui.TreeNode {
	return t.children[index]
}

func (t genericTreeNode) Create(theme gxui.Theme) gxui.Control {
	label := theme.CreateLabel()
	label.SetText(t.name)
	label.SetColor(t.color)
	return label
}

type Location struct {
	filename string
	pos      int
}

func (l Location) File() string {
	return l.filename
}

func (l Location) Pos() int {
	return l.pos
}

type Name struct {
	genericTreeNode
	Location
}

type TOC struct {
	genericTreeNode

	pkg  string
	path string
}

func NewTOC(path string) *TOC {
	toc := &TOC{}
	toc.Init(path)
	return toc
}

func (t *TOC) Init(path string) {
	t.path = path
	t.Reload()
}

func (t *TOC) Reload() {
	pkgs, err := parser.ParseDir(token.NewFileSet(), t.path, nil, 0)
	if err != nil {
		return
	}
	for name, pkg := range pkgs {
		if strings.HasSuffix(name, "_test") {
			continue
		}
		t.pkg = name
		t.addPkg(pkg)
	}
}

func (t *TOC) addPkg(pkg *ast.Package) {
	var (
		consts  []gxui.TreeNode
		vars    []gxui.TreeNode
		typeMap = make(map[string]*Name)
		types   []gxui.TreeNode
		funcs   []gxui.TreeNode
	)
	for filename, f := range pkg.Files {
		if strings.HasSuffix(filename, "_test.go") {
			continue
		}
		for _, decl := range f.Decls {
			switch src := decl.(type) {
			case *ast.GenDecl:
				switch src.Tok.String() {
				case "const":
					consts = append(consts, valueNamesFrom(filename, "constants", src.Specs)...)
				case "var":
					vars = append(vars, valueNamesFrom(filename, "global vars", src.Specs)...)
				case "type":
					// I have yet to see a case where a type declaration has more than one Specs.
					typeSpec := src.Specs[0].(*ast.TypeSpec)
					typeName := typeSpec.Name.String()

					// We can't guarantee that the type declaration was found before method
					// declarations, so the value may already exist in the map.
					typ, ok := typeMap[typeName]
					if !ok {
						typ = &Name{}
						typeMap[typeName] = typ
					}
					typ.name = typeName
					typ.path = "types." + typ.name
					typ.color = dirColor
					typ.filename = filename
					typ.pos = int(typeSpec.Pos())
					types = append(types, typ)
				}
			case *ast.FuncDecl:
				var name Name
				name.name = src.Name.String()
				name.path = "funcs." + name.name
				name.color = dirColor
				name.filename = filename
				name.pos = int(src.Pos())
				if src.Recv == nil {
					funcs = append(funcs, name)
					continue
				}
				recvTyp := src.Recv.List[0].Type
				if starExpr, ok := recvTyp.(*ast.StarExpr); ok {
					recvTyp = starExpr.X
				}
				recvTypeName := recvTyp.(*ast.Ident).String()
				typ, ok := typeMap[recvTypeName]
				if !ok {
					typ = &Name{}
					typeMap[recvTypeName] = typ
				}
				name.path = "types." + recvTypeName + "." + name.name
				typ.children = append(typ.children, name)
			}
		}
	}
	t.children = append(t.children,
		genericTreeNode{name: "constants", path: "constants", color: dirColor, children: consts},
		genericTreeNode{name: "global vars", path: "global vars", color: dirColor, children: vars},
		genericTreeNode{name: "types", path: "types", color: dirColor, children: types},
		genericTreeNode{name: "funcs", path: "funcs", color: dirColor, children: funcs},
	)
}

func valueNamesFrom(filename, parentName string, specs []ast.Spec) (names []gxui.TreeNode) {
	for _, spec := range specs {
		valSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, name := range valSpec.Names {
			var newName Name
			newName.name = name.String()
			newName.path = parentName + "." + newName.name
			newName.color = dirColor
			newName.filename = filename
			newName.pos = int(name.Pos())
			names = append(names, newName)
		}
	}
	return
}
