// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package navigator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/command/focus"
	"github.com/nelsam/vidar/commander/bind"
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
	nonGoColor = gxui.Color{
		R: 0.9,
		G: 0.9,
		B: 0.9,
		A: 1,
	}
	errColor = gxui.Color{
		R: 0.9,
		G: 0.2,
		B: 0,
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
	gooses = []string{
		"android", "darwin", "dragonfly", "freebsd", "linux", "nacl", "netbsd", "openbsd", "plan9", "solaris", "windows",
	}
	goarches = []string{
		"386", "amd64", "amd64p32", "arm", "armbe", "arm64", "arm64be", "ppc64", "ppc64le", "mips", "mipsle", "mips64",
		"mips64le", "mips64p32", "mips64p32le", "ppc", "s390", "s390x", "sparc", "sparc64",
	}
)

type Commander interface {
	Bindable(name string) bind.Bindable
	Execute(bind.Bindable)
}

type Opener interface {
	For(...focus.Opt) bind.Bindable
}

type genericNode struct {
	mixins.LinearLayout

	driver gxui.Driver
	theme  gxui.Theme

	button   *treeButton
	children gxui.LinearLayout
}

func newGenericNode(driver gxui.Driver, theme gxui.Theme, name string, color gxui.Color) *genericNode {
	g := &genericNode{}
	g.Init(g, driver, theme, name, color)
	return g
}

func (n *genericNode) Init(outer mixins.LinearLayoutOuter, driver gxui.Driver, theme gxui.Theme, name string, color gxui.Color) {
	n.LinearLayout.Init(outer, theme)

	n.driver = driver
	n.theme = theme

	n.SetDirection(gxui.TopToBottom)

	n.button = newTreeButton(driver, theme.(*basic.Theme), name)
	n.button.Label().SetColor(color)
	n.LinearLayout.AddChild(n.button)

	n.children = theme.CreateLinearLayout()
	n.children.SetDirection(gxui.TopToBottom)
	n.children.SetMargin(math.Spacing{L: 10})
}

func (n *genericNode) AddChild(c gxui.Control) *gxui.Child {
	child := n.children.AddChild(c)
	if len(n.children.Children()) > 1 {
		return child
	}
	n.button.SetExpandable(true)
	n.button.OnClick(func(gxui.MouseEvent) {
		if n.children.Attached() {
			n.LinearLayout.RemoveChild(n.children)
			n.button.Collapse()
			return
		}
		n.LinearLayout.AddChild(n.children)
		n.button.Expand()
	})
	return child
}

func (n *genericNode) MissingChild() gxui.Control {
	if n.children.Attached() {
		return nil
	}
	return n.children
}

type packageNode struct {
	genericNode

	cmdr                       Commander
	consts, vars, types, funcs *genericNode
	typeMap                    map[string]*Name
}

func newPackageNode(cmdr Commander, driver gxui.Driver, theme gxui.Theme, name string) *packageNode {
	node := &packageNode{
		cmdr:    cmdr,
		consts:  newGenericNode(driver, theme, "constants", genericColor),
		vars:    newGenericNode(driver, theme, "vars", genericColor),
		types:   newGenericNode(driver, theme, "types", genericColor),
		funcs:   newGenericNode(driver, theme, "funcs", genericColor),
		typeMap: make(map[string]*Name),
	}
	node.Init(node, driver, theme, name, skippableColor)
	node.AddChild(node.consts)
	node.AddChild(node.vars)
	node.AddChild(node.types)
	node.AddChild(node.funcs)
	return node
}

func (p *packageNode) expand() {
	p.consts.button.Click(gxui.MouseEvent{})
	p.vars.button.Click(gxui.MouseEvent{})
	p.types.button.Click(gxui.MouseEvent{})
	p.funcs.button.Click(gxui.MouseEvent{})
}

func (p *packageNode) addConsts(consts ...*Name) {
	for _, c := range consts {
		p.consts.AddChild(c)
	}
}

func (p *packageNode) addVars(vars ...*Name) {
	for _, v := range vars {
		p.vars.AddChild(v)
	}
}

func (p *packageNode) addTypes(types ...*Name) {
	for _, typ := range types {
		// Since we can't guarantee that we parsed the type declaration before
		// any method declarations, we have to check to see if the type already
		// exists.
		existingType, ok := p.typeMap[typ.button.Text()]
		if !ok {
			p.typeMap[typ.button.Text()] = typ
			p.types.AddChild(typ)
			continue
		}
		existingType.button.Label().SetColor(typ.button.Label().Color())
		existingType.filepath = typ.filepath
		existingType.position = typ.position
	}
}

func (p *packageNode) addFuncs(funcs ...*Name) {
	for _, f := range funcs {
		p.funcs.AddChild(f)
	}
}

func (p *packageNode) addMethod(typeName string, method *Name) {
	typ, ok := p.typeMap[typeName]
	if !ok {
		typ = newName(p.cmdr, p.driver, p.theme, typeName, nameColor)
		p.typeMap[typeName] = typ
		p.types.AddChild(typ)
	}
	typ.AddChild(method)
}

type Location struct {
	filepath string
	position token.Position
}

func (l Location) File() string {
	return l.filepath
}

func (l Location) Position() token.Position {
	return l.position
}

type Name struct {
	genericNode
	Location

	cmdr Commander
}

func newName(cmdr Commander, driver gxui.Driver, theme gxui.Theme, name string, color gxui.Color) *Name {
	node := &Name{cmdr: cmdr}
	node.Init(node, driver, theme, name, color)
	node.button.OnClick(func(gxui.MouseEvent) {
		cmd := node.cmdr.Bindable("focus-location").(Opener)
		node.cmdr.Execute(cmd.For(focus.Path(node.File()), focus.Offset(node.Position().Offset)))
	})
	return node
}

type TOC struct {
	mixins.LinearLayout

	cmdr   Commander
	driver gxui.Driver
	theme  gxui.Theme

	dir        string
	fileSet    *token.FileSet
	packageMap map[string]*packageNode

	lock sync.Mutex
}

func NewTOC(cmdr Commander, driver gxui.Driver, theme gxui.Theme, dir string) *TOC {
	toc := &TOC{
		cmdr:   cmdr,
		driver: driver,
		theme:  theme,
		dir:    dir,
	}
	toc.Init(toc, theme)
	toc.Reload()
	return toc
}

func (t *TOC) Reload() {
	t.lock.Lock()
	defer t.lock.Unlock()
	defer t.expandPackages()
	t.fileSet = token.NewFileSet()
	t.RemoveAll()
	t.packageMap = make(map[string]*packageNode)
	allFiles, err := ioutil.ReadDir(t.dir)
	if err != nil {
		log.Printf("Received error reading directory %s: %s", t.dir, err)
		return
	}
	t.parseFiles(t.dir, allFiles...)
}

func (t *TOC) expandPackages() {
	for _, c := range t.Children() {
		pkg, ok := c.Control.(*packageNode)
		if !ok {
			continue
		}
		pkg.button.Click(gxui.MouseEvent{})
		pkg.expand()
	}
}

func (t *TOC) parseFiles(dir string, files ...os.FileInfo) {
	filesNode := newGenericNode(t.driver, t.theme, "files", skippableColor)
	t.AddChild(filesNode)
	defer filesNode.button.Click(gxui.MouseEvent{})
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileNode := t.parseFile(dir, file)
		fileNode.filepath = filepath.Join(dir, file.Name())
		filesNode.AddChild(fileNode)
	}
}

func (t *TOC) parseFile(dir string, file os.FileInfo) *Name {
	if !strings.HasSuffix(file.Name(), ".go") {
		return newName(t.cmdr, t.driver, t.theme, file.Name(), nonGoColor)
	}
	path := filepath.Join(dir, file.Name())
	f, err := parser.ParseFile(t.fileSet, path, nil, parser.ParseComments)
	if err != nil {
		return newName(t.cmdr, t.driver, t.theme, file.Name(), errColor)
	}
	t.parseAstFile(path, f)
	return newName(t.cmdr, t.driver, t.theme, file.Name(), nameColor)
}

func (t *TOC) parseAstFile(filepath string, file *ast.File) *packageNode {
	buildTags := findBuildTags(filepath, file)
	buildTagLine := strings.Join(buildTags, " ")

	packageName := file.Name.String()
	pkgNode, ok := t.packageMap[packageName]
	if !ok {
		pkgNode = newPackageNode(t.cmdr, t.driver, t.theme, packageName)
		t.packageMap[pkgNode.button.Text()] = pkgNode
		t.AddChild(pkgNode)
	}
	for _, decl := range file.Decls {
		switch src := decl.(type) {
		case *ast.GenDecl:
			t.parseGenDecl(pkgNode, src, filepath, buildTagLine)
		case *ast.FuncDecl:
			if src.Name.String() == "init" {
				// There can be multiple inits in the package, so this
				// doesn't really help us in the TOC.
				continue
			}
			text := src.Name.String()
			if buildTagLine != "" {
				text = fmt.Sprintf("%s (%s)", text, buildTagLine)
			}
			name := newName(t.cmdr, t.driver, t.theme, text, nameColor)
			name.filepath = filepath
			name.position = t.fileSet.Position(src.Pos())
			if src.Recv == nil {
				pkgNode.addFuncs(name)
				continue
			}
			recvTyp := src.Recv.List[0].Type
			if starExpr, ok := recvTyp.(*ast.StarExpr); ok {
				recvTyp = starExpr.X
			}
			recvTypeName := recvTyp.(*ast.Ident).String()
			pkgNode.addMethod(recvTypeName, name)
		}
	}
	return pkgNode
}

func (t *TOC) parseGenDecl(pkgNode *packageNode, decl *ast.GenDecl, filepath, buildTags string) {
	switch decl.Tok.String() {
	case "const":
		pkgNode.addConsts(t.valueNamesFrom(filepath, buildTags, decl.Specs)...)
	case "var":
		pkgNode.addVars(t.valueNamesFrom(filepath, buildTags, decl.Specs)...)
	case "type":
		// I have yet to see a case where a type declaration has len(Specs) != 0.
		typeSpec := decl.Specs[0].(*ast.TypeSpec)
		name := typeSpec.Name.String()
		if buildTags != "" {
			name = fmt.Sprintf("%s (%s)", name, buildTags)
		}
		typ := newName(t.cmdr, t.driver, t.theme, name, nameColor)
		typ.filepath = filepath
		typ.position = t.fileSet.Position(typeSpec.Pos())
		pkgNode.addTypes(typ)
	}
}

func (t *TOC) valueNamesFrom(filepath, buildTags string, specs []ast.Spec) (names []*Name) {
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
			n := name.String()
			if buildTags != "" {
				n = fmt.Sprintf("%s (%s)", n, buildTags)
			}
			node := newName(t.cmdr, t.driver, t.theme, n, nameColor)
			node.filepath = filepath
			node.position = t.fileSet.Position(name.Pos())
			names = append(names, node)
		}
	}
	return names
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
