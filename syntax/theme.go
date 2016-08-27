// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax

import (
	"github.com/nelsam/gxui"
)

type Color struct {
	Foreground, Background gxui.Color
}

type Colors struct {
	Keyword Color
	Builtin Color
	Func    Color
	Type    Color

	String Color
	Num    Color
	Nil    Color

	Comment Color

	Bad Color
}

type Theme struct {
	Colors  Colors
	Rainbow *Rainbow
}
