// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package theme

var Default = Theme{
	Rainbow: DefaultRainbow,
	Constructs: ConstructHighlights{
		Bad: Highlight{
			Foreground: Color{
				R: 0.3,
				G: 0,
				B: 0,
				A: 1,
			},
			Background: Color{
				R: 0.9,
				G: 0,
				B: 0.2,
				A: 1,
			},
		},
		Ident: Highlight{Foreground: Color{
			R: 0.9,
			G: 0.9,
			B: 0.9,
			A: 1,
		}},
		Builtin: Highlight{Foreground: Color{
			R: 0.9,
			G: 0.3,
			B: 0,
			A: 1,
		}},
		Nil: Highlight{Foreground: Color{
			R: 1,
			G: 0.2,
			B: 0.2,
			A: 1,
		}},
		Keyword: Highlight{Foreground: Color{
			R: 0,
			G: 0.6,
			B: 0.8,
			A: 1,
		}},
		Func: Highlight{Foreground: Color{
			R: 0.3,
			G: 0.6,
			B: 0,
			A: 1,
		}},
		Type: Highlight{Foreground: Color{
			R: 0.3,
			G: 0.6,
			B: 0.3,
			A: 1,
		}},
		String: Highlight{Foreground: Color{
			R: 0,
			G: 0.8,
			B: 0,
			A: 1,
		}},
		Num: Highlight{Foreground: Color{
			R: 0.6,
			G: 0,
			B: 0.1,
			A: 1,
		}},
		Comment: Highlight{Foreground: Color{
			R: 0.6,
			G: 0.6,
			B: 0.6,
			A: 1.0,
		}},
	},
}
