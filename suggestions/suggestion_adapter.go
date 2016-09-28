// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package suggestions

import (
	"math"
	"sort"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/scoring"
)

// Adapter is an adapter that is based on gxui's
// FilteredListAdapter and CodeSuggestionAdapter.  There are some
// differences mostly revolving around displaying the suggestions.
type Adapter struct {
	gxui.DefaultAdapter
	suggestions []gxui.CodeSuggestion
	scores      []float64
	end         int
}

func (a *Adapter) SetSuggestions(suggestions []gxui.CodeSuggestion) {
	a.suggestions = suggestions
	a.DefaultAdapter.SetItems(suggestions)
	a.end = len(suggestions)
}

func (a *Adapter) Suggestion(item gxui.AdapterItem) gxui.CodeSuggestion {
	return item.(gxui.CodeSuggestion)
}

func (a *Adapter) Sort(partial string) {
	a.scores = make([]float64, len(a.suggestions))
	match := []rune(partial)
	for i, suggestion := range a.suggestions {
		a.scores[i] = scoring.Score([]rune(suggestion.Name()), match)
	}
	sort.Sort(a)

	a.end = len(a.suggestions)
	for a.end > 0 && a.scores[a.end-1] == math.MaxFloat64 {
		a.end--
	}
	a.DefaultAdapter.SetItems(a.suggestions[:a.end])
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
