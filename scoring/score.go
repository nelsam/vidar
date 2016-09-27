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
	changeCase   = 0.1
	substitution = 6
	insertion    = 4
	deletion     = 6
	append       = 0.2
	prepend      = 10

	// minMatchPercent is the percentage of a partial that
	// must match (case insensitive) for an iteration's
	// score to be counted.
	minMatchPercent = 0.5
)

// Sort reorganizes values using Score.
func Sort(values []string, partial string) []string {
	v := &sortable{
		values:  values,
		scores:  make([]float64, len(values)),
		partial: []rune(partial),
	}
	sort.Sort(v)
	v.trim()
	return v.values
}

type sortable struct {
	values  []string
	scores  []float64
	partial []rune
}

func (s *sortable) Len() int {
	return len(s.values)
}

func (s *sortable) score(i int) float64 {
	if score := s.scores[i]; score > 0 {
		return score
	}
	s.scores[i] = Score([]rune(s.values[i]), s.partial)
	return s.scores[i]
}

func (s *sortable) Less(i, j int) bool {
	return s.score(i) < s.score(j)
}

func (s *sortable) Swap(i, j int) {
	s.values[i], s.values[j] = s.values[j], s.values[i]
	s.scores[i], s.scores[j] = s.scores[j], s.scores[i]
}

func (s *sortable) trim() {
	for i := len(s.values) - 1; i >= 0; i-- {
		if s.score(i) != math.MaxFloat64 {
			break
		}
		s.scores = s.scores[:i]
		s.values = s.values[:i]
	}
}

func Score(suggestion, partial []rune) float64 {
	best := math.MaxFloat64
	for i := len(suggestion) - 1; i >= 0; i-- {
		current := suggestion[i:]
		difficulty := float64(i) * prepend
		missed := float64(0)
		matched := 0
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
				missed++
				continue
			}
			matched++
			switch {
			case matched == 1 && missed < float64(idx):
				difficulty += missed*substitution + (float64(idx)-missed)*prepend
			case missed <= float64(idx):
				difficulty += missed*substitution + (float64(idx)-missed)*insertion
			case missed > float64(idx):
				difficulty += float64(idx)*substitution + (missed-float64(idx))*deletion
			}
			missed = 0
			current = current[idx+1:]
		}
		if float64(matched)/float64(len(partial)) < minMatchPercent {
			continue
		}
		remainder := float64(len(current))
		switch {
		case missed > 0 && missed <= remainder:
			difficulty += missed*substitution + (remainder-missed)*append
		case missed > remainder:
			difficulty += remainder*substitution + (missed-remainder)*deletion
		}
		difficulty += float64(len(current)) * append
		if difficulty < best {
			best = difficulty
		}
	}
	return best
}
