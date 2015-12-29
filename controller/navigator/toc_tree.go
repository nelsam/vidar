package navigator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
)

var (
	// Since the const values aren't exported by go/build, I've just copied them
	// from https://github.com/golang/go/blob/master/src/go/build/syslist.go
	gooses   = []string{"android", "darwin", "dragonfly", "freebsd", "linux", "nacl", "netbsd", "openbsd", "plan9", "solaris", "windows"}
	goarches = []string{"386", "amd64", "amd64p32", "arm", "armbe", "arm64", "arm64be", "ppc64", "ppc64le", "mips", "mipsle", "mips64", "mips64le", "mips64p32", "mips64p32le", "ppc", "s390", "s390x", "sparc", "sparc64"}
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
	pkgs, err := parser.ParseDir(token.NewFileSet(), t.path, nil, parser.ParseComments)
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

		fileTags := readTags(f)
		// TODO: use the fileTags to add sub-categories.

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

// popMatchingLastField only pops the last field in fields if it is found
// to match any item in haystack.
func popMatchingLastField(fields, haystack []string) (fieldsWithoutFoundField []string, foundField string) {
	if len(fields) == 0 {
		return fields, false
	}
	lastField := fields[len(fields)-1]
	for _, blade := range haystack {
		if blade == lastField {
			return fields[:len(fields)-1], lastField
		}
	}
	return fields, ""
}

type tagList struct {
	tags []interface{}
	op   string
}

// parseTags parses a build tag string, following these rules:
//
// 1. Space-separated tags imply "OR".
// 2. Comma-separated tags imply "AND".
// 3. Tags on separate lines will not be passed to this function.
func parseTags(tags string) tagList {
	orList := tagList{op: "OR"}
	for _, andTag := range strings.Split(tags, " ") {
		andList := tagList{op: "AND"}
		for _, tag := range strings.Split(andTag, ",") {
			andList.tags = append(andList.tags, tag)
		}
		orList.tags = append(orList.tags, andList)
	}
	return orList
}

func (t tagList) String() string {
	var tagRep string
	for _, tag := range tags {
		if len(tagRep) > 0 {
			tagRep = fmt.Sprintf("%s %s", tagRep, t.op)
		}
		switch src := tag.(type) {
		case string:
			tagRep = fmt.Sprintf("%s %s", tagRep, src)
		case tagList:
			subTagRep := src.String()
			if len(src.tags) > 1 {
				subTagRep = fmt.Sprintf("(%s)", subTagRep)
			}
			tagRep = fmt.Sprintf("%s %s", tagRep, subTagRep)
		}
	}
	return tagRep
}

// readTags performs ugly, ugly parsing of file name to figure out if it ends with
// goos and/or goarch values, then gathers all the tags together with those values
// and returns a representation of all applicable tags.
func readTags(f *ast.File) string {
	var (
		tags  = tagList{op: "AND"}
		found string
	)
	nameParts := strings.Split(strings.TrimSuffix(filename, ".go"), "_")
	nameParts, found = popMatchingLastField(nameParts, goarches)
	if len(found) > 0 {
		tags.tags = append(tags.tags, found)
	}
	nameParts, found = popMatchingLastField(nameParts, gooses)
	if len(found) > 0 {
		tags.tags = append(tags.tags, found)
	}

	for _, block := range f.Comments {
		for _, comment := range block.List {
			commentText = strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
			if strings.HasPrefix(commentText, "+build") {
				tagLine := strings.TrimSpace(strings.TrimPrefix(commentText, "+build"))
				newTags := parseTags(tagLine)
				if len(newTags.tags) > 0 {
					tags.tags = append(tags.tags, newTags)
				}
			}
		}
	}
	return tags.String()
}
