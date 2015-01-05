package noisey

/* Copyright 2014, Timothy Bogdala <tdb@animal-machine.com>
See the LICENSE file for more details. */

/* This module contains code to easily build 'maps' of random noise. */

import "math"

// BuilderSource2D is an interface defining how the Builder* types get noise.
type BuilderSource2D interface {
	Get2D(float64, float64) float64
}

// Builder2DBounds is a simple rectangle type.
type Builder2DBounds struct {
	MinX, MinY, MaxX, MaxY float64
}

// Builder2D contains the parameters and data for the noise 'map' generated with Build().
type Builder2D struct {
	Source BuilderSource2D
	Width  int
	Height int
	Bounds Builder2DBounds
	Values []float64
}

// NewBuilder2D creates a new 2D noise 'map' builder of the given size
func NewBuilder2D(s BuilderSource2D, width int, height int) (b Builder2D) {
	b.Source = s
	b.Width = width
	b.Height = height
	b.Values = make([]float64, width*height)
	return
}

// Build gets noise from Source for each spot in the data array. These steps
// are real numbers so that Bounds does not have to match Width/Height.
func (b *Builder2D) Build() {
	// setup the initial parameters controlling how the noise is sampled
	xExtent := b.Bounds.MaxX - b.Bounds.MinX
	yExtent := b.Bounds.MaxY - b.Bounds.MinY
	xDelta := xExtent / float64(b.Width)
	yDelta := yExtent / float64(b.Height)
	xCur := b.Bounds.MinX
	yCur := b.Bounds.MinY

	for y := 0; y < b.Height; y++ {
		xCur = b.Bounds.MinX
		for x := 0; x < b.Width; x++ {
			value := b.Source.Get2D(xCur, yCur)
			b.Values[(y*b.Width)+x] = value
			xCur += xDelta
		}
		yCur += yDelta
	}
}

// GetMinMax returns the lowest and the highest Values
func (b *Builder2D) GetMinMax() (min float64, max float64) {
	var low float64 = math.MaxFloat64
	var high float64 = math.SmallestNonzeroFloat64

	totalIndex := b.Width * b.Height
	for i := 0; i < totalIndex; i++ {
		v := b.Values[i]
		if v < low {
			low = v
		}
		if v > high {
			high = v
		}
	}

	return low, high
}
