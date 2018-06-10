// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax

import (
	"go/ast"
	"go/parser"
	"go/token"
	"unicode/utf8"

	"github.com/nelsam/vidar/input"
	"github.com/nelsam/vidar/theme"
)

// Syntax is a type that reads Go source code to provide information
// on it.
type Syntax struct {
	scope       theme.LanguageConstruct
	fileSet     *token.FileSet
	layers      map[theme.LanguageConstruct]*input.SyntaxLayer
	runeOffsets []int
}

// New constructs a new *Syntax value with theme as its Theme field.
func New() *Syntax {
	return &Syntax{}
}

// Parse parses the passed in Go source code, replacing s's stored
// context with that of the parsed source.  It returns any error
// encountered while parsing source, but will still store as much
// information as possible.
func (s *Syntax) Parse(source string) error {
	s.runeOffsets = make([]int, len(source))
	byteOffset := 0
	for runeIdx, r := range []rune(source) {
		byteIdx := runeIdx + byteOffset
		bytes := utf8.RuneLen(r)
		for i := byteIdx; i < byteIdx+bytes; i++ {
			s.runeOffsets[i] = -byteOffset
		}
		byteOffset += bytes - 1
	}

	s.fileSet = token.NewFileSet()
	s.scope = theme.ScopePair
	s.layers = make(map[theme.LanguageConstruct]*input.SyntaxLayer)
	f, err := parser.ParseFile(s.fileSet, "", source, parser.ParseComments)

	// Parse everything we can before returning the error.
	if f.Package.IsValid() {
		s.add(theme.Keyword, f.Package, len("package"))
	}
	for _, importSpec := range f.Imports {
		s.addNode(theme.String, importSpec)
	}
	for _, comment := range f.Comments {
		s.addNode(theme.Comment, comment)
	}
	for _, decl := range f.Decls {
		s.addDecl(decl)
	}
	for _, unresolved := range f.Unresolved {
		s.addUnresolved(unresolved)
	}
	return err
}

// Layers returns a gxui.CodeSyntaxLayer for each construct used from
// s.Theme when s.Parse was called.  The corresponding
// gxui.CodeSyntaxLayer will have its foreground and background
// constructs set, and all positions that should be highlighted that
// construct will be stored.
func (s *Syntax) Layers() []input.SyntaxLayer {
	l := make([]input.SyntaxLayer, 0, len(s.layers))
	for _, layer := range s.layers {
		l = append(l, *layer)
	}
	return l
}

func (s *Syntax) rainbowScope(openStart token.Pos, openLen int, closeStart token.Pos, closeLen int) (unscope func()) {
	s.add(s.scope, openStart, openLen)
	s.add(s.scope, closeStart, closeLen)
	s.scope++
	return func() { s.scope-- }
}

func (s *Syntax) add(construct theme.LanguageConstruct, pos token.Pos, byteLength int) {
	if byteLength == 0 {
		return
	}
	layer, ok := s.layers[construct]
	if !ok {
		layer = &input.SyntaxLayer{
			Construct: construct,
		}
		s.layers[construct] = layer
	}
	bytePos := s.fileSet.Position(pos).Offset
	if bytePos >= len(s.runeOffsets) {
		return
	}
	idx := s.runePos(bytePos)
	end := s.runePos(bytePos + byteLength)
	layer.Spans = append(layer.Spans, input.Span{Start: idx, End: end})
}

func (s *Syntax) runePos(bytePos int) int {
	if bytePos >= len(s.runeOffsets) {
		return -1
	}
	return bytePos + s.runeOffsets[bytePos]
}

func (s *Syntax) addNode(construct theme.LanguageConstruct, node ast.Node) {
	s.add(construct, node.Pos(), int(node.End()-node.Pos()))
}
