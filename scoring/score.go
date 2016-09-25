// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package scoring

import (
	"math"
	"sort"
	"unicode"
)

const (
	// Costs of operations, in order to make the sorting
	// feel natural.
	changeCase   = 0.5
	substitution = 4
	insertion    = 6
	deletion     = 8
	append       = 1
	prepend      = 10
)

func Sort(values []string, partial string) []string {
	v := sortable{
		values:  values,
		scores:  make([]float64, len(values)),
		partial: []rune(partial),
	}
	sort.Sort(v)
	return v.values
}

type sortable struct {
	values  []string
	scores  []float64
	partial []rune
}

func (s sortable) Len() int {
	return len(s.values)
}

func (s sortable) score(i int) float64 {
	if score := s.scores[i]; score > 0 {
		return score
	}
	s.scores[i] = Score([]rune(s.values[i]), s.partial)
	return s.scores[i]
}

func (s sortable) Less(i, j int) bool {
	return s.score(i) < s.score(j)
}

func (s sortable) Swap(i, j int) {
	s.values[i], s.values[j] = s.values[j], s.values[i]
	s.scores[i], s.scores[j] = s.scores[j], s.scores[i]
}

func Score(suggestion, partial []rune) float64 {
	best := math.MaxFloat64
	for i := len(suggestion) - 1; i >= 0; i-- {
		current := suggestion[i:]
		difficulty := float64(i) * prepend
		missed := 0
		for _, r := range partial {
			idx := -1
			for k, h := range current {
				if r == h {
					idx = k
					break
				}
				if unicode.ToLower(r) == unicode.ToLower(h) {
					idx = k
					difficulty += changeCase
					break
				}
			}
			if idx == -1 {
				difficulty += substitution
				missed++
				continue
			}
			inserted := float64(idx - missed)
			switch {
			case inserted > 0:
				difficulty += inserted * insertion
			case inserted < 0:
				difficulty += (0 - inserted) * deletion
			}
			missed = 0
			current = current[idx+1:]
		}
		difficulty += float64(len(current)) * append
		if difficulty < best {
			best = difficulty
		}
	}
	return best
}
