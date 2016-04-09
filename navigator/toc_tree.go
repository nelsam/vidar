package navigator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/go-fsnotify/fsnotify"
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

type packageNode struct {
	genericTreeNode

	consts, vars, types, funcs *genericTreeNode
	typeMap                    map[string]*Name
}

func newPackageNode(name string) *packageNode {
	node := &packageNode{}
	node.name = name
	node.consts = &genericTreeNode{name: "constants", path: name + ".constants", color: genericColor}
	node.vars = &genericTreeNode{name: "global vars", path: name + ".global vars", color: genericColor}
	node.types = &genericTreeNode{name: "types", path: name + ".types", color: genericColor}
	node.funcs = &genericTreeNode{name: "funcs", path: name + ".funcs", color: genericColor}
	node.typeMap = make(map[string]*Name)
	node.children = []gxui.TreeNode{node.consts, node.vars, node.types, node.funcs}
	return node
}

func (p *packageNode) addConsts(consts ...Name) {
	for _, c := range consts {
		c.path = p.consts.path + "." + c.name
		p.consts.children = append(p.consts.children, c)
	}
}

func (p *packageNode) addVars(vars ...Name) {
	for _, v := range vars {
		v.path = p.vars.path + "." + v.name
		p.vars.children = append(p.vars.children, v)
	}
}

func (p *packageNode) addTypes(types ...Name) {
	for _, typ := range types {
		// Since we can't guarantee that we parsed the type declaration before
		// any method declarations, we have to check to see if the type already
		// exists.
		existingType, ok := p.typeMap[typ.name]
		if !ok {
			typ.path = p.types.path + "." + typ.name
			p.typeMap[typ.name] = &typ
			p.types.children = append(p.types.children, &typ)
			continue
		}
		existingType.color = typ.color
		existingType.filepath = typ.filepath
		existingType.position = typ.position
	}
}

func (p *packageNode) addFuncs(funcs ...Name) {
	for _, f := range funcs {
		f.path = p.funcs.path + "." + f.name
		p.funcs.children = append(p.funcs.children, f)
	}
}

func (p *packageNode) addMethod(typeName string, method Name) {
	typ, ok := p.typeMap[typeName]
	if !ok {
		typ = &Name{}
		typ.name = typeName
		typ.path = p.types.path + "." + typ.name
		p.typeMap[typeName] = typ
		p.types.children = append(p.types.children, typ)
	}
	method.path = typ.path + "." + method.name
	typ.children = append(typ.children, method)
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
	genericTreeNode
	Location
}

type TOC struct {
	skippingTreeNode

	path       string
	fileSet    *token.FileSet
	packageMap map[string]*packageNode
	watcher    *fsnotify.Watcher

	lock sync.Mutex
}

func NewTOC(path string) *TOC {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("TOC: Could not create watcher")
	}
	toc := &TOC{
		path:    path,
		watcher: watcher,
	}
	toc.path = path
	if watcher != nil {
		toc.watch()
	}
	toc.Reload()
	return toc
}

func (t *TOC) watch() {
	allFiles, err := ioutil.ReadDir(t.path)
	if err != nil {
		log.Printf("Received error reading directory %s for watching: %s", t.path, err)
		return
	}
	t.watcher.Add(t.path)
	for _, f := range allFiles {
		if f.IsDir() {
			continue
		}
		fPath := path.Join(t.path, f.Name())
		if err := t.watcher.Add(fPath); err != nil {
			log.Printf("Could not watch file %s: %s", fPath, err)
		}
	}
	go t.watchEvents()
	go t.watchErrors()
}

func (t *TOC) watchEvents() {
	for event := range t.watcher.Events {
		switch event.Op {
		case fsnotify.Remove:
			t.watcher.Remove(event.Name)
		case fsnotify.Create:
			t.watcher.Add(event.Name)
		}
		t.Reload()
	}
}

func (t *TOC) watchErrors() {
	for err := range t.watcher.Errors {
		log.Printf("Received error %s from watcher", err)
	}
}

func (t *TOC) Reload() {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.fileSet = token.NewFileSet()
	t.children = make([]gxui.TreeNode, 0, len(t.children))
	t.packageMap = make(map[string]*packageNode)
	allFiles, err := ioutil.ReadDir(t.path)
	if err != nil {
		log.Printf("Received error reading directory %s: %s", t.path, err)
		return
	}
	t.parseFiles(t.path, allFiles...)
}

func (t *TOC) parseFiles(dir string, files ...os.FileInfo) {
	filesNode := &genericTreeNode{
		name:  "files",
		path:  "files",
		color: skippableColor,
	}
	t.children = append(t.children, filesNode)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileNode := t.parseFile(dir, file)
		filesNode.children = append(filesNode.children, fileNode)
	}
}

func (t *TOC) parseFile(dir string, file os.FileInfo) (fileNode Name) {
	fileNode.name = file.Name()
	fileNode.filepath = path.Join(dir, fileNode.name)
	fileNode.path = "files." + fileNode.name
	if !strings.HasSuffix(file.Name(), ".go") {
		fileNode.color = nonGoColor
		return fileNode
	}
	f, err := parser.ParseFile(t.fileSet, fileNode.filepath, nil, parser.ParseComments)
	if err != nil {
		fileNode.color = errColor
		return fileNode
	}
	fileNode.color = nameColor
	t.parseAstFile(fileNode.filepath, f)
	return fileNode
}

func (t *TOC) parseGenDecl(pkgNode *packageNode, decl *ast.GenDecl, filepath, buildTags string) {
	switch decl.Tok.String() {
	case "const":
		pkgNode.addConsts(t.valueNamesFrom(filepath, buildTags, decl.Specs)...)
	case "var":
		pkgNode.addVars(t.valueNamesFrom(filepath, buildTags, decl.Specs)...)
	case "type":
		// I have yet to see a case where a type declaration has more than one Specs.
		typeSpec := decl.Specs[0].(*ast.TypeSpec)
		typeName := typeSpec.Name.String()

		typ := Name{}
		typ.name = typeName
		if buildTags != "" {
			typ.name = fmt.Sprintf("%s (%s)", typ.name, buildTags)
		}
		typ.color = nameColor
		typ.filepath = filepath
		typ.position = t.fileSet.Position(typeSpec.Pos())
		pkgNode.addTypes(typ)
	}
}

func (t *TOC) parseAstFile(filepath string, file *ast.File) *packageNode {
	buildTags := findBuildTags(filepath, file)
	buildTagLine := strings.Join(buildTags, " ")

	packageName := file.Name.String()
	pkgNode, ok := t.packageMap[packageName]
	if !ok {
		pkgNode = newPackageNode(packageName)
		pkgNode.path = pkgNode.name
		pkgNode.color = skippableColor
		t.packageMap[pkgNode.name] = pkgNode
		t.children = append(t.children, pkgNode)
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
			var name Name
			name.name = src.Name.String()
			if buildTagLine != "" {
				name.name = fmt.Sprintf("%s (%s)", name.name, buildTagLine)
			}
			name.color = nameColor
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

func (t *TOC) valueNamesFrom(filepath, buildTags string, specs []ast.Spec) (names []Name) {
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
			newName.color = nameColor
			newName.filepath = filepath
			newName.position = t.fileSet.Position(name.Pos())
			names = append(names, newName)
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
