// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// Package test includes some common test code specific to the
// vidar project.
package test

import (
	"github.com/fatih/color"
	"github.com/poy/onpar/diff"
)

func Differ() *diff.Differ {
	return diff.New(
		diff.Actual(diff.WithSprinter(color.New(color.FgRed))),
		diff.Expected(diff.WithSprinter(color.New(color.FgYellow))),
	)
}
