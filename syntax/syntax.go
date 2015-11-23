// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package syntax

import "github.com/nelsam/gxui"

var (
	badColor = gxui.Color{
		R: 0.3,
		G: 0,
		B: 0,
		A: 1,
	}
	badBackground = gxui.Color{
		R: 0.9,
		G: 0,
		B: 0.2,
		A: 1,
	}
	builtinColor = gxui.Color{
		R: 0.9,
		G: 0.3,
		B: 0,
		A: 1,
	}
	nilColor = gxui.Color{
		R: 1,
		G: 0.2,
		B: 0.2,
		A: 1,
	}
	keywordColor = gxui.Color{
		R: 0,
		G: 0.6,
		B: 0.8,
		A: 1,
	}
	functionColor = gxui.Color{
		R: 0.3,
		G: 0.6,
		B: 0,
		A: 1,
	}
	typeColor = gxui.Color{
		R: 0.3,
		G: 0.6,
		B: 0.3,
		A: 1,
	}
	stringColor = gxui.Color{
		R: 0,
		G: 0.8,
		B: 0,
		A: 1,
	}
	numColor = gxui.Color{
		R: 0.6,
		G: 0,
		B: 0.1,
		A: 1,
	}
	commentColor = gxui.Gray60
)
