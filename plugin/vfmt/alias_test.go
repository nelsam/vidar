// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package vfmt_test

import (
	"github.com/nelsam/hel/pers"
	"github.com/poy/onpar/expect"
	"github.com/poy/onpar/matchers"
)

// The following are aliases to make tests read cleaner.
var (
	equal        = matchers.Equal
	not          = matchers.Not
	haveOccurred = matchers.HaveOccurred

	haveMethodExecuted = pers.HaveMethodExecuted
	withArgs           = pers.WithArgs
)

type expectation = expect.Expectation
