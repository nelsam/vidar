// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package vsyntax_test

import (
	"github.com/nelsam/hel/pers"
	"github.com/poy/onpar/expect"
	"github.com/poy/onpar/matchers"
)

type expectation = expect.Expectation

var (
	equal   = matchers.Equal
	contain = matchers.Contain

	haveMethodExecuted = pers.HaveMethodExecuted
	withArgs           = pers.WithArgs
)
