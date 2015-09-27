// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package main

import (
	"time"

	"github.com/google/gxui"
)

type statefulEditor struct {
	gxui.CodeEditor

	lastModified time.Time
	hasChanges   bool
	filepath     string
}
