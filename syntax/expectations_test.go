// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax_test

import (
	"errors"
	"fmt"

	"github.com/nelsam/vidar/commander/input"
)

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

type position struct {
	src, match string
	idx        int
}

func (p position) Match(actual interface{}) error {
	span, ok := actual.(input.Span)
	if !ok {
		return errors.New("be of type input.Span")
	}
	// Switch to []rune, since we want to match positions based on rune index, not byte
	runes, sep := []rune(p.src), []rune(p.match)
	humanIdx := fmt.Sprintf("%d%s", p.idx+1, numSuffix(p.idx+1))

	var expectedStart, expectedEnd int
	for ; p.idx >= 0; p.idx-- {
		expectedStart = expectedEnd + index(runes[expectedEnd:], sep)
		if expectedStart == -1 {
			return fmt.Errorf("have a %s %s in %s", humanIdx, p.match, p.src)
		}
		expectedEnd = expectedStart + len(sep)
	}

	if span.Start != expectedStart {
		return fmt.Errorf("have a start position of %d for the %s %s in %s", expectedStart, humanIdx, p.match, p.src)
	}
	if span.End != expectedEnd {
		return fmt.Errorf("have an end position of %d for the %s %s in %s", expectedEnd, humanIdx, p.match, p.src)
	}
	return nil
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
