package navigator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"path"
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
)

var (
	genericColor = gxui.Color{
		R: 0.6,
		G: 0.8,
		B: 1,
		A: 1,
	}
	nameColor = gxui.Color{
		R: 0.6,
		G: 1,
		B: 0.5,
		A: 1,
	}
	skippableColor = gxui.Color{
		R: 0.9,
		G: 0.6,
		B: 0.8,
		A: 1,
	}

	// Since the const values aren't exported by go/build, I've just copied them
	// from https://github.com/golang/go/blob/master/src/go/build/syslist.go
	gooses   = []string{"android", "darwin", "dragonfly", "freebsd", "linux", "nacl", "netbsd", "openbsd", "plan9", "solaris", "windows"}
	goarches = []string{"386", "amd64", "amd64p32", "arm", "armbe", "arm64", "arm64be", "ppc64", "ppc64le", "mips", "mipsle", "mips64", "mips64le", "mips64p32", "mips64p32le", "ppc", "s390", "s390x", "sparc", "sparc64"}
)

func zeroBasedPos(pos token.Pos) int {
	return int(pos) - 1
}

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
	path := item.(string)
	for i, child := range t.children {
		childPath := child.Item().(string)
		if path == childPath || strings.HasPrefix(path, childPath+".") {
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

// skippingTreeNode is a gxui.TreeNode that will skip to the child
// node if there is exactly one child node on implementations of all
// gxui.TreeNode methods.  This means that any number of nested
// skippingTreeNodes will display as a single node, so long as no
// child contains more than one child.
type skippingTreeNode struct {
	genericTreeNode
}

func (t skippingTreeNode) Count() int {
	if len(t.children) == 1 {
		return t.children[0].Count()
	}
	return t.genericTreeNode.Count()
}

func (t skippingTreeNode) ItemIndex(item gxui.AdapterItem) int {
	if len(t.children) == 1 {
		return t.children[0].ItemIndex(item)
	}
	return t.genericTreeNode.ItemIndex(item)
}

func (t skippingTreeNode) Item() gxui.AdapterItem {
	if len(t.children) == 1 {
		return t.children[0].Item()
	}
	return t.genericTreeNode.Item()
}

func (t skippingTreeNode) NodeAt(index int) gxui.TreeNode {
	if len(t.children) == 1 {
		return t.children[0].NodeAt(index)
	}
	return t.genericTreeNode.NodeAt(index)
}

func (t skippingTreeNode) Create(theme gxui.Theme) gxui.Control {
	if len(t.children) == 1 {
		return t.children[0].Create(theme)
	}
	return t.genericTreeNode.Create(theme)
}

type Location struct {
	filepath string
	pos      int
}

func (l Location) File() string {
	return l.filepath
}

func (l Location) Pos() int {
	return l.pos
}

type Name struct {
	genericTreeNode
	Location
}

type TOC struct {
	skippingTreeNode

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
	pkgs, err := parser.ParseDir(token.NewFileSet(), t.path, nil, parser.ParseComments)
	if err != nil {
		log.Printf("Error parsing dir: %s", err)
	}
	for _, pkg := range pkgs {
		t.children = append(t.children, t.parsePkg(pkg))
	}
}

func (t *TOC) parsePkg(pkg *ast.Package) genericTreeNode {
	var (
		pkgNode = genericTreeNode{name: pkg.Name, path: pkg.Name, color: skippableColor}

		files   []gxui.TreeNode
		consts  []gxui.TreeNode
		vars    []gxui.TreeNode
		typeMap = make(map[string]*Name)
		types   []gxui.TreeNode
		funcs   []gxui.TreeNode
	)
	for filepath, f := range pkg.Files {
		var fileNode Name
		filename := path.Base(filepath)
		fileNode.name = filename
		fileNode.path = pkg.Name + ".files." + filepath
		fileNode.color = nameColor
		fileNode.filepath = filepath
		files = append(files, fileNode)

		buildTags := findBuildTags(filename, f)
		buildTagLine := strings.Join(buildTags, " ")
		for _, decl := range f.Decls {
			switch src := decl.(type) {
			case *ast.GenDecl:
				switch src.Tok.String() {
				case "const":
					consts = append(consts, valueNamesFrom(filepath, buildTagLine, pkg.Name+".constants", src.Specs)...)
				case "var":
					vars = append(vars, valueNamesFrom(filepath, buildTagLine, pkg.Name+".global vars", src.Specs)...)
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
					if buildTagLine != "" {
						typ.name = fmt.Sprintf("%s (%s)", typ.name, buildTagLine)
					}
					typ.path = pkg.Name + ".types." + typ.name
					typ.color = nameColor
					typ.filepath = filepath
					typ.pos = zeroBasedPos(typeSpec.Pos())
					types = append(types, typ)
				}
			case *ast.FuncDecl:
				if src.Name.String() == "init" {
					// There can be multiple inits in the package, so this
					// doesn't really help us in the TOC.
					continue
				}
				var name Name
				name.name = src.Name.String()
				if buildTagLine != "" {
					name.name = fmt.Sprintf("%s (%s)", name.name, buildTagLine)
				}
				name.path = pkg.Name + ".funcs." + name.name
				name.color = nameColor
				name.filepath = filepath
				name.pos = zeroBasedPos(src.Pos())
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
				name.path = pkg.Name + ".types." + recvTypeName + "." + name.name
				typ.children = append(typ.children, name)
			}
		}
	}
	pkgNode.children = []gxui.TreeNode{
		genericTreeNode{name: "files", path: pkg.Name + ".files", color: genericColor, children: files},
		genericTreeNode{name: "constants", path: pkg.Name + ".constants", color: genericColor, children: consts},
		genericTreeNode{name: "global vars", path: pkg.Name + ".global vars", color: genericColor, children: vars},
		genericTreeNode{name: "types", path: pkg.Name + ".types", color: genericColor, children: types},
		genericTreeNode{name: "funcs", path: pkg.Name + ".funcs", color: genericColor, children: funcs},
	}
	return pkgNode
}

func findBuildTags(filename string, file *ast.File) (tags []string) {
	fileTag := parseFileTag(filename)
	if fileTag != "" {

		tags = append(tags, fileTag)
	}

	for _, commentBlock := range file.Comments {
		if commentBlock.End() >= file.Package-1 {
			// A build tag comment *must* have an empty line between it
			// and the `package` declaration.
			continue
		}

		for _, comment := range commentBlock.List {
			newTags := parseTag(comment)
			tags = applyTags(newTags, tags)

		}
	}
	return tags
}

func applyTags(newTags, prevTags []string) (combinedTags []string) {
	if len(prevTags) == 0 {
		return newTags
	}
	if len(newTags) == 0 {
		return prevTags
	}
	for _, newTag := range newTags {
		for _, prevTag := range prevTags {
			combinedTags = append(combinedTags, prevTag+","+newTag)
		}
	}
	return
}

func parseFileTag(filename string) (fileTag string) {
	filename = strings.TrimSuffix(filename, ".go")
	filename = strings.TrimSuffix(filename, "_test")
	var goarch, goos string
	for _, arch := range goarches {
		archSuffix := "_" + arch
		if strings.HasSuffix(filename, archSuffix) {
			goarch = arch
			filename = strings.TrimSuffix(filename, archSuffix)
			break
		}
	}
	for _, os := range gooses {
		osSuffix := "_" + os
		if strings.HasSuffix(filename, osSuffix) {
			goos = os
			filename = strings.TrimSuffix(filename, osSuffix)
			break
		}
	}
	if goos != "" {
		fileTag += goos
	}
	if goarch != "" {
		if fileTag != "" {
			fileTag += ","
		}
		fileTag += goarch
	}
	return fileTag
}

func parseTag(commentLine *ast.Comment) []string {
	comment := commentLine.Text
	if !strings.HasPrefix(comment, "// +build ") {
		return nil
	}
	tags := strings.TrimPrefix(comment, "// +build ")
	return strings.Split(tags, " ")
}

func valueNamesFrom(filepath, buildTags, parentName string, specs []ast.Spec) (names []gxui.TreeNode) {
	for _, spec := range specs {
		valSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, name := range valSpec.Names {
			if name.String() == "_" {
				// _ can be redeclared multiple times, so providing access to
				// it in the TOC isn't that useful.
				continue
			}
			var newName Name
			newName.name = name.String()
			if buildTags != "" {
				newName.name = fmt.Sprintf("%s (%s)", newName.name, buildTags)
			}
			newName.path = parentName + "." + newName.name
			newName.color = nameColor
			newName.filepath = filepath
			newName.pos = zeroBasedPos(name.Pos())
			names = append(names, newName)
		}
	}
	return
}
