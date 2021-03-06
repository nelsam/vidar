// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

//go:generate hel

package vsyntax

import (
	"context"
	"sync"

	"github.com/nelsam/vidar/commander/text"
	"github.com/nelsam/vidar/syntax"
	"github.com/nelsam/vidar/theme"
)

// Editor is used to force mocks to be generated locally for a
// remote type.  We kind of _have_ to accept input.Editor in our
// local methods because we're implementing an interface.
type Editor interface {
	text.Editor
}

// Highlight implements a syntax highlighting plugin for the V language.
type Highlight struct {
	ctx    context.Context
	layers []text.SyntaxLayer
	parser syntax.Generic

	mu sync.Mutex
}

func New() *Highlight {
	return &Highlight{
		parser: syntax.Generic{
			Scopes: []syntax.Scope{
				{Open: "{", Close: "}"},
				{Open: "(", Close: ")"},
				{Open: "[", Close: "]"},
			},
			Wrapped: []syntax.Wrapped{
				{Open: "'", Close: "'", Escapes: []string{`\'`}, Construct: theme.String},
				{Open: `"`, Close: `"`, Escapes: []string{`\"`}, Construct: theme.String},
				{Open: "/*", Close: "*/", Nested: true, Construct: theme.Comment},
				{Open: "//", Close: "\n", Construct: theme.Comment},
			},
			StaticWords: map[string]theme.LanguageConstruct{
				"break":     theme.Keyword,
				"const":     theme.Keyword,
				"continue":  theme.Keyword,
				"defer":     theme.Keyword,
				"else":      theme.Keyword,
				"enum":      theme.Keyword,
				"fn":        theme.Keyword,
				"for":       theme.Keyword,
				"go":        theme.Keyword,
				"goto":      theme.Keyword,
				"if":        theme.Keyword,
				"import":    theme.Keyword,
				"in":        theme.Keyword,
				"interface": theme.Keyword,
				"match":     theme.Keyword,
				"module":    theme.Keyword,
				"mut":       theme.Keyword,
				"none":      theme.Keyword,
				"or":        theme.Keyword,
				"pub":       theme.Keyword,
				"return":    theme.Keyword,
				"struct":    theme.Keyword,
				"type":      theme.Keyword,
				"println":   theme.Builtin,
				"eprintln":  theme.Builtin,
			},
			DynamicWords: []syntax.Word{
				{Before: "fn", Construct: theme.Func},
				{Before: "type", Construct: theme.Type},
			},
		},
	}
}

func (h *Highlight) Name() string {
	return "go-syntax-highlight"
}

func (h *Highlight) OpName() string {
	return "input-handler"
}

func (h *Highlight) Applied(e text.Editor, edits []text.Edit) {
	layers := e.SyntaxLayers()
	for i, l := range layers {
		layers[i] = h.moveLayer(l, edits)
	}
	e.SetSyntaxLayers(layers)
}

func (h *Highlight) moveLayer(l text.SyntaxLayer, edits []text.Edit) text.SyntaxLayer {
	for i, s := range l.Spans {
		l.Spans[i] = h.moveSpan(s, edits)
	}
	return l
}

func (h *Highlight) moveSpan(s text.Span, edits []text.Edit) text.Span {
	for _, e := range edits {
		if e.At > s.End {
			return s
		}
		delta := len(e.New) - len(e.Old)
		if delta == 0 {
			continue
		}
		s.End += delta
		if s.End < e.At {
			s.End = e.At
		}
		if e.At > s.Start {
			continue
		}
		s.Start += delta
		if s.Start < e.At {
			s.Start = e.At
		}
	}
	return s
}

func (h *Highlight) Init(e text.Editor, text []rune) {
	h.TextChanged(context.Background(), e, nil)
}

func (h *Highlight) TextChanged(ctx context.Context, editor text.Editor, _ []text.Edit) {
	h.mu.Lock()
	defer h.mu.Unlock()
	// TODO: only update layers that changed.
	m := h.parser.Parse(editor.Runes())
	select {
	case <-ctx.Done():
		return
	default:
	}

	h.layers = m.Layers()
}

func (h *Highlight) Apply(e text.Editor) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	e.SetSyntaxLayers(h.layers)
	return nil
}
