package editor

import "github.com/nelsam/gxui"

// EditTreeNode is a node within an edit history.  It
// can have any number of child nodes.  Current contains
// the currently in-use path.
type EditTreeNode struct {
	Edit     gxui.TextBoxEdit
	Parent   *EditTreeNode
	Current  *EditTreeNode
	Children []*EditTreeNode
}

func (n EditTreeNode) Length() int {
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
		new := &EditTreeNode{Edit: edit, Parent: h.current}
		h.current.Children = append(h.current.Children, new)
		h.current.Current = new
		h.current = new
	}
}

func (h *History) Shrink(size int) {
	for h.head.Length() > size {
		h.head = h.head.Current
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

func (h *History) Redos() uint {
	return uint(len(h.current.Children))
}

func (h *History) Redo(index uint) gxui.TextBoxEdit {
	if index >= h.Redos() {
		return gxui.TextBoxEdit{}
	}
	h.current = h.current.Children[index]
	return h.current.Edit
}
