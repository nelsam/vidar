// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package history

import (
	"sync"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/input"
	"github.com/nelsam/vidar/plugin/command"
)

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

// Bindables returns the slice of bind.Bindable types that is implemented
// by this package.
func Bindables(_ command.Commander, _ gxui.Driver, theme *basic.Theme) []bind.Bindable {
	h := History{current: &node{}, all: make(map[string]*node)}
	onOpen := OnOpen{theme: theme}
	return []bind.Bindable{&h, &onOpen}
}

// History keeps track of change history for a file.
type History struct {
	current *node
	skip    []input.Edit
	mu      sync.Mutex

	all map[string]*node
}

// Name returns the name of h
func (h *History) Name() string {
	return "history"
}

// OpNames returns the name of bind.Op types that
// h needs to bind to.
func (h *History) OpNames() []string {
	return []string{"input-handler", "focus-location"}
}

// Init implements input.ChangeHook
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
	h.mu.Lock()
	defer h.mu.Unlock()
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
	h.mu.Lock()
	defer h.mu.Unlock()
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

// Apply is unused on history - recording the changes is all
// that is needed.
func (h *History) Apply(input.Editor) error { return nil }

// FileChanged updates the current history when the focused
// file is changed.
func (h *History) FileChanged(oldPath, newPath string) {
	if oldPath != "" {
		h.all[oldPath] = h.current
	}
	h.current = &node{}
	if n, ok := h.all[newPath]; ok {
		h.current = n
	}
	h.skip = nil
}
