// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax_test

import (
	"errors"
	"fmt"

	"github.com/nelsam/vidar/commander/text"
	"github.com/nelsam/vidar/theme"
)

func findLayer(c theme.LanguageConstruct, l []text.SyntaxLayer) text.SyntaxLayer {
	for _, layer := range l {
		if layer.Construct == c {
			return layer
		}
	}
	return text.SyntaxLayer{}
}

func numSuffix(n int) string {
	switch n % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}

type matchPosition struct {
	src, match string
	idx        int
}

func (p matchPosition) Match(actual interface{}) (interface{}, error) {
	span, ok := actual.(text.Span)
	if !ok {
		return actual, errors.New("be of type text.Span")
	}
	// Switch to []rune, since we want to match positions based on rune index, not byte
	runes, sep := []rune(p.src), []rune(p.match)
	humanIdx := fmt.Sprintf("%d%s", p.idx+1, numSuffix(p.idx+1))

	var expectedStart, expectedEnd int
	for ; p.idx >= 0; p.idx-- {
		expectedStart = expectedEnd + index(runes[expectedEnd:], sep)
		if expectedStart == -1 {
			return actual, fmt.Errorf("have a %s %s in %s", humanIdx, p.match, p.src)
		}
		expectedEnd = expectedStart + len(sep)
	}

	if span.Start != expectedStart {
		return actual, fmt.Errorf("expected the start position %d to be %d for the %s %s in %s", span.Start, expectedStart, humanIdx, p.match, p.src)
	}
	if span.End != expectedEnd {
		return actual, fmt.Errorf("expected the end position %d to be %d for the %s %s in %s", span.End, expectedEnd, humanIdx, p.match, p.src)
	}
	return actual, nil
}

// index is like strings.Index, but for rune slices.
func index(runes, sep []rune) int {
	for i := range runes {
		matched := true
		for j, s := range sep {
			if runes[i+j] != s {
				matched = false
				break
			}
		}
		if matched {
			return i
		}
	}
	return -1
}
