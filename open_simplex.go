package noisey

/* Copyright 2014, Timothy Bogdala <tdb@animal-machine.com>
See the LICENSE file for more details. */

/*

This is a port of the OpenSimplex noise algorithm. A lot of the comments in
this code are pulled from the original sources.

The following references were used to implement this algorithm:

* Initial blog post: http://uniblock.tumblr.com/post/97868843242/noise
  and the following update: http://uniblock.tumblr.com/post/99279694832/2d-and-4d-noise-too

* Initial Java implemention by Kurt Spencer: https://gist.github.com/KdotJPG/b1270127455a94ac5d19

* Port to C by Stephen M. Cameron: https://github.com/smcameron/open-simplex-noise-in-c/blob/master/open-simplex-noise.c

*/

import (
	"math"
)

const (
	permTableSize     = 256
	stretchConstant2D = -0.211324865405187 // (1 / sqrt(2 + 1) - 1 ) / 2;
	squishConstant2D  = 0.366025403784439  // (sqrt(2 + 1) -1) / 2;
	stretchConstant3D = -1.0 / 6.0         // (1 / sqrt(3 + 1) - 1) / 3;
	squishConstant3D  = 1.0 / 3.0          // (sqrt(3+1)-1)/3;
	normConstant2D    = 47.0
	normConstant3D    = 103.0
)

var (
	// Gradients for 2D. They approximate the directions to the
	// vertices of an octagon from the center.
	gradients2D = []int8{
		5, 2, 2, 5,
		-5, 2, -2, 5,
		5, -2, 2, -5,
		-5, -2, -2, -5,
	}

	// Gradients for 3D. They approximate the directions to the
	// vertices of a rhombicuboctahedron from the center, skewed so
	// that the triangular and square facets can be inscribed inside
	// circles of the same radius.
	gradients3D = []int8{
		-11, 4, 4, -4, 11, 4, -4, 4, 11,
		11, 4, 4, 4, 11, 4, 4, 4, 11,
		-11, -4, 4, -4, -11, 4, -4, -4, 11,
		11, -4, 4, 4, -11, 4, 4, -4, 11,
		-11, 4, -4, -4, 11, -4, -4, 4, -11,
		11, 4, -4, 4, 11, -4, 4, 4, -11,
		-11, -4, -4, -4, -11, -4, -4, -4, -11,
		11, -4, -4, 4, -11, -4, 4, -4, -11,
	}
)

// OpenSimplexGenerator2D stores the state information for generating opensimplex noise.
type OpenSimplexGenerator2D struct {
	Rng             RandomSource // random number generator interface
	Permutations    []int        // the random permutation table
	PermGradIndex3D []int
}

// NewOpenSimplexGenerator2D creates a new state object for the 2D opensimplex noise generator
func NewOpenSimplexGenerator2D(rng RandomSource) (osg OpenSimplexGenerator2D) {
	osg.Rng = rng
	osg.Permutations = rng.Perm(permTableSize)

	// construct the gradient index table
	osg.PermGradIndex3D = make([]int, permTableSize)
	gradLengthDiv3 := len(gradients3D) / 3
	for i := range osg.PermGradIndex3D {
		osg.PermGradIndex3D[i] = (osg.Permutations[i] % gradLengthDiv3) * 3
	}

	return
}

func (osg *OpenSimplexGenerator2D) extrapolate2(xsb int, ysb int, dx float64, dy float64) float64 {
	index := osg.Permutations[(osg.Permutations[xsb&0xFF]+ysb)&0xFF] & 0x0E
	return float64(gradients2D[index])*dx + float64(gradients2D[index+1])*dy
}

// Get2D calculates the noise at a given 2D coordinate
func (osg *OpenSimplexGenerator2D) Get2D(x float64, y float64) float64 {
	// place input coordinates onto grid
	stretchOffset := (x + y) * stretchConstant2D
	xs := x + stretchOffset
	ys := y + stretchOffset

	// floor to get grid coordinates of rhombus (stretched square) super-cell origin
	xsb := int(math.Floor(xs))
	ysb := int(math.Floor(ys))

	// skew out to get actual coordinates of rhombus origin. we'll need these later
	squishOffset := float64(xsb+ysb) * squishConstant2D
	xb := float64(xsb) + squishOffset
	yb := float64(ysb) + squishOffset

	// compute grid coordinates relative to rhombus origin
	xins := xs - float64(xsb)
	yins := ys - float64(ysb)

	// sum those together to get a value that determines which region we're in
	inSum := xins + yins

	// positions relative to origin point
	dx0 := x - xb
	dy0 := y - yb

	// we'll be defining these inside the next block and using them afterwards
	var dx_ext, dy_ext float64
	var xsv_ext, ysv_ext int

	var value float64

	// contribution (1,0)
	dx1 := dx0 - 1 - squishConstant2D
	dy1 := dy0 - 0 - squishConstant2D
	attn1 := 2 - dx1*dx1 - dy1*dy1
	if attn1 > 0 {
		attn1 *= attn1
		value += attn1 * attn1 * osg.extrapolate2(xsb+1, ysb, dx1, dy1)
	}

	// contribution (0,1)
	dx2 := dx0 - 0 - squishConstant2D
	dy2 := dy0 - 1 - squishConstant2D
	attn2 := 2 - dx2*dx2 - dy2*dy2
	if attn2 > 0 {
		attn2 *= attn2
		value += attn2 * attn2 * osg.extrapolate2(xsb, ysb+1, dx2, dy2)
	}

	if inSum <= 1 { // we're inside the triangle (2-Simplex) at (0,0)
		zins := 1 - inSum
		if (zins > xins) || (zins > yins) { // (0,0) is one of the closest two triangle vertices
			if xins > yins {
				xsv_ext = xsb + 1
				ysv_ext = ysb - 1
				dx_ext = dx0 - 1
				dy_ext = dy0 + 1
			} else {
				xsv_ext = xsb - 1
				ysv_ext = ysb + 1
				dx_ext = dx0 + 1
				dy_ext = dy0 - 1
			}
		} else { // (1,0) and (0,1) are the closest two vertices
			xsv_ext = xsb + 1
			ysv_ext = ysb + 1
			dx_ext = dx0 - 1 - 2*squishConstant2D
			dy_ext = dy0 - 1 - 2*squishConstant2D
		}
	} else { // we're inside the triangle (2-Simplex) at (1,1)
		zins := 2 - inSum
		if (zins < xins) || (zins < yins) { // (0,0) is one of the closest two triangle vertices
			if xins > yins {
				xsv_ext = xsb + 2
				ysv_ext = ysb
				dx_ext = dx0 - 2 - 2*squishConstant2D
				dy_ext = dy0 - 2*squishConstant2D
			} else {
				xsv_ext = xsb
				ysv_ext = ysb + 2
				dx_ext = dx0 - 2*squishConstant2D
				dy_ext = dy0 - 2 - 2*squishConstant2D
			}
		} else { // (1,0) and (0,1) are the closest two vertices
			xsv_ext = int(dx0)
			ysv_ext = int(dy0)
			dx_ext = float64(xsb)
			dy_ext = float64(ysb)
		}
		xsb += 1
		ysb += 1
		dx0 = dx0 - 1 - 2*squishConstant2D
		dy0 = dy0 - 1 - 2*squishConstant2D
	}

	// contribution (0,0) or (1,1)
	attn0 := 2 - dx0*dx0 - dy0*dy0
	if attn0 > 0 {
		attn0 *= attn0
		value += attn0 * attn0 * osg.extrapolate2(xsb, ysb, dx0, dy0)
	}

	// extra vertex
	attn_ext := 2 - dx_ext*dx_ext - dy_ext*dy_ext
	if attn_ext > 0 {
		attn_ext *= attn_ext
		value += attn_ext * attn_ext * osg.extrapolate2(xsv_ext, ysv_ext, dx_ext, dy_ext)
	}

	return value / normConstant2D
}
