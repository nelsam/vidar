// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package focus_test

import (
	"testing"

	"github.com/apoydence/onpar"
	"github.com/apoydence/onpar/expect"
	"github.com/apoydence/onpar/matchers"
	"github.com/nelsam/vidar/bind"
	"github.com/nelsam/vidar/command/focus"
)

func TestLocation(t *testing.T) {
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) (expect.Expectation, *focus.Location) {
		loc := focus.NewLocation(nil)
		expect := expect.New(t)
		return expect, loc
	})

	o.Group("Store", func() {
		o.BeforeEach(func(e expect.Expectation, f *focus.Location) (expect.Expectation, *focus.Location, *mockMover, *mockEditorOpener, *mockOpener, *mockBinder) {
			return e, f, newMockMover(), newMockEditorOpener(), newMockOpener(), newMockBinder()
		})

		o.Spec("It returns executing when passed the necessary elements", func(expect expect.Expectation, l *focus.Location, m *mockMover, eo *mockEditorOpener, o *mockOpener, b *mockBinder) {
			expect(l.Store(m)).To(matchers.Equal(bind.Waiting))
			expect(l.Store(eo)).To(matchers.Equal(bind.Waiting))
			expect(l.Store(o)).To(matchers.Equal(bind.Waiting))
			expect(l.Store(b)).To(matchers.Equal(bind.Executing))
		})

		o.Spec("It doesn't require an opener to execute", func(expect expect.Expectation, l *focus.Location, m *mockMover, eo *mockEditorOpener, o *mockOpener, b *mockBinder) {
			expect(l.Store(m)).To(matchers.Equal(bind.Waiting))
			expect(l.Store(eo)).To(matchers.Equal(bind.Waiting))
			expect(l.Store(b)).To(matchers.Equal(bind.Executing))
		})

		o.Spec("It doesn't require a mover to execute", func(expect expect.Expectation, l *focus.Location, m *mockMover, eo *mockEditorOpener, o *mockOpener, b *mockBinder) {
			expect(l.Store(eo)).To(matchers.Equal(bind.Waiting))
			expect(l.Store(b)).To(matchers.Equal(bind.Executing))
		})

		o.Spec("It does require an editoropener to execute", func(expect expect.Expectation, l *focus.Location, m *mockMover, eo *mockEditorOpener, o *mockOpener, b *mockBinder) {
			expect(l.Store(b)).To(matchers.Equal(bind.Waiting))
		})

		o.Spec("It does require a binder to execute", func(expect expect.Expectation, l *focus.Location, m *mockMover, eo *mockEditorOpener, o *mockOpener, b *mockBinder) {
			expect(l.Store(eo)).To(matchers.Equal(bind.Waiting))
		})
	})

	o.Group("Line", func() {
		o.BeforeEach(func(e expect.Expectation, l *focus.Location) (expect.Expectation, *focus.Location, *mockMover, *mockEditorOpener, *mockBinder) {
			l = l.For(focus.Line(3)).(*focus.Location)
			return e, l, newMockMover(), newMockEditorOpener(), newMockBinder()
		})

		o.Spec("It requires a mover to execute", func(expect expect.Expectation, l *focus.Location, m *mockMover, eo *mockEditorOpener, b *mockBinder) {
			expect(l.Store(eo)).To(matchers.Equal(bind.Waiting))
			expect(l.Store(b)).To(matchers.Equal(bind.Waiting))
			expect(l.Store(m)).To(matchers.Equal(bind.Executing))
		})
	})

	o.Group("Path", func() {
		o.BeforeEach(func(e expect.Expectation, l *focus.Location) (expect.Expectation, *focus.Location, *mockMover, *mockEditorOpener, *mockBinder) {
			l = l.For(focus.Path("foo")).(*focus.Location)
			return e, l, newMockMover(), newMockEditorOpener(), newMockBinder()
		})

		o.Spec("It doesn't require a mover to execute", func(expect expect.Expectation, l *focus.Location, m *mockMover, eo *mockEditorOpener, b *mockBinder) {
			expect(l.Store(eo)).To(matchers.Equal(bind.Waiting))
			expect(l.Store(b)).To(matchers.Equal(bind.Executing))
		})
	})

	o.Group("Line", func() {
		o.BeforeEach(func(e expect.Expectation, l *focus.Location) (expect.Expectation, *focus.Location, *mockMover, *mockEditorOpener, *mockBinder) {
			l = l.For(focus.Line(3)).(*focus.Location)
			return e, l, newMockMover(), newMockEditorOpener(), newMockBinder()
		})

		o.Spec("It requires a mover to execute", func(expect expect.Expectation, l *focus.Location, m *mockMover, eo *mockEditorOpener, b *mockBinder) {
			expect(l.Store(eo)).To(matchers.Equal(bind.Waiting))
			expect(l.Store(b)).To(matchers.Equal(bind.Waiting))
			expect(l.Store(m)).To(matchers.Equal(bind.Executing))
		})

		o.Spec("It doesn't warn if a column is set", func(expect expect.Expectation, l *focus.Location, m *mockMover, eo *mockEditorOpener, b *mockBinder) {
			l = l.For(focus.Column(20)).(*focus.Location)
			expect(l.Warn).To(matchers.Equal(""))
		})

		o.Spec("It warns if an offset is set", func(expect expect.Expectation, l *focus.Location, m *mockMover, eo *mockEditorOpener, b *mockBinder) {
			l = l.For(focus.Offset(20)).(*focus.Location)
			expect(l.Warn).To(matchers.Not(matchers.Equal("")))
		})
	})

	o.Group("Column", func() {
		o.BeforeEach(func(e expect.Expectation, l *focus.Location) (expect.Expectation, *focus.Location, *mockMover, *mockEditorOpener, *mockBinder) {
			l = l.For(focus.Column(3)).(*focus.Location)
			return e, l, newMockMover(), newMockEditorOpener(), newMockBinder()
		})

		o.Spec("It requires a mover to execute", func(expect expect.Expectation, l *focus.Location, m *mockMover, eo *mockEditorOpener, b *mockBinder) {
			expect(l.Store(eo)).To(matchers.Equal(bind.Waiting))
			expect(l.Store(b)).To(matchers.Equal(bind.Waiting))
			expect(l.Store(m)).To(matchers.Equal(bind.Executing))
		})

		o.Spec("It doesn't warn if a line is set", func(expect expect.Expectation, l *focus.Location, m *mockMover, eo *mockEditorOpener, b *mockBinder) {
			l = l.For(focus.Line(3)).(*focus.Location)
			expect(l.Warn).To(matchers.Equal(""))
		})

		o.Spec("It warns if an offset is set", func(expect expect.Expectation, l *focus.Location, m *mockMover, eo *mockEditorOpener, b *mockBinder) {
			l = l.For(focus.Offset(20)).(*focus.Location)
			expect(l.Warn).To(matchers.Not(matchers.Equal("")))
		})
	})

	o.Group("Offset", func() {
		o.BeforeEach(func(e expect.Expectation, l *focus.Location) (expect.Expectation, *focus.Location, *mockMover, *mockEditorOpener, *mockBinder) {
			l = l.For(focus.Offset(210)).(*focus.Location)
			return e, l, newMockMover(), newMockEditorOpener(), newMockBinder()
		})

		o.Spec("It requires a mover to execute", func(expect expect.Expectation, l *focus.Location, m *mockMover, eo *mockEditorOpener, b *mockBinder) {
			expect(l.Store(eo)).To(matchers.Equal(bind.Waiting))
			expect(l.Store(b)).To(matchers.Equal(bind.Waiting))
			expect(l.Store(m)).To(matchers.Equal(bind.Executing))
		})

		o.Spec("It warns if a line is set", func(expect expect.Expectation, l *focus.Location, m *mockMover, eo *mockEditorOpener, b *mockBinder) {
			l = l.For(focus.Line(3)).(*focus.Location)
			expect(l.Warn).To(matchers.Not(matchers.Equal("")))
		})

		o.Spec("It warns if a column is set", func(expect expect.Expectation, l *focus.Location, m *mockMover, eo *mockEditorOpener, b *mockBinder) {
			l = l.For(focus.Column(20)).(*focus.Location)
			expect(l.Warn).To(matchers.Not(matchers.Equal("")))
		})
	})
}
