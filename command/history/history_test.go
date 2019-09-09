// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package history_test

import (
	"testing"

	"github.com/nelsam/vidar/command/history"
	"github.com/nelsam/vidar/commander/input"
	"github.com/poy/onpar"
	"github.com/poy/onpar/expect"
)

func TestHistory(t *testing.T) {
	o := onpar.New()
	defer o.Run(t)

	// errEdit is the type of edit that we expect when we request history
	// that doesn't exist.
	var errEdit = input.Edit{At: -1}

	o.BeforeEach(func(t *testing.T) (expect.Expectation, *history.History, input.Edit) {
		all := history.Bindables(nil, nil, nil)
		hist := findHistory(t, all)
		ed := input.Edit{At: 300, Old: []rune("foo"), New: []rune("bacon")}
		hist.TextChanged(nil, ed)
		return expect.New(t), hist, ed
	})

	o.Spec("it knows how to rewind history", func(expect expect.Expectation, h *history.History, ed input.Edit) {
		expected := input.Edit{At: ed.At, Old: ed.New, New: ed.Old}
		expect(h.Rewind()).To(equal(expected))
	})

	o.Spec("it ignores edits returned from Rewind()", func(expect expect.Expectation, h *history.History, ed input.Edit) {
		h.TextChanged(nil, h.Rewind())
		expect(h.Rewind()).To(equal(errEdit))
	})

	o.Group("after rewinding history", func() {
		o.BeforeEach(func(expect expect.Expectation, h *history.History, ed input.Edit) (expect.Expectation, *history.History, input.Edit) {
			h.TextChanged(nil, h.Rewind())
			return expect, h, ed
		})

		o.Spec("it knows how to fast forward history", func(expect expect.Expectation, h *history.History, ed input.Edit) {
			expect(h.FastForward(0)).To(equal(ed))
		})

		o.Spec("it ignores edits returned from FastForward()", func(expect expect.Expectation, h *history.History, ed input.Edit) {
			h.TextChanged(nil, h.FastForward(0))
			expect(h.FastForward(0)).To(equal(errEdit))
		})

		o.Spec("it handles branching history", func(expect expect.Expectation, h *history.History, ed input.Edit) {
			expect(h.Branches()).To(equal(uint(1)))

			branch := input.Edit{At: 123, Old: []rune("eggs"), New: []rune("eggs")}
			h.TextChanged(nil, branch)
			h.TextChanged(nil, h.Rewind())
			expect(h.Branches()).To(equal(uint(2)))

			ff := h.FastForward(0)
			expect(ff).To(equal(ed))

			h.TextChanged(nil, ff)
			h.TextChanged(nil, h.Rewind())
			ff = h.FastForward(1)
			expect(ff).To(equal(branch))
		})
	})
}
