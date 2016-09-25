// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package suggestions

import (
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
}

func (a *Adapter) SetSuggestions(suggestions []gxui.CodeSuggestion) {
	a.suggestions = suggestions
	a.DefaultAdapter.SetItems(suggestions)
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
	a.DefaultAdapter.SetItems(a.suggestions)
}

func (a *Adapter) Len() int {
	return len(a.suggestions)
}

func (a *Adapter) Less(i, j int) bool {
	return a.scores[i] < a.scores[j]
}

func (a *Adapter) Swap(i, j int) {
	a.suggestions[i], a.suggestions[j] = a.suggestions[j], a.suggestions[i]
	a.scores[i], a.scores[j] = a.scores[j], a.scores[i]
}
