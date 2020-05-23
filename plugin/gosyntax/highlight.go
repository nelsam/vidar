// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package gosyntax

import (
	"context"
	"sync"

	"github.com/nelsam/vidar/commander/text"
	"github.com/nelsam/vidar/syntax"
)

type Highlight struct {
	ctx    context.Context
	layers []text.SyntaxLayer
	syntax *syntax.Syntax

	mu sync.Mutex
}

func New() *Highlight {
	return &Highlight{syntax: syntax.New()}
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
	err := h.syntax.Parse(editor.Text())
	if err != nil {
		// TODO: Report the error in the UI
		_ = err
	}
	select {
	case <-ctx.Done():
		return
	default:
	}

	h.layers = h.syntax.Layers()
}

func (h *Highlight) Apply(e text.Editor) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	e.SetSyntaxLayers(h.layers)
	return nil
}
