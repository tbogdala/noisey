package noisey

/* Copyright 2014, Timothy Bogdala <tdb@animal-machine.com>
See the LICENSE file for more details. */

type BuilderSource2D interface {
	Get(float64, float64) float64
}

type Builder2DBounds struct {
	MinX, MinY, MaxX, MaxY float64
}

type Builder2D struct {
	Source BuilderSource2D
	Width  int
	Height int
	Bounds Builder2DBounds
	Values []float64
}

func NewBuilder2D(s BuilderSource2D, width int, height int) (b Builder2D) {
	b.Source = s
	b.Width = width
	b.Height = height
	b.Values = make([]float64, width*height)
	return
}

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
			value := b.Source.Get(xCur, yCur)
			b.Values[(y*b.Width)+x] = value
			xCur += xDelta
		}
		yCur += yDelta
	}
}
