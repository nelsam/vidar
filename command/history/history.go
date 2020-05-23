// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package history

import (
	"sync"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/text"
	"github.com/nelsam/vidar/plugin/command"
)

// Bindables returns the slice of bind.Bindable types that is implemented
// by this package.
func Bindables(_ command.Commander, _ gxui.Driver, theme *basic.Theme) []bind.Bindable {
	h := History{all: make(map[string]*branch)}
	h.resetCurrent("")
	onOpen := OnOpen{theme: theme}
	return []bind.Bindable{&h, &onOpen}
}

// History keeps track of change history for a file.
type History struct {
	current tree
	skip    node

	// all stores history for all files - it is used when the open
	// file is changed, to store the history for the previously open
	// file.  We're just using a mutex to synchronize access to it
	// because it will only be accessed when the open file is changed.
	all   map[string]*branch
	allMu sync.Mutex
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
func (h *History) Init(text.Editor, []rune) {}

// resetCurrent resets h.current.trunk to a previous history (if one
// exists for path) or a new empty branch.
func (h *History) resetCurrent(path string) {
	if n, ok := h.all[path]; ok {
		h.current.setTrunk(n)
		return
	}
	h.current.setTrunk(&branch{})
}

// addSkip takes an edit that has been returned by h (from either
// Rewind or FastForward) and adds it to h.skipP, so that h will
// ignore e when it is triggered by TextChanged.
func (h *History) addSkip(e text.Edit) {
	n := &node{edit: e}
	added := h.skip.casNext(nil, n)
	if !added {
		for curr := h.skip.next(); !added; curr = curr.next() {
			added = curr.casNext(nil, n)
		}
	}
}

// shouldSkip reports whether e is an edit that was created by h and
// should be skipped.
func (h *History) shouldSkip(e text.Edit) bool {
	skip := h.skip.next()
	if skip == nil {
		return false
	}
	return skip.edit.At == e.At &&
		string(skip.edit.Old) == string(e.Old) &&
		string(skip.edit.New) == string(e.New)
}

// TextChanged hooks into the input handler to trigger off of changes
// in the editor so that h can track the history of those changes.
func (h *History) TextChanged(_ text.Editor, e text.Edit) {
	if h.shouldSkip(e) {
		h.skip.setNext(h.skip.next().next())
		return
	}
	h.current.setTrunk(h.current.trunk().push(e))
}

// Rewind tells h to rewind its current state and return the
// text.Edit that needs to be applied in order to rewind the
// text to its previous state.
func (h *History) Rewind() text.Edit {
	curr := h.current.trunk()
	prev := curr.prev()
	if prev == nil {
		return text.Edit{At: -1}
	}
	e := curr.edit
	h.current.setTrunk(prev)
	undo := text.Edit{
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
	next := h.current.trunk().next(0)
	if next == nil {
		return 0
	}
	return uint(len(next.siblings()) + 1)
}

// FastForward moves h's state forward in the history, based
// on branch.  To fast forward the most recent undo, run
// h.FastForward(h.Branches() - 1).
func (h *History) FastForward(branch uint) text.Edit {
	ff := h.current.trunk().next(branch)
	if ff == nil {
		return text.Edit{At: -1}
	}
	h.current.setTrunk(ff)
	h.addSkip(ff.edit)
	return ff.edit
}

// Apply is unused on history - recording the changes is all
// that is needed.
func (h *History) Apply(text.Editor) error { return nil }

// FileChanged updates the current history when the focused
// file is changed.
func (h *History) FileChanged(oldPath, newPath string) {
	h.allMu.Lock()
	defer h.allMu.Unlock()
	if oldPath != "" {
		h.all[oldPath] = h.current.trunk()
	}
	h.resetCurrent(newPath)
	h.skip.setNext(nil)
}
