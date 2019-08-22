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

type ll struct {
	edit input.Edit
	next unsafe.Pointer
}

type node struct {
	// prevP *node
	prevP unsafe.Pointer
	// nextP *node
	nextP unsafe.Pointer
	// in order to keep access to multiple children goroutine-safe
	// without using something slow, we keep access to siblings too.
	// siblingP *node
	siblingP unsafe.Pointer
	// the above should be kept first in the struct for byte alignment.

	edit input.Edit
}

func (n *node) prev() *node {
	return (*node)(atomic.LoadPointer(&n.prevP))
}

func (n *node) next() *node {
	return (*node)(atomic.LoadPointer(&n.nextP))
}

func (n *node) sibling() *node {
	return (*node)(atomic.LoadPointer(&n.siblingP))
}

func (n *node) push(e input.Edit) *node {
	next := &node{edit: e, prevP: unsafe.Pointer(n)}
	np := unsafe.Pointer(next)
	done := atomic.CompareAndSwapPointer(&n.nextP, nil, np)
	for curr := n.next(); !done; curr = curr.sibling() {
		done = atomic.CompareAndSwapPointer(&curr.siblingP, nil, np)
	}
	return next
}

// Bindables returns the slice of bind.Bindable types that is implemented
// by this package.
func Bindables(_ command.Commander, _ gxui.Driver, theme *basic.Theme) []bind.Bindable {
	h := History{currentP: unsafe.Pointer(&node{}), all: make(map[string]*node)}
	onOpen := OnOpen{theme: theme}
	return []bind.Bindable{&h, &onOpen}
}

// History keeps track of change history for a file.
type History struct {
	// current *node
	currentP unsafe.Pointer
	// skip *node
	skipP unsafe.Pointer
	// the above must be kept first in the struct for byte alignment

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

func (h *History) current() *node {
	return (*node)(atomic.LoadPointer(&h.currentP))
}

func (h *History) skip() *node {
	return (*node)(atomic.LoadPointer(&h.skipP))
}

func (h *History) addSkip(e input.Edit) {
	p := unsafe.Pointer(&node{edit: e})
	added := atomic.CompareAndSwapPointer(&h.skipP, nil, p)
	for curr := h.skip(); !added; curr = curr.next() {
		added = atomic.CompareAndSwapPointer(&curr.nextP, nil, p)
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
// forward to at the current node.
func (h *History) Branches() uint {
	var count uint
	for curr := h.current().next(); curr != nil; curr = curr.sibling() {
		count++
	}
	return count
}

// FastForward moves h's state forward in the history, based
// on branch.  To fast forward the most recent undo, run
// h.FastForward(h.Branches() - 1).
func (h *History) FastForward(branch uint) input.Edit {
	if branch > h.Branches() {
		return input.Edit{At: -1}
	}
	ff := h.current().next()
	for ; branch > 0; branch-- {
		ff = ff.sibling()
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
	h.currentP = unsafe.Pointer(&node{})
	if n, ok := h.all[newPath]; ok {
		atomic.StorePointer(&h.currentP, unsafe.Pointer(n))
	}
	atomic.StorePointer(&h.skipP, nil)
}
