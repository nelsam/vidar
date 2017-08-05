// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package history

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/input"
)

func New(theme gxui.Theme) (*History, *Undo, *Redo) {
	h := History{current: &node{}}
	u := Undo{history: &h}
	u.Theme = theme
	r := Redo{history: &h}
	r.Theme = theme
	return &h, &u, &r
}

type node struct {
	edit   input.Edit
	next   []*node
	parent *node
}

func (n *node) push(e input.Edit) *node {
	next := &node{edit: e, parent: n}
	n.next = append(n.next, next)
	return next
}

type History struct {
	current *node
	skip    []input.Edit
}

func (h *History) Name() string {
	return "history"
}

func (h *History) CommandName() string {
	return "input-handler"
}

func (h *History) Init(input.Editor, []rune) {}

func (h *History) shouldSkip(e input.Edit) bool {
	if len(h.skip) == 0 {
		return false
	}
	nextSkip := h.skip[0]
	return nextSkip.At == e.At &&
		string(nextSkip.Old) == string(e.Old) &&
		string(nextSkip.New) == string(e.New)
}

func (h *History) TextChanged(_ input.Editor, e input.Edit) {
	if h.shouldSkip(e) {
		copy(h.skip, h.skip[1:])
		h.skip = h.skip[:len(h.skip)-1]
		return
	}
	h.current = h.current.push(e)
}

// Rewind tells h to rewind its current state and return the
// input.Edit that needs to be applied in order to rewind the
// text to its previous state.
func (h *History) Rewind() input.Edit {
	if h.current.parent == nil {
		return input.Edit{At: -1}
	}
	e := h.current.edit
	h.current = h.current.parent
	undo := input.Edit{
		At:  e.At,
		Old: e.New,
		New: e.Old,
	}
	h.skip = append(h.skip, undo)
	return undo
}

// Branches returns the number of branches available to fast
// forward to at the current node.
func (h *History) Branches() uint {
	return uint(len(h.current.next))
}

// FastForward moves h's state forward in the history, based
// on branch.  To fast forward the most recent undo, run
// h.FastForward(h.Branches() - 1).
func (h *History) FastForward(branch uint) input.Edit {
	if branch > uint(len(h.current.next)) {
		return input.Edit{At: -1}
	}
	h.current = h.current.next[branch]
	h.skip = append(h.skip, h.current.edit)
	return h.current.edit
}

func (h *History) Apply(input.Editor) error { return nil }
