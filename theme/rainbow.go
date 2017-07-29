// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package theme

import (
	"math/rand"
)

var (
	DefaultRainbow = &Rainbow{
		Range: HighlightRange{
			Min: Highlight{
				Foreground: Color{
					R: 0.3,
					G: 0.3,
					B: 0.3,
					A: 1,
				},
			},
			Max: Highlight{
				Foreground: Color{
					R: 0.7,
					G: 0.7,
					B: 0.7,
					A: 1,
				},
			},
		},
		Available: []Highlight{
			{
				Foreground: Color{
					R: 0.7,
					G: 0.3,
					B: 0.3,
					A: 1,
				},
			},
			{
				Foreground: Color{
					R: 0.3,
					G: 0.7,
					B: 0.3,
					A: 1,
				},
			},
			{
				Foreground: Color{
					R: 0.3,
					G: 0.3,
					B: 0.7,
					A: 1,
				},
			},
			{
				Foreground: Color{
					R: 0.7,
					G: 0.7,
					B: 0.3,
					A: 1,
				},
			},
			{
				Foreground: Color{
					R: 0.7,
					G: 0.3,
					B: 0.7,
					A: 1,
				},
			},
			{
				Foreground: Color{
					R: 0.3,
					G: 0.7,
					B: 0.7,
					A: 1,
				},
			},
		},
	}
)

type HighlightRange struct {
	Min, Max Highlight
}

func (c HighlightRange) New() Highlight {
	return Highlight{
		Foreground: Color{
			R: c.Min.Foreground.R + rand.Float32()*(c.Max.Foreground.R-c.Min.Foreground.R),
			G: c.Min.Foreground.G + rand.Float32()*(c.Max.Foreground.G-c.Min.Foreground.G),
			B: c.Min.Foreground.B + rand.Float32()*(c.Max.Foreground.B-c.Min.Foreground.B),
			A: c.Min.Foreground.A + rand.Float32()*(c.Max.Foreground.A-c.Min.Foreground.A),
		},
		Background: Color{
			R: c.Min.Background.R + rand.Float32()*(c.Max.Background.R-c.Min.Background.R),
			G: c.Min.Background.G + rand.Float32()*(c.Max.Background.G-c.Min.Background.G),
			B: c.Min.Background.B + rand.Float32()*(c.Max.Background.B-c.Min.Background.B),
			A: c.Min.Background.A + rand.Float32()*(c.Max.Background.A-c.Min.Background.A),
		},
	}
}

type Rainbow struct {
	Range            HighlightRange
	Available, inUse []Highlight
}

func (r *Rainbow) Reset() {
	for i := len(r.inUse) - 1; i >= 0; i-- {
		r.Available = append(r.Available, r.inUse[i])
	}
	r.inUse = r.inUse[:0]
}

func (r *Rainbow) New() Highlight {
	next := r.Range.New()
	if len(r.Available) > 0 {
		lastIdx := len(r.Available) - 1
		next = r.Available[lastIdx]
		r.Available = r.Available[:lastIdx]
	}
	r.inUse = append(r.inUse, next)
	return next
}

func (r *Rainbow) Pop() Highlight {
	lastIdx := len(r.inUse) - 1
	color := r.inUse[lastIdx]
	r.inUse = r.inUse[:lastIdx]
	r.Available = append(r.Available, color)
	return color
}
