// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package suggestion

import (
	"math"
	"sort"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/score"
)

type suggestion struct {
	name  string
	score float64
}

// Adapter is an adapter that is based on gxui's
// FilteredListAdapter and CodeSuggestionAdapter.  There are some
// differences mostly revolving around displaying the suggestions.
type Adapter struct {
	gxui.DefaultAdapter
	pos         int
	suggestions []Suggestion
	scores      []float64
	end         int
}

func (a *Adapter) Set(pos int, suggestions ...Suggestion) {
	if a.pos == pos {
		return
	}
	a.pos = pos
	a.suggestions = suggestions
}

func (a *Adapter) Pos() int {
	return a.pos
}

func (a *Adapter) Sort(partial []rune) (longest int) {
	a.scores = make([]float64, len(a.suggestions))
	for i, suggestion := range a.suggestions {
		a.scores[i] = score.Score([]rune(suggestion.Name), partial)
		if a.scores[i] != math.MaxFloat64 {
			longest = len([]rune(suggestion.String()))
		}
	}

	a.end = len(a.suggestions)
	sort.Sort(a)
	for a.end > 0 && a.scores[a.end-1] == math.MaxFloat64 {
		a.end--
	}
	a.SetItems(a.suggestions[:a.end])
	return longest
}

func (a *Adapter) Len() int {
	return a.end
}

func (a *Adapter) Less(i, j int) bool {
	return a.scores[i] < a.scores[j]
}

func (a *Adapter) Swap(i, j int) {
	a.suggestions[i], a.suggestions[j] = a.suggestions[j], a.suggestions[i]
	a.scores[i], a.scores[j] = a.scores[j], a.scores[i]
}
