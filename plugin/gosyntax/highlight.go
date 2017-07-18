// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package gosyntax

import (
	"context"
	"log"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commands"
	"github.com/nelsam/vidar/syntax"
)

type TextEditor interface {
	Text() string
}

type Highlight struct {
	ctx    context.Context
	layers gxui.CodeSyntaxLayers
	syntax *syntax.Syntax
}

func New() *Highlight {
	return &Highlight{syntax: syntax.New(syntax.DefaultTheme)}
}

func (h Highlight) Name() string {
	return "go-syntax-highlight"
}

func (h Highlight) CommandName() string {
	return "input-handler"
}

func (h Highlight) TextChanged(ctx context.Context, editor commands.Editor, _ []commands.Edit) {
	// TODO: only update layers that changed.
	err := h.syntax.Parse(editor.(TextEditor).Text())
	if err != nil {
		// TODO: Report the error in the UI
		log.Printf("Error parsing syntax: %s", err)
	}
	layers := h.syntax.Layers()
	h.layers = make(gxui.CodeSyntaxLayers, 0, len(layers))
	for _, layer := range layers {
		h.layers = append(h.layers, layer)
	}
}

func (h Highlight) Apply(e commands.Editor) {
	e.SetSyntaxLayers(h.layers)
}
