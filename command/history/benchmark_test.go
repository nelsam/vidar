// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package history_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/nelsam/vidar/command/history"
	"github.com/nelsam/vidar/commander/input"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func BenchmarkHistorySimple(b *testing.B) {
	b.StopTimer()
	all := history.Bindables(nil, nil, nil)
	hist := findHistory(b, all)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		hist.TextChanged(nil, input.Edit{At: 210, Old: []rune(""), New: []rune("v")})
		if i%16 == 0 {
			hist.TextChanged(nil, hist.Rewind())
			if i%64 == 0 {
				hist.TextChanged(nil, hist.FastForward(0))
			}
		}
	}
}

func BenchmarkHistoryBranches(b *testing.B) {
	b.StopTimer()
	all := history.Bindables(nil, nil, nil)
	hist := findHistory(b, all)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		hist.TextChanged(nil, input.Edit{At: 210, Old: []rune(""), New: []rune("v")})
		hist.TextChanged(nil, hist.Rewind())
	}
}

func BenchmarkHistoryFF(b *testing.B) {
	b.StopTimer()
	all := history.Bindables(nil, nil, nil)
	hist := findHistory(b, all)
	hist.TextChanged(nil, input.Edit{At: 210, Old: []rune(""), New: []rune("v")})
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		hist.TextChanged(nil, hist.Rewind())
		hist.TextChanged(nil, hist.FastForward(0))
	}
}

func BenchmarkHistoryExtreme(b *testing.B) {
	b.StopTimer()
	all := history.Bindables(nil, nil, nil)
	hist := findHistory(b, all)
	for i := 0; i < 100000; i++ {
		hist.TextChanged(nil, input.Edit{At: 210, Old: []rune("a quick brown fox"), New: []rune("a silent, deadly wolf")})
		if rand.Intn(100) == 0 {
			undos := rand.Intn(50)
			for i := 0; i < undos; i++ {
				hist.TextChanged(nil, hist.Rewind())
			}
			var redos int
			if undos > 0 {
				redos = rand.Intn(undos)
			}
			for i := 0; i < redos; i++ {
				hist.TextChanged(nil, hist.FastForward(0))
			}
		}
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		hist.TextChanged(nil, input.Edit{At: 210, Old: []rune(""), New: []rune("v")})
		if i%16 == 0 {
			hist.TextChanged(nil, hist.Rewind())
			if i%64 == 0 {
				hist.TextChanged(nil, hist.FastForward(0))
			}
		}
	}
}
