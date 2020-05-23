// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package history

import (
	"sync/atomic"
	"unsafe"

	"github.com/nelsam/vidar/commander/text"
)

// node is an entry in a linked list.
type node struct {
	// nextP *node - the next entry in the list.
	nextP unsafe.Pointer

	// the above should be kept first in the struct for byte alignment.

	edit text.Edit
}

// next performs atomic incantations to load n.nextP, returning
// it as a *node.
func (n *node) next() *node {
	return (*node)(atomic.LoadPointer(&n.nextP))
}

// setNext performs atomic incantations to set n.nextP to nn
func (n *node) setNext(nn *node) {
	atomic.StorePointer(&n.nextP, unsafe.Pointer(nn))
}

// casNext performs atomic compare-and-swap operations to set
// n.nextP to nn, but only if it is currently equal to old.
// It returns whether or not n.nextP was equal to old.
func (n *node) casNext(old, nn *node) bool {
	return atomic.CompareAndSwapPointer(&n.nextP, unsafe.Pointer(old), unsafe.Pointer(nn))
}

// A branch is an entry in a (non-binary) tree.  The first child
// will be at next(); all other children will be at
// next().siblings().
type branch struct {
	// prevP *branch - the parent node
	prevP unsafe.Pointer

	// nextP *branch - the first child node.  Additional branching
	// nodes should be accessed via this node's siblings.
	nextP unsafe.Pointer

	// siblingsP *[]*branch - the siblings of this node.  Normal
	// (i.e. non-branching) edits perform much better if next()
	// directly returns a *node, but if we have a sibling() which
	// also returns a *node we end up hitting double digits of
	// milliseconds per thousand operations in heavily branching
	// history.  So we take the extra performance penalty of
	// using a slice *only* on branching edits, to prevent the
	// performance hit from being paid per child node.
	siblingsP unsafe.Pointer

	// the above should be kept first in the struct for byte alignment.

	edit text.Edit
}

// prev performs atomic incantations to load b.prevP and return
// it as a *branch.
func (b *branch) prev() *branch {
	return (*branch)(atomic.LoadPointer(&b.prevP))
}

// next performs atomic incantations to load the i'th child branch
// of b and return it as a *branch.
func (b *branch) next(i uint) *branch {
	next := (*branch)(atomic.LoadPointer(&b.nextP))
	if i == 0 || next == nil {
		return next
	}
	sibIdx := i - 1
	sibs := next.siblings()
	if sibIdx >= uint(len(sibs)) {
		return nil
	}
	return sibs[sibIdx]
}

// siblings performs atomic incantations to load b.siblingsP and
// return it as a []*branch.
func (b *branch) siblings() []*branch {
	sibs := (*[]*branch)(atomic.LoadPointer(&b.siblingsP))
	if sibs == nil {
		return nil
	}
	return *sibs
}

// push adds e to the next empty child branch of b.
func (b *branch) push(e text.Edit) *branch {
	next := &branch{edit: e, prevP: unsafe.Pointer(b)}
	np := unsafe.Pointer(next)
	done := atomic.CompareAndSwapPointer(&b.nextP, nil, np)
	if !done {
		fchild := b.next(0)

		// We have to do _some_ kind of special case for nil; might as
		// well just try it here.
		fsib := []*branch{next}
		done = atomic.CompareAndSwapPointer(&fchild.siblingsP, nil, unsafe.Pointer(&fsib))
		for !done {
			oldSibsP := (*[]*branch)(atomic.LoadPointer(&fchild.siblingsP))
			sibs := append(*oldSibsP, next)
			done = atomic.CompareAndSwapPointer(&fchild.siblingsP, unsafe.Pointer(oldSibsP), unsafe.Pointer(&sibs))
		}
	}
	return next
}

// tree is a simple tree implementation to atomically store a branching
// history of text.Edit values.
type tree struct {
	// trunkP *branch
	trunkP unsafe.Pointer

	// the above should be kept first in the struct for byte alignment.
}

// trunk performs atomic incantations to load t.trunkP and return it as
// a *branch.
func (t *tree) trunk() *branch {
	return (*branch)(atomic.LoadPointer(&t.trunkP))
}

// setTrunk performs atomic incantations to set t.trunkP to b.
func (t *tree) setTrunk(b *branch) {
	atomic.StorePointer(&t.trunkP, unsafe.Pointer(b))
}
