// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax

import (
	"math/rand"

	"github.com/nelsam/gxui"
)

var (
	DefaultRainbow = &Rainbow{
		colorRange: defaultRange,
		available: []Color{
			{
				Foreground: gxui.Color{
					R: 0.7,
					G: 0.3,
					B: 0.3,
					A: 1,
				},
			},
			{
				Foreground: gxui.Color{
					R: 0.3,
					G: 0.7,
					B: 0.3,
					A: 1,
				},
			},
			{
				Foreground: gxui.Color{
					R: 0.3,
					G: 0.3,
					B: 0.7,
					A: 1,
				},
			},
			{
				Foreground: gxui.Color{
					R: 0.7,
					G: 0.7,
					B: 0.3,
					A: 1,
				},
			},
			{
				Foreground: gxui.Color{
					R: 0.7,
					G: 0.3,
					B: 0.7,
					A: 1,
				},
			},
			{
				Foreground: gxui.Color{
					R: 0.3,
					G: 0.7,
					B: 0.7,
					A: 1,
				},
			},
		},
	}

	defaultRange = ColorRange{
		min: Color{
			Foreground: gxui.Color{
				R: 0.3,
				G: 0.3,
				B: 0.3,
				A: 1,
			},
		},
		max: Color{
			Foreground: gxui.Color{
				R: 0.7,
				G: 0.7,
				B: 0.7,
				A: 1,
			},
		},
	}
)

type ColorRange struct {
	min, max Color
}

func (c ColorRange) New() Color {
	return Color{
		Foreground: gxui.Color{
			R: c.min.Foreground.R + rand.Float32()*(c.max.Foreground.R-c.min.Foreground.R),
			G: c.min.Foreground.G + rand.Float32()*(c.max.Foreground.G-c.min.Foreground.G),
			B: c.min.Foreground.B + rand.Float32()*(c.max.Foreground.B-c.min.Foreground.B),
			A: c.min.Foreground.A + rand.Float32()*(c.max.Foreground.A-c.min.Foreground.A),
		},
		Background: gxui.Color{
			R: c.min.Background.R + rand.Float32()*(c.max.Background.R-c.min.Background.R),
			G: c.min.Background.G + rand.Float32()*(c.max.Background.G-c.min.Background.G),
			B: c.min.Background.B + rand.Float32()*(c.max.Background.B-c.min.Background.B),
			A: c.min.Background.A + rand.Float32()*(c.max.Background.A-c.min.Background.A),
		},
	}
}

type Rainbow struct {
	colorRange       ColorRange
	available, inUse []Color
}

func (r *Rainbow) New() Color {
	next := r.colorRange.New()
	if len(r.available) > 0 {
		lastIdx := len(r.available) - 1
		next = r.available[lastIdx]
		r.available = r.available[:lastIdx]
	}
	r.inUse = append(r.inUse, next)
	return next
}

func (r *Rainbow) Pop() Color {
	lastIdx := len(r.inUse) - 1
	color := r.inUse[lastIdx]
	r.inUse = r.inUse[:lastIdx]
	r.available = append(r.available, color)
	return color
}
