package noisey

/* Copyright 2014, Timothy Bogdala <tdb@animal-machine.com>
See the LICENSE file for more details. */

/*

The Perlin noise math FAQ:
http://webstaff.itn.liu.se/~stegu/TNM022-2005/perlinnoiselinks/perlin-noise-math-faq.html

Other helpful links:
  * http://www.angelcode.com/dev/perlin/perlin.html
  * https://code.google.com/p/fractalterraingeneration/wiki/Perlin_Noise#Algorithm
  * http://libnoise.sourceforge.net/noisegen/index.html

Based in part off of an implementation of perlin noise by github.com/nsf
posted here: https://gist.github.com/nsf/1170424#file-test-go

*/

import (
	"math"
)

// Enumeration controlling the quality of smoothing done in the noise calcualtions
const (
	FastQuality = iota
	StandardQuality
	HighQuality
)

const (
	tableSize = 256
)

// PerlinGenerator stores the state information for generating perlin noise.
type PerlinGenerator2D struct {
	Rng                 RandomSource // random number generator interface
	Permutations        []int        // the random permutation table
	RandomGradients     []Vec2f      // the random gradient table
	Quality             int          // controls the blending of values (FastQuality|StandardQuality|HighQuality)
	calculatedGradients [4]Vec2f     // the last calculated gradient table
	calculatedOrigins   [4]Vec2f     // the last calculated origins table
}

// makeRandomGradient2D creates a random gradient used for initializing a 2D perlin state
func (pg *PerlinGenerator2D) makeRandomGradient2D() Vec2f {
	v := pg.Rng.Float64() * math.Pi * 2
	return Vec2f{math.Cos(v), math.Sin(v)}
}

// NewPerlinGenerator2D creates a new state object for the 2D perlin noise generator
func NewPerlinGenerator2D(rng RandomSource, quality int) (pg PerlinGenerator2D) {
	pg.Rng = rng
	pg.Quality = quality
	pg.Permutations = rng.Perm(tableSize)
	pg.RandomGradients = make([]Vec2f, tableSize)
	for i := range pg.RandomGradients {
		pg.RandomGradients[i] = pg.makeRandomGradient2D()
	}
	return
}

func (pg *PerlinGenerator2D) calcGradient(x, y int) Vec2f {
	pMask := len(pg.Permutations) - 1
	i := pg.Permutations[x&pMask] + pg.Permutations[y&pMask]
	return pg.RandomGradients[i&pMask]
}

func (pg *PerlinGenerator2D) calcGradientsAndOrigins(x, y float64) {
	x0f := math.Floor(x)
	y0f := math.Floor(y)
	x0 := int(x0f)
	y0 := int(y0f)
	x1 := x0 + 1
	y1 := y0 + 1

	pg.calculatedGradients[0] = pg.calcGradient(x0, y0)
	pg.calculatedGradients[1] = pg.calcGradient(x1, y0)
	pg.calculatedGradients[2] = pg.calcGradient(x0, y1)
	pg.calculatedGradients[3] = pg.calcGradient(x1, y1)

	pg.calculatedOrigins[0] = Vec2f{x0f, y0f}
	pg.calculatedOrigins[1] = Vec2f{x0f + 1.0, y0f}
	pg.calculatedOrigins[2] = Vec2f{x0f, y0f + 1.0}
	pg.calculatedOrigins[3] = Vec2f{x0f + 1.0, y0f + 1.0}
}

func getPointGradient2D(origin, gradient, point Vec2f) float64 {
	s := Vec2f{point.X - origin.X, point.Y - origin.Y}
	return gradient.X*s.X + gradient.Y*s.Y
}

// standard quality smoothing with range from 0.0 to 1.0
func calcCubicSCurve(v float64) float64 {
	return v * v * (3 - 2*v)
}

// highest quality smoothing with range from 0.0 to 1.0
func calcQuinticSCurve(v float64) float64 {
	v3 := v * v * v
	v4 := v3 * v
	v5 := v4 * v
	return (6.0 * v5) - (15.0 * v4) + (10.0 * v3)
}

func lerp(a, b, v float64) float64 {
	return a*(1-v) + b*v
}

// GetValue2D calculates the perlin noise at a given 2D coordinate
func (pg *PerlinGenerator2D) GetValue2D(x float64, y float64) float64 {
	pg.calcGradientsAndOrigins(x, y)

	p := Vec2f{x, y}
	v0 := getPointGradient2D(pg.calculatedOrigins[0], pg.calculatedGradients[0], p)
	v1 := getPointGradient2D(pg.calculatedOrigins[1], pg.calculatedGradients[1], p)
	v2 := getPointGradient2D(pg.calculatedOrigins[2], pg.calculatedGradients[2], p)
	v3 := getPointGradient2D(pg.calculatedOrigins[3], pg.calculatedGradients[3], p)

	// smooth out the interpolation of the noise depending on the selected quality
	var fx, fy float64
	switch pg.Quality {
	case StandardQuality:
		fx = calcCubicSCurve(x - pg.calculatedOrigins[0].X)
		fy = calcCubicSCurve(y - pg.calculatedOrigins[0].Y)
	case HighQuality:
		fx = calcQuinticSCurve(x - pg.calculatedOrigins[0].X)
		fy = calcQuinticSCurve(y - pg.calculatedOrigins[0].Y)
	case FastQuality:
		fx = x - pg.calculatedOrigins[0].X
		fy = y - pg.calculatedOrigins[0].Y
	}

	vx0 := lerp(v0, v1, fx)
	vx1 := lerp(v2, v3, fx)
	return lerp(vx0, vx1, fy)
}
