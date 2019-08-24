// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package history

import (
	"sync/atomic"
	"unsafe"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/input"
	"github.com/nelsam/vidar/plugin/command"
)

// node is an entry in a linked list
type node struct {
	// nextP *node - the next entry in the list
	nextP unsafe.Pointer

	// the above should be kept first in the struct for byte alignment

	edit input.Edit
}

func (n *node) next() *node {
	return (*node)(atomic.LoadPointer(&n.nextP))
}

func (n *node) setNext(nn *node) {
	atomic.StorePointer(&n.nextP, unsafe.Pointer(nn))
}

func (n *node) casNext(old, nn *node) bool {
	return atomic.CompareAndSwapPointer(&n.nextP, unsafe.Pointer(old), unsafe.Pointer(nn))
}

// A branch is an entry in a (non-binary) tree.  The first child
// will be at next(); all other children will be at
// next().siblings()
type branch struct {
	// prevP *branch - the parent node
	prevP unsafe.Pointer

	// nextP *branch - the first child node.  Additional branching
	// nodes should be accessed via this node's siblings.
	nextP unsafe.Pointer

	// siblingsP *[]*branch - the siblings of this node.  Normal
	// (i.e. non-branching) edits perform much better if next()
	// directly returns a *node, but if siblings() also returns a
	// *node we end up hitting double digits of milliseconds per
	// thousand operations in heavily branching history.  So we
	// take the extra performance penalty of using a slice *only*
	// on branching edits, to prevent the performance hit from
	// being paid for every child node.
	siblingsP unsafe.Pointer

	// the above should be kept first in the struct for byte alignment.

	edit input.Edit
}

func (b *branch) prev() *branch {
	return (*branch)(atomic.LoadPointer(&b.prevP))
}

func (b *branch) next(i uint) *branch {
	next := (*branch)(atomic.LoadPointer(&b.nextP))
	if i == 0 {
		return next
	}
	sibIdx := i - 1
	sibs := next.siblings()
	if sibIdx >= uint(len(sibs)) {
		return nil
	}
	return sibs[sibIdx]
}

func (b *branch) siblings() []*branch {
	sibs := (*[]*branch)(atomic.LoadPointer(&b.siblingsP))
	if sibs == nil {
		return nil
	}
	return *sibs
}

func (b *branch) push(e input.Edit) *branch {
	next := &branch{edit: e, prevP: unsafe.Pointer(b)}
	np := unsafe.Pointer(next)
	done := atomic.CompareAndSwapPointer(&b.nextP, nil, np)
	if !done {
		sibs := b.siblings()
		sibs = append(sibs, next)
		atomic.StorePointer(&b.siblingsP, unsafe.Pointer(&sibs))
	}
	return next
}

// Bindables returns the slice of bind.Bindable types that is implemented
// by this package.
func Bindables(_ command.Commander, _ gxui.Driver, theme *basic.Theme) []bind.Bindable {
	h := History{currentP: unsafe.Pointer(&branch{}), all: make(map[string]*branch)}
	onOpen := OnOpen{theme: theme}
	return []bind.Bindable{&h, &onOpen}
}

// History keeps track of change history for a file.
type History struct {
	// current *branch
	currentP unsafe.Pointer
	// skip *branch
	skipP unsafe.Pointer
	// the above must be kept first in the struct for byte alignment

	all map[string]*branch
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

func (h *History) current() *branch {
	return (*branch)(atomic.LoadPointer(&h.currentP))
}

func (h *History) skip() *node {
	return (*node)(atomic.LoadPointer(&h.skipP))
}

func (h *History) addSkip(e input.Edit) {
	n := &node{edit: e}
	p := unsafe.Pointer(n)
	added := atomic.CompareAndSwapPointer(&h.skipP, nil, p)
	if !added {
		for curr := h.skip(); !added; curr = curr.next() {
			added = curr.casNext(nil, n)
		}
	}
}

// shouldSkip reports whether e is an edit that was created by h and
// should be skipped
func (h *History) shouldSkip(e input.Edit) bool {
	skip := h.skip()
	if skip == nil {
		return false
	}
	return skip.edit.At == e.At &&
		string(skip.edit.Old) == string(e.Old) &&
		string(skip.edit.New) == string(e.New)
}

// TextChanged hooks into the input handler to trigger off of changes
// in the editor so that h can track the history of those changes.
func (h *History) TextChanged(_ input.Editor, e input.Edit) {
	if h.shouldSkip(e) {
		atomic.StorePointer(&h.skipP, unsafe.Pointer(h.skip().next()))
		return
	}
	atomic.StorePointer(&h.currentP, unsafe.Pointer(h.current().push(e)))
}

// Rewind tells h to rewind its current state and return the
// input.Edit that needs to be applied in order to rewind the
// text to its previous state.
func (h *History) Rewind() input.Edit {
	curr := h.current()
	prev := curr.prev()
	if prev == nil {
		return input.Edit{At: -1}
	}
	e := curr.edit
	atomic.StorePointer(&h.currentP, unsafe.Pointer(prev))
	undo := input.Edit{
		At:  e.At,
		Old: e.New,
		New: e.Old,
	}
	h.addSkip(undo)
	return undo
}

// Branches returns the number of branches available to fast
// forward to at the current branch.
func (h *History) Branches() uint {
	next := h.current().next(0)
	if next == nil {
		return 0
	}
	return uint(len(next.siblings()) + 1)
}

// FastForward moves h's state forward in the history, based
// on branch.  To fast forward the most recent undo, run
// h.FastForward(h.Branches() - 1).
func (h *History) FastForward(branch uint) input.Edit {
	ff := h.current().next(branch)
	if ff == nil {
		return input.Edit{At: -1}
	}
	atomic.StorePointer(&h.currentP, unsafe.Pointer(ff))
	h.addSkip(ff.edit)
	return ff.edit
}

// Apply is unused on history - recording the changes is all
// that is needed.
func (h *History) Apply(input.Editor) error { return nil }

// FileChanged updates the current history when the focused
// file is changed.
func (h *History) FileChanged(oldPath, newPath string) {
	if oldPath != "" {
		h.all[oldPath] = h.current()
	}
	h.currentP = unsafe.Pointer(&branch{})
	if n, ok := h.all[newPath]; ok {
		atomic.StorePointer(&h.currentP, unsafe.Pointer(n))
	}
	atomic.StorePointer(&h.skipP, nil)
}
