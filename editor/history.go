package editor

import "github.com/nelsam/gxui"

const editCombineLen = 5

// EditTreeNode is a node within an edit history.  It
// can have any number of child nodes.  Current contains
// the currently in-use path.
type EditTreeNode struct {
	Edit       gxui.TextBoxEdit
	Parent     *EditTreeNode
	CurrentIdx int
	Children   []*EditTreeNode
}

func (n *EditTreeNode) SequentialWrite(edit gxui.TextBoxEdit) bool {
	return n.Edit.At+n.Edit.Delta == edit.At &&
		n.Edit.Delta < editCombineLen &&
		edit.Delta == 1 &&
		len(n.Edit.Old) == 0 &&
		len(edit.Old) == 0
}

func (n *EditTreeNode) SequentialDelete(edit gxui.TextBoxEdit) bool {
	return n.Edit.At+edit.Delta == edit.At &&
		-n.Edit.Delta < editCombineLen &&
		edit.Delta == -1 &&
		len(n.Edit.New) == 0 &&
		len(edit.New) == 0
}

func (n *EditTreeNode) Add(edit gxui.TextBoxEdit) *EditTreeNode {
	if n.SequentialWrite(edit) {
		// Combine multiple instances of new characters into a single event
		n.Edit.Delta++
		n.Edit.New = append(n.Edit.New, edit.New...)
		return n
	}
	if n.SequentialDelete(edit) {
		// Combine multiple instances of deleted characters into a single event
		n.Edit.At = edit.At
		n.Edit.Delta--
		n.Edit.Old = append(edit.Old, n.Edit.Old...)
		return n
	}
	new := &EditTreeNode{Edit: edit, Parent: n}
	n.Children = append(n.Children, new)
	n.CurrentIdx = len(n.Children) - 1
	return new
}

func (n *EditTreeNode) Length() int {
	length := 1
	for _, child := range n.Children {
		length += child.Length()
	}
	return length
}

type History struct {
	head    *EditTreeNode
	current *EditTreeNode
}

func NewHistory() *History {
	firstNode := &EditTreeNode{}
	return &History{
		head:    firstNode,
		current: firstNode,
	}
}

func (h *History) Add(edits ...gxui.TextBoxEdit) {
	for _, edit := range edits {
		h.current = h.current.Add(edit)
	}
}

func (h *History) Shrink(size int) {
	for h.head.Length() > size {
		h.head = h.head.Children[h.head.CurrentIdx]
		h.head.Parent = nil
	}
}

func (h *History) Undo() gxui.TextBoxEdit {
	if h.current.Parent == nil {
		return gxui.TextBoxEdit{}
	}
	edit := h.current.Edit
	h.current = h.current.Parent
	return edit
}

func (h *History) RedoCurrent() gxui.TextBoxEdit {
	return h.Redo(uint(h.current.CurrentIdx))
}

func (h *History) Redo(index uint) gxui.TextBoxEdit {
	if index >= uint(len(h.current.Children)) {
		return gxui.TextBoxEdit{}
	}
	h.current = h.current.Children[index]
	return h.current.Edit
}
