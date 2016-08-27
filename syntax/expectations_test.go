// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax_test

import "testing"

type ranger interface {
	Range() (int, int)
}

func expectPositionMatch(t *testing.T, src, match string, layer ranger) {
	expectIndexedPositionMatch(t, src, match, 0, layer)
}

func expectIndexedPositionMatch(t *testing.T, src, match string, idx int, layer ranger) {
	// Switch to []rune, since we want to match positions based on rune index, not byte
	runes, sep := []rune(src), []rune(match)

	var expectedStart int
	for ; idx >= 0; idx-- {
		expectedStart = index(runes, sep)
	}
	expectedEnd := expectedStart + len(sep)

	start, end := layer.Range()
	if start != expectedStart {
		t.Errorf("%s expected to be at position %d in %s (was at %d)", match, expectedStart, src, start)
	}
	if end != expectedEnd {
		t.Errorf("%s expected to end at position %d in %s (was at %d)", match, expectedEnd, src, end)
	}
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
