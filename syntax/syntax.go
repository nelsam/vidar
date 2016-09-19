// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax

import "github.com/nelsam/gxui"

var DefaultTheme = Theme{
	Rainbow: DefaultRainbow,
	Colors: Colors{
		Bad: Color{
			Foreground: gxui.Color{
				R: 0.3,
				G: 0,
				B: 0,
				A: 1,
			},
			Background: gxui.Color{
				R: 0.9,
				G: 0,
				B: 0.2,
				A: 1,
			},
		},
		Ident: Color{Foreground: gxui.Color{
			R: 0.9,
			G: 0.9,
			B: 0.9,
			A: 1,
		}},
		Builtin: Color{Foreground: gxui.Color{
			R: 0.9,
			G: 0.3,
			B: 0,
			A: 1,
		}},
		Nil: Color{Foreground: gxui.Color{
			R: 1,
			G: 0.2,
			B: 0.2,
			A: 1,
		}},
		Keyword: Color{Foreground: gxui.Color{
			R: 0,
			G: 0.6,
			B: 0.8,
			A: 1,
		}},
		Func: Color{Foreground: gxui.Color{
			R: 0.3,
			G: 0.6,
			B: 0,
			A: 1,
		}},
		Type: Color{Foreground: gxui.Color{
			R: 0.3,
			G: 0.6,
			B: 0.3,
			A: 1,
		}},
		String: Color{Foreground: gxui.Color{
			R: 0,
			G: 0.8,
			B: 0,
			A: 1,
		}},
		Num: Color{Foreground: gxui.Color{
			R: 0.6,
			G: 0,
			B: 0.1,
			A: 1,
		}},
		Comment: Color{Foreground: gxui.Gray60},
	},
}
