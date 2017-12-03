// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package gosyntax

import (
	"github.com/nelsam/vidar/commander/input"
)

type Highlight struct {
	ctx    context.Context
	layers []input.SyntaxLayer

	state state
}

func New() *Highlight {
	return &Highlight{}
}

func (h *Highlight) Name() string {
	return "go-syntax-highlight"
}

func (h *Highlight) OpName() string {
	return "input-handler"
}

func (h *Highlight) Applied(e input.Editor, edits []input.Edit) {
	layers := e.SyntaxLayers()
	for i, l := range layers {
		layers[i] = h.moveLayer(l, edits)
	}
	e.SetSyntaxLayers(layers)
}

func (h *Highlight) moveLayer(l input.SyntaxLayer, edits []input.Edit) input.SyntaxLayer {
	for i, s := range l.Spans {
		l.Spans[i] = h.moveSpan(s, edits)
	}
	return l
}

func (h *Highlight) moveSpan(s input.Span, edits []input.Edit) input.Span {
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

func (h *Highlight) Init(e input.Editor, text []rune) {
	h.TextChanged(context.Background(), e, nil)
}

func (h *Highlight) TextChanged(ctx context.Context, editor input.Editor, _ []input.Edit) {
	// TODO: only update layers that changed.
	var (
		s []string
		w string
	)
	for _, r := range []rune(editor.Text()) {
		if end := len(s) - 1; end >= 0 {
			switch s[end] {
			case "{-":
				if r == '{' {
					w = "{"
					continue
				}
				if r == '-' {
					if w == "{"
						s = append(s, "{-")
						w = ""
						continue
					}
					w = "-"
					continue
				}
				if w == "-" && r == '}' {
					s = s[:end]
					w = ""
				}
			case `"`:
				if r == '"' {
					s = s[:end]
				}
			case `"""`:
				switch r {
				case '"':
					w = append(w, `"`)
				default:
					w = ""
				}
				if w == `"""` {
					s = s[:end]
				}
			case "--":
				if r == '\n' {
					s = s[:end]
				}
			}
			continue
		}
		switch r {
		case '{':
			w = "{"
		case '-':
			switch w {
			case "{":
				s = append(s, "{-")
			case "-":
				s = append(s, "--")
			default:
				w = "-"
			}
		case '"':
			switch w {
			case `"`:
				w = `""`
			case `""`:
				s = append(s, `"""`)
			default:
				w = `"`
			}
		default:
			if w == `"` {
				s = append(s, `"`)
			}
			w = ""
		}
	}
}

type state struct {
	stack []context
}

func (s *state) push(c context) {
	s.stack = append(s.stack, c)
}

func (s *state) pop() context {
	last := len(s.stack) - 1
	t := s.stack[last]
	s.stack = s.stack[:last]
	return t
}

type context int

const (
	none context = iota

	// For strings using single "
	strStart
	str

	// for strings using triple """
	mStrStart
	mStr
	mStrEnd

	// for -- comments
	comment

	// for {- -} comments
	mCommentStart
	mComment
	mCommentEnd
)
