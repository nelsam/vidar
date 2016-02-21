// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax

import (
	"math/rand"

	"github.com/nelsam/gxui"
)

var (
	defaultRainbow = &rainbow{
		colorRange: defaultRange,
		available: []gxui.Color{
			{
				R: 0.7,
				G: 0.3,
				B: 0.3,
				A: 1,
			},
			{
				R: 0.3,
				G: 0.7,
				B: 0.3,
				A: 1,
			},
			{
				R: 0.3,
				G: 0.3,
				B: 0.7,
				A: 1,
			},
			{
				R: 0.7,
				G: 0.7,
				B: 0.3,
				A: 1,
			},
			{
				R: 0.7,
				G: 0.3,
				B: 0.7,
				A: 1,
			},
			{
				R: 0.3,
				G: 0.7,
				B: 0.7,
				A: 1,
			},
		},
	}

	defaultRange = colorRange{
		min: gxui.Color{
			R: 0.3,
			G: 0.3,
			B: 0.3,
			A: 1,
		},
		max: gxui.Color{
			R: 0.7,
			G: 0.7,
			B: 0.7,
			A: 1,
		},
	}
)

type colorRange struct {
	min, max gxui.Color
}

func (c colorRange) New() gxui.Color {
	return gxui.Color{
		R: c.min.R + rand.Float32()*(c.max.R-c.min.R),
		G: c.min.G + rand.Float32()*(c.max.G-c.min.G),
		B: c.min.B + rand.Float32()*(c.max.B-c.min.B),
		A: c.min.A + rand.Float32()*(c.max.A-c.min.A),
	}
}

type rainbow struct {
	colorRange       colorRange
	available, inUse []gxui.Color
}

func (r *rainbow) New() gxui.Color {
	next := r.colorRange.New()
	if len(r.available) > 0 {
		lastIdx := len(r.available) - 1
		next = r.available[lastIdx]
		r.available = r.available[:lastIdx]
	}
	r.inUse = append(r.inUse, next)
	return next
}

func (r *rainbow) Pop() gxui.Color {
	lastIdx := len(r.inUse) - 1
	color := r.inUse[lastIdx]
	r.inUse = r.inUse[:lastIdx]
	r.available = append(r.available, color)
	return color
}
