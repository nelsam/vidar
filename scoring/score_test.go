// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package scoring_test

import (
	"testing"

	"github.com/nelsam/vidar/scoring"
	"github.com/poy/onpar/expect"
	"github.com/poy/onpar/matchers"
)

func TestSort(t *testing.T) {
	expect := expect.New(t)

	v := []string{
		"SomeLongThing",
		"Something",
		"Thing",
		"thing",
		"thisIsBad",
		"thingimajigger",
		"thingy",
		"thisIsNotAGoodMatch",
		"Thang",
		"tang",
		"bacon",
		"eggs",
	}
	expect(scoring.Sort(v, "thing")).To(matchers.Equal([]string{
		"thing",
		"Thing",
		"thingy",
		"thingimajigger",
		"Thang",
		"tang",
		"thisIsBad",
		"thisIsNotAGoodMatch",
		"Something",
		"SomeLongThing",
	}))
}
