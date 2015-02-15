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

// OpenSimplexGenerator stores the state information for generating opensimplex noise.
type OpenSimplexGenerator struct {
	Rng             RandomSource // random number generator interface
	Permutations    []int        // the random permutation table
	PermGradIndex3D []int
}

// NewOpenSimplexGenerator creates a new state object for the open simplex noise generator
func NewOpenSimplexGenerator(rng RandomSource) (osg OpenSimplexGenerator) {
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

func (osg *OpenSimplexGenerator) extrapolate2(xsb int, ysb int, dx float64, dy float64) float64 {
	index := osg.Permutations[(osg.Permutations[xsb&0xFF]+ysb)&0xFF] & 0x0E
	return float64(gradients2D[index])*dx + float64(gradients2D[index+1])*dy
}

// Get2D calculates the noise at a given 2D coordinate
func (osg *OpenSimplexGenerator) Get2D(x float64, y float64) float64 {
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
			dx_ext = dx0
			dy_ext = dy0
			xsv_ext = xsb
			ysv_ext = ysb
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

func (osg *OpenSimplexGenerator) extrapolate3(xsb int, ysb int, zsb int, dx float64, dy float64, dz float64) float64 {
	px := osg.Permutations[xsb&0xFF]
	py := osg.Permutations[(px+ysb)&0xFF]
	index := osg.PermGradIndex3D[(py+zsb)&0xFF]
	return float64(gradients3D[index])*dx + float64(gradients3D[index+1])*dy + float64(gradients3D[index+2])*dz
}

// Get3D calculates the noise at a given 3D coordinate
func (osg *OpenSimplexGenerator) Get3D(x float64, y float64, z float64) float64 {
	// Place input coordinates on simplectic honeycomb
	stretchOffset := (x + y + z) * stretchConstant3D
	xs := x + stretchOffset
	ys := y + stretchOffset
	zs := z + stretchOffset

	// Floor to get simplectic honeycomb coordinates of rhombohedron (stretched cube) super-cell origin.
	xsb := int(math.Floor(xs))
	ysb := int(math.Floor(ys))
	zsb := int(math.Floor(zs))

	// Skew out to get actual coordinates of rhombohedron origin. We'll need these later.
	squishOffset := float64(xsb+ysb+zsb) * squishConstant3D
	xb := float64(xsb) + squishOffset
	yb := float64(ysb) + squishOffset
	zb := float64(zsb) + squishOffset

	// Compute simplectic honeycomb coordinates relative to rhombohedral origin.
	xins := xs - float64(xsb)
	yins := ys - float64(ysb)
	zins := zs - float64(zsb)

	// Sum those together to get a value that determines which region we're in.
	var inSum float64 = xins + yins + zins

	// Positions relative to origin point.
	dx0 := x - xb
	dy0 := y - yb
	dz0 := z - zb

	// We'll be defining these inside the next block and using them afterwards.
	var dx_ext0, dy_ext0, dz_ext0 float64
	var dx_ext1, dy_ext1, dz_ext1 float64
	var xsv_ext0, ysv_ext0, zsv_ext0 int
	var xsv_ext1, ysv_ext1, zsv_ext1 int

	var value float64 = 0.0
	if inSum <= 1.0 { // We're inside the tetrahedron (3-Simplex) at (0,0,0)
		// Determine which two of (0,0,1), (0,1,0), (1,0,0) are closest.
		var aPoint byte = 0x01
		var aScore float64 = xins
		var bPoint byte = 0x02
		var bScore float64 = yins
		if aScore >= bScore && zins > bScore {
			bScore = zins
			bPoint = 0x04
		} else if aScore < bScore && zins > aScore {
			aScore = zins
			aPoint = 0x04
		}

		// Now we determine the two lattice points not part of the tetrahedron that may contribute.
		// This depends on the closest two tetrahedral vertices, including (0,0,0)
		var c byte
		var wins float64 = 1.0 - inSum
		if wins > aScore || wins > bScore { // (0,0,0) is one of the closest two tetrahedral vertices.
			// Our other closest vertex is the closest out of a and b.
			if bScore > aScore {
				c = bPoint
			} else {
				c = aPoint
			}

			if c&0x01 == 0 {
				xsv_ext0 = xsb - 1
				xsv_ext1 = xsb
				dx_ext0 = dx0 + 1.0
				dx_ext1 = dx0
			} else {
				xsv_ext0 = xsb + 1
				xsv_ext1 = xsv_ext0
				dx_ext0 = dx0 - 1.0
				dx_ext1 = dx_ext0
			}

			if c&0x02 == 0 {
				ysv_ext0 = ysb
				ysv_ext1 = ysb
				dy_ext0 = dy0
				dy_ext1 = dy0
				if c&0x01 == 0 {
					ysv_ext1 -= 1
					dy_ext1 += 1.0
				} else {
					ysv_ext0 -= 1
					dy_ext0 += 1.0
				}
			} else {
				ysv_ext0 = ysb + 1
				ysv_ext1 = ysv_ext0
				dy_ext0 = dy0 - 1.0
				dy_ext1 = dy_ext0
			}

			if c&0x04 == 0 {
				zsv_ext0 = zsb
				zsv_ext1 = zsb - 1
				dz_ext0 = dz0
				dz_ext1 = dz0 + 1.0
			} else {
				zsv_ext0 = zsb + 1
				zsv_ext1 = zsv_ext0
				dz_ext0 = dz0 - 1.0
				dz_ext1 = dz_ext0
			}
		} else { // (0,0,0) is not one of the closest two tetrahedral vertices.
			c = aPoint | bPoint // Our two extra vertices are determined by the closest two.

			if c&0x01 == 0 {
				xsv_ext0 = xsb
				xsv_ext1 = xsb - 1
				dx_ext0 = dx0 - 2*squishConstant3D
				dx_ext1 = dx0 + 1 - squishConstant3D
			} else {
				xsv_ext0 = xsb + 1
				xsv_ext1 = xsv_ext0
				dx_ext0 = dx0 - 1 - 2*squishConstant3D
				dx_ext1 = dx0 - 1 - squishConstant3D
			}

			if c&0x02 == 0 {
				ysv_ext0 = ysb
				ysv_ext1 = ysb - 1
				dy_ext0 = dy0 - 2*squishConstant3D
				dy_ext1 = dy0 + 1 - squishConstant3D
			} else {
				ysv_ext0 = ysb + 1
				ysv_ext1 = ysv_ext0
				dy_ext0 = dy0 - 1 - 2*squishConstant3D
				dy_ext1 = dy0 - 1 - squishConstant3D
			}

			if c&0x04 == 0 {
				zsv_ext0 = zsb
				zsv_ext1 = zsb - 1
				dz_ext0 = dz0 - 2*squishConstant3D
				dz_ext1 = dz0 + 1 - squishConstant3D
			} else {
				zsv_ext0 = zsb + 1
				zsv_ext1 = zsv_ext0
				dz_ext0 = dz0 - 1 - 2*squishConstant3D
				dz_ext1 = dz0 - 1 - squishConstant3D
			}
		}

		// Contribution (0,0,0)
		var attn0 float64 = 2 - dx0*dx0 - dy0*dy0 - dz0*dz0
		if attn0 > 0 {
			attn0 *= attn0
			value += attn0 * attn0 * osg.extrapolate3(xsb+0, ysb+0, zsb+0, dx0, dy0, dz0)
		}

		// Contribution (1,0,0)
		var dx1 float64 = dx0 - 1 - squishConstant3D
		var dy1 float64 = dy0 - 0 - squishConstant3D
		var dz1 float64 = dz0 - 0 - squishConstant3D
		var attn1 float64 = 2 - dx1*dx1 - dy1*dy1 - dz1*dz1
		if attn1 > 0 {
			attn1 *= attn1
			value += attn1 * attn1 * osg.extrapolate3(xsb+1, ysb+0, zsb+0, dx1, dy1, dz1)
		}

		// Contribution (0,1,0)
		var dx2 float64 = dx0 - 0 - squishConstant3D
		var dy2 float64 = dy0 - 1 - squishConstant3D
		var dz2 float64 = dz1
		var attn2 float64 = 2 - dx2*dx2 - dy2*dy2 - dz2*dz2
		if attn2 > 0 {
			attn2 *= attn2
			value += attn2 * attn2 * osg.extrapolate3(xsb+0, ysb+1, zsb+0, dx2, dy2, dz2)
		}

		// Contribution (0,0,1)
		var dx3 float64 = dx2
		var dy3 float64 = dy1
		var dz3 float64 = dz0 - 1 - squishConstant3D
		var attn3 float64 = 2 - dx3*dx3 - dy3*dy3 - dz3*dz3
		if attn3 > 0 {
			attn3 *= attn3
			value += attn3 * attn3 * osg.extrapolate3(xsb+0, ysb+0, zsb+1, dx3, dy3, dz3)
		}
	} else if inSum >= 2 { // We're inside the tetrahedron (3-Simplex) at (1,1,1)
		// Determine which two tetrahedral vertices are the closest, out of (1,1,0), (1,0,1), (0,1,1) but not (1,1,1).
		var aPoint byte = 0x06
		var aScore float64 = xins
		var bPoint byte = 0x05
		var bScore float64 = yins
		if aScore <= bScore && zins < bScore {
			bScore = zins
			bPoint = 0x03
		} else if aScore > bScore && zins < aScore {
			aScore = zins
			aPoint = 0x03
		}

		// Now we determine the two lattice points not part of the tetrahedron that may contribute.
		// This depends on the closest two tetrahedral vertices, including (1,1,1)
		var c byte
		var wins float64 = 3.0 - inSum
		if wins < aScore || wins < bScore { // (1,1,1) is one of the closest two tetrahedral vertices.
			// Our other closest vertex is the closest out of a and b.
			if bScore < aScore {
				c = bPoint
			} else {
				c = aPoint
			}

			if c&0x01 != 0 {
				xsv_ext0 = xsb + 2
				xsv_ext1 = xsb + 1
				dx_ext0 = dx0 - 2 - 3*squishConstant3D
				dx_ext1 = dx0 - 1 - 3*squishConstant3D
			} else {
				xsv_ext0 = xsb
				xsv_ext1 = xsb
				dx_ext0 = dx0 - 3*squishConstant3D
				dx_ext1 = dx_ext0
			}

			if c&0x02 != 0 {
				ysv_ext0 = ysb + 1
				ysv_ext1 = ysv_ext0
				dy_ext0 = dy0 - 1 - 3*squishConstant3D
				dy_ext1 = dy_ext0
				if c&0x01 != 0 {
					ysv_ext1 += 1
					dy_ext1 -= 1
				} else {
					ysv_ext0 += 1
					dy_ext0 -= 1
				}
			} else {
				ysv_ext0 = ysb
				ysv_ext1 = ysb
				dy_ext0 = dy0 - 3*squishConstant3D
				dy_ext1 = dy_ext0
			}

			if c&0x04 != 0 {
				zsv_ext0 = zsb + 1
				zsv_ext1 = zsb + 2
				dz_ext0 = dz0 - 1 - 3*squishConstant3D
				dz_ext1 = dz0 - 2 - 3*squishConstant3D
			} else {
				zsv_ext0 = zsb
				zsv_ext1 = zsb
				dz_ext0 = dz0 - 3*squishConstant3D
				dz_ext1 = dz_ext0
			}
		} else { // (1,1,1) is not one of the closest two tetrahedral vertices.
			c = aPoint & bPoint // Our two extra vertices are determined by the closest two.

			if c&0x01 != 0 {
				xsv_ext0 = xsb + 1
				xsv_ext1 = xsb + 2
				dx_ext0 = dx0 - 1 - squishConstant3D
				dx_ext1 = dx0 - 2 - 2*squishConstant3D
			} else {
				xsv_ext0 = xsb
				xsv_ext1 = xsb
				dx_ext0 = dx0 - squishConstant3D
				dx_ext1 = dx0 - 2*squishConstant3D
			}

			if c&0x02 != 0 {
				ysv_ext0 = ysb + 1
				ysv_ext1 = ysb + 2
				dy_ext0 = dy0 - 1 - squishConstant3D
				dy_ext1 = dy0 - 2 - 2*squishConstant3D
			} else {
				ysv_ext0 = ysb
				ysv_ext1 = ysb
				dy_ext0 = dy0 - squishConstant3D
				dy_ext1 = dy0 - 2*squishConstant3D
			}

			if c&0x04 != 0 {
				zsv_ext0 = zsb + 1
				zsv_ext1 = zsb + 2
				dz_ext0 = dz0 - 1 - squishConstant3D
				dz_ext1 = dz0 - 2 - 2*squishConstant3D
			} else {
				zsv_ext0 = zsb
				zsv_ext1 = zsb
				dz_ext0 = dz0 - squishConstant3D
				dz_ext1 = dz0 - 2*squishConstant3D
			}
		}

		// Contribution (1,1,0)
		var dx3 float64 = dx0 - 1 - 2*squishConstant3D
		var dy3 float64 = dy0 - 1 - 2*squishConstant3D
		var dz3 float64 = dz0 - 0 - 2*squishConstant3D
		var attn3 float64 = 2 - dx3*dx3 - dy3*dy3 - dz3*dz3
		if attn3 > 0 {
			attn3 *= attn3
			value += attn3 * attn3 * osg.extrapolate3(xsb+1, ysb+1, zsb+0, dx3, dy3, dz3)
		}

		// Contribution (1,0,1)
		var dx2 float64 = dx3
		var dy2 float64 = dy0 - 0 - 2*squishConstant3D
		var dz2 float64 = dz0 - 1 - 2*squishConstant3D
		var attn2 float64 = 2 - dx2*dx2 - dy2*dy2 - dz2*dz2
		if attn2 > 0 {
			attn2 *= attn2
			value += attn2 * attn2 * osg.extrapolate3(xsb+1, ysb+0, zsb+1, dx2, dy2, dz2)
		}

		//Contribution (0,1,1)
		var dx1 float64 = dx0 - 0 - 2*squishConstant3D
		var dy1 float64 = dy3
		var dz1 float64 = dz2
		var attn1 float64 = 2 - dx1*dx1 - dy1*dy1 - dz1*dz1
		if attn1 > 0 {
			attn1 *= attn1
			value += attn1 * attn1 * osg.extrapolate3(xsb+0, ysb+1, zsb+1, dx1, dy1, dz1)
		}

		//Contribution (1,1,1)
		dx0 = dx0 - 1 - 3*squishConstant3D
		dy0 = dy0 - 1 - 3*squishConstant3D
		dz0 = dz0 - 1 - 3*squishConstant3D
		var attn0 float64 = 2 - dx0*dx0 - dy0*dy0 - dz0*dz0
		if attn0 > 0 {
			attn0 *= attn0
			value += attn0 * attn0 * osg.extrapolate3(xsb+1, ysb+1, zsb+1, dx0, dy0, dz0)
		}
	} else { // We're inside the octahedron (Rectified 3-Simplex) in between.
		var aScore float64
		var aPoint byte
		var aIsFurtherSide bool
		var bPoint byte
		var bScore float64
		var bIsFurtherSide bool

		// Decide between point (0,0,1) and (1,1,0) as closest
		var p1 float64 = xins + yins
		if p1 > 1.0 {
			aScore = p1 - 1
			aPoint = 0x03
			aIsFurtherSide = true
		} else {
			aScore = 1 - p1
			aPoint = 0x04
			aIsFurtherSide = false
		}

		// Decide between point (0,1,0) and (1,0,1) as closest
		var p2 float64 = xins + zins
		if p2 > 1.0 {
			bScore = p2 - 1
			bPoint = 0x05
			bIsFurtherSide = true
		} else {
			bScore = 1 - p2
			bPoint = 0x02
			bIsFurtherSide = false
		}

		// The closest out of the two (1,0,0) and (0,1,1) will replace the furthest out of the two decided above, if closer.
		var p3 float64 = yins + zins
		if p3 > 1.0 {
			var score float64 = p3 - 1
			if aScore <= bScore && aScore < score {
				aScore = score
				aPoint = 0x06
				aIsFurtherSide = true
			} else if aScore > bScore && bScore < score {
				bScore = score
				bPoint = 0x06
				bIsFurtherSide = true
			}
		} else {
			var score float64 = 1 - p3
			if aScore <= bScore && aScore < score {
				aScore = score
				aPoint = 0x01
				aIsFurtherSide = false
			} else if aScore > bScore && bScore < score {
				bScore = score
				bPoint = 0x01
				bIsFurtherSide = false
			}
		}

		// Where each of the two closest points are determines how the extra two vertices are calculated.
		if aIsFurtherSide == bIsFurtherSide {
			if aIsFurtherSide { // Both closest points on (1,1,1) side
				// One of the two extra points is (1,1,1)
				dx_ext0 = dx0 - 1 - 3*squishConstant3D
				dy_ext0 = dy0 - 1 - 3*squishConstant3D
				dz_ext0 = dz0 - 1 - 3*squishConstant3D
				xsv_ext0 = xsb + 1
				ysv_ext0 = ysb + 1
				zsv_ext0 = zsb + 1

				// Other extra point is based on the shared axis.
				var c byte = aPoint & bPoint
				if c&0x01 != 0 {
					dx_ext1 = dx0 - 2 - 2*squishConstant3D
					dy_ext1 = dy0 - 2*squishConstant3D
					dz_ext1 = dz0 - 2*squishConstant3D
					xsv_ext1 = xsb + 2
					ysv_ext1 = ysb
					zsv_ext1 = zsb
				} else if c&0x02 != 0 {
					dx_ext1 = dx0 - 2*squishConstant3D
					dy_ext1 = dy0 - 2 - 2*squishConstant3D
					dz_ext1 = dz0 - 2*squishConstant3D
					xsv_ext1 = xsb
					ysv_ext1 = ysb + 2
					zsv_ext1 = zsb
				} else {
					dx_ext1 = dx0 - 2*squishConstant3D
					dy_ext1 = dy0 - 2*squishConstant3D
					dz_ext1 = dz0 - 2 - 2*squishConstant3D
					xsv_ext1 = xsb
					ysv_ext1 = ysb
					zsv_ext1 = zsb + 2
				}
			} else { // Both closest points on (0,0,0) side
				// one of the two extra points is (0,0,0)
				dx_ext0 = dx0
				dy_ext0 = dy0
				dz_ext0 = dz0
				xsv_ext0 = xsb
				ysv_ext0 = ysb
				zsv_ext0 = zsb

				// Other extra point is based on the omitted axis.
				var c byte = aPoint | bPoint
				if c&0x01 == 0 {
					dx_ext1 = dx0 + 1 - squishConstant3D
					dy_ext1 = dy0 - 1 - squishConstant3D
					dz_ext1 = dz0 - 1 - squishConstant3D
					xsv_ext1 = xsb - 1
					ysv_ext1 = ysb + 1
					zsv_ext1 = zsb + 1
				} else if c&0x02 == 0 {
					dx_ext1 = dx0 - 1 - squishConstant3D
					dy_ext1 = dy0 + 1 - squishConstant3D
					dz_ext1 = dz0 - 1 - squishConstant3D
					xsv_ext1 = xsb + 1
					ysv_ext1 = ysb - 1
					zsv_ext1 = zsb + 1
				} else {
					dx_ext1 = dx0 - 1 - squishConstant3D
					dy_ext1 = dy0 - 1 - squishConstant3D
					dz_ext1 = dz0 + 1 - squishConstant3D
					xsv_ext1 = xsb + 1
					ysv_ext1 = ysb + 1
					zsv_ext1 = zsb - 1
				}
			}
		} else { // One point on (0,0,0) side, one point on (1,1,1) side
			var c1, c2 byte
			if aIsFurtherSide {
				c1 = aPoint
				c2 = bPoint
			} else {
				c1 = bPoint
				c2 = aPoint
			}

			// One contribution is a permutation of (1,1,-1)
			if c1&0x01 == 0 {
				dx_ext0 = dx0 + 1 - squishConstant3D
				dy_ext0 = dy0 - 1 - squishConstant3D
				dz_ext0 = dz0 - 1 - squishConstant3D
				xsv_ext0 = xsb - 1
				ysv_ext0 = ysb + 1
				zsv_ext0 = zsb + 1
			} else if c1&0x02 == 0 {
				dx_ext0 = dx0 - 1 - squishConstant3D
				dy_ext0 = dy0 + 1 - squishConstant3D
				dz_ext0 = dz0 - 1 - squishConstant3D
				xsv_ext0 = xsb + 1
				ysv_ext0 = ysb - 1
				zsv_ext0 = zsb + 1
			} else {
				dx_ext0 = dx0 - 1 - squishConstant3D
				dy_ext0 = dy0 - 1 - squishConstant3D
				dz_ext0 = dz0 + 1 - squishConstant3D
				xsv_ext0 = xsb + 1
				ysv_ext0 = ysb + 1
				zsv_ext0 = zsb - 1
			}

			// One contribution is a permutation of (0,0,2)
			dx_ext1 = dx0 - 2*squishConstant3D
			dy_ext1 = dy0 - 2*squishConstant3D
			dz_ext1 = dz0 - 2*squishConstant3D
			xsv_ext1 = xsb
			ysv_ext1 = ysb
			zsv_ext1 = zsb
			if c2&0x01 != 0 {
				dx_ext1 -= 2
				xsv_ext1 += 2
			} else if c2&0x02 != 0 {
				dy_ext1 -= 2
				ysv_ext1 += 2
			} else {
				dz_ext1 -= 2
				zsv_ext1 += 2
			}
		}

		// Contribution (1,0,0)
		var dx1 float64 = dx0 - 1 - squishConstant3D
		var dy1 float64 = dy0 - 0 - squishConstant3D
		var dz1 float64 = dz0 - 0 - squishConstant3D
		var attn1 float64 = 2 - dx1*dx1 - dy1*dy1 - dz1*dz1
		if attn1 > 0 {
			attn1 *= attn1
			value += attn1 * attn1 * osg.extrapolate3(xsb+1, ysb+0, zsb+0, dx1, dy1, dz1)
		}

		// Contribution (0,1,0)
		var dx2 float64 = dx0 - 0 - squishConstant3D
		var dy2 float64 = dy0 - 1 - squishConstant3D
		var dz2 float64 = dz1
		var attn2 float64 = 2 - dx2*dx2 - dy2*dy2 - dz2*dz2
		if attn2 > 0 {
			attn2 *= attn2
			value += attn2 * attn2 * osg.extrapolate3(xsb+0, ysb+1, zsb+0, dx2, dy2, dz2)
		}

		// Contribution (0,0,1)
		var dx3 float64 = dx2
		var dy3 float64 = dy1
		var dz3 float64 = dz0 - 1 - squishConstant3D
		var attn3 float64 = 2 - dx3*dx3 - dy3*dy3 - dz3*dz3
		if attn3 > 0 {
			attn3 *= attn3
			value += attn3 * attn3 * osg.extrapolate3(xsb+0, ysb+0, zsb+1, dx3, dy3, dz3)
		}

		// Contribution (1,1,0)
		var dx4 float64 = dx0 - 1 - 2*squishConstant3D
		var dy4 float64 = dy0 - 1 - 2*squishConstant3D
		var dz4 float64 = dz0 - 0 - 2*squishConstant3D
		var attn4 float64 = 2 - dx4*dx4 - dy4*dy4 - dz4*dz4
		if attn4 > 0 {
			attn4 *= attn4
			value += attn4 * attn4 * osg.extrapolate3(xsb+1, ysb+1, zsb+0, dx4, dy4, dz4)
		}

		// Contribution (1,0,1)
		var dx5 float64 = dx4
		var dy5 float64 = dy0 - 0 - 2*squishConstant3D
		var dz5 float64 = dz0 - 1 - 2*squishConstant3D
		var attn5 float64 = 2 - dx5*dx5 - dy5*dy5 - dz5*dz5
		if attn5 > 0 {
			attn5 *= attn5
			value += attn5 * attn5 * osg.extrapolate3(xsb+1, ysb+0, zsb+1, dx5, dy5, dz5)
		}

		// Contribution (0,1,1)
		var dx6 float64 = dx0 - 0 - 2*squishConstant3D
		var dy6 float64 = dy4
		var dz6 float64 = dz5
		var attn6 float64 = 2 - dx6*dx6 - dy6*dy6 - dz6*dz6
		if attn6 > 0 {
			attn6 *= attn6
			value += attn6 * attn6 * osg.extrapolate3(xsb+0, ysb+1, zsb+1, dx6, dy6, dz6)
		}
	}

	// First extra vertex
	var attn_ext0 float64 = 2 - dx_ext0*dx_ext0 - dy_ext0*dy_ext0 - dz_ext0*dz_ext0
	if attn_ext0 > 0 {
		attn_ext0 *= attn_ext0
		value += attn_ext0 * attn_ext0 * osg.extrapolate3(xsv_ext0, ysv_ext0, zsv_ext0, dx_ext0, dy_ext0, dz_ext0)
	}

	// Second extra vertex
	var attn_ext1 float64 = 2 - dx_ext1*dx_ext1 - dy_ext1*dy_ext1 - dz_ext1*dz_ext1
	if attn_ext1 > 0 {
		attn_ext1 *= attn_ext1
		value += attn_ext1 * attn_ext1 * osg.extrapolate3(xsv_ext1, ysv_ext1, zsv_ext1, dx_ext1, dy_ext1, dz_ext1)
	}

	return value / normConstant3D
}
