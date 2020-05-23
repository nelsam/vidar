// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package vsyntax_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/nelsam/hel/pers"
	"github.com/nelsam/vidar/command/input"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/text"
	"github.com/nelsam/vidar/plugin/vsyntax"
	"github.com/nelsam/vidar/test"
	"github.com/nelsam/vidar/theme"
	"github.com/poy/onpar"
	"github.com/poy/onpar/expect"
)

var (
	// implementation checks
	highlight *vsyntax.Highlight
	_         bind.Bindable           = highlight
	_         bind.OpHook             = highlight
	_         input.AppliedChangeHook = highlight
	_         input.ContextChangeHook = highlight
)

func TestVSyntax(t *testing.T) {
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) (expectation, *vsyntax.Highlight) {
		return expect.New(t, expect.WithDiffer(test.Differ())), vsyntax.New()
	})

	o.Spec("it binds to the input-handler", func(expect expectation, h *vsyntax.Highlight) {
		ih := &input.Handler{}
		expect(h.OpName()).To(equal(ih.Name()))
	})

	for _, kw := range []string{
		"break",
		"const",
		"continue",
		"defer",
		"else",
		"enum",
		"fn",
		"for",
		"go",
		"goto",
		"if",
		"import",
		"in",
		"interface",
		"match",
		"module",
		"mut",
		"none",
		"or",
		"pub",
		"return",
		"struct",
		"type",
	} {
		kw := kw
		o.Spec(fmt.Sprintf("it highlights the %s keyword", kw), func(expect expectation, h *vsyntax.Highlight) {
			e := newMockEditor()
			pers.Return(e.RunesOutput, []rune(kw))
			h.TextChanged(context.Background(), e, nil)
			h.Apply(e)
			expect(e).To(haveMethodExecuted("SetSyntaxLayers", withArgs(
				contain(text.SyntaxLayer{
					Construct: theme.Keyword,
					Spans:     []text.Span{{Start: 0, End: len(kw)}},
				}),
			)))
		})
	}

	for _, builtin := range []string{
		"println",
		"eprintln",
	} {
		builtin := builtin
		o.Spec(fmt.Sprintf("it highlights the %s builtin", builtin), func(expect expectation, h *vsyntax.Highlight) {
			e := newMockEditor()
			pers.Return(e.RunesOutput, []rune(builtin))
			h.TextChanged(context.Background(), e, nil)
			h.Apply(e)
			expect(e).To(haveMethodExecuted("SetSyntaxLayers", withArgs(
				contain(text.SyntaxLayer{
					Construct: theme.Builtin,
					Spans:     []text.Span{{Start: 0, End: len(builtin)}},
				}),
			)))
		})
	}

	o.Spec("it highlights comments", func(expect expectation, h *vsyntax.Highlight) {
		comment := "// this is a comment"
		e := newMockEditor()
		pers.Return(e.RunesOutput, []rune(comment))
		h.TextChanged(context.Background(), e, nil)
		h.Apply(e)
		expect(e).To(haveMethodExecuted("SetSyntaxLayers", withArgs(
			contain(text.SyntaxLayer{
				Construct: theme.Comment,
				Spans:     []text.Span{{Start: 0, End: len(comment)}},
			}),
		)))
	})
}
