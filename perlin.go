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
  * https://gist.github.com/nsf/1170424#file-test-go

This implementation is basically directly ported from noise-rs rust code:
https://github.com/bjz/noise-rs/blob/master/src/perlin.rs

*/

import (
	"math"
)

const (
	tableSize = 256
)

// PerlinGenerator stores the state information for generating perlin noise.
type PerlinGenerator struct {
	Rng             RandomSource // random number generator interface
	Permutations    []int        // the random permutation table
	RandomGradients []Vec4f      // the random gradient table
}

// NewPerlinGenerator creates a new state object for the #D perlin noise generator
func NewPerlinGenerator(rng RandomSource) (pg PerlinGenerator) {
	pg.Rng = rng
	pg.Permutations = rng.Perm(tableSize)

	pg.RandomGradients = make([]Vec4f, 32)
	pg.RandomGradients[1] = Vec4f{0.0, 1.0, 1.0, -1.0}    //  [ zero,  one,   one,  -one],
	pg.RandomGradients[2] = Vec4f{0.0, 1.0, -1.0, 1.0}    // [ zero,  one,  -one,   one],
	pg.RandomGradients[3] = Vec4f{0.0, 1.0, -1.0, -1.0}   // [ zero,  one,  -one,  -one],
	pg.RandomGradients[4] = Vec4f{0.0, -1.0, 1.0, 1.0}    // [ zero, -one,   one,   one],
	pg.RandomGradients[5] = Vec4f{0.0, -1.0, 1.0, -1.0}   // [ zero, -one,   one,  -one],
	pg.RandomGradients[6] = Vec4f{0.0, -1.0, -1.0, 1.0}   // [ zero, -one,  -one,   one],
	pg.RandomGradients[7] = Vec4f{0.0, -1.0, -1.0, -1.0}  // [ zero, -one,  -one,  -one],
	pg.RandomGradients[8] = Vec4f{1.0, 0.0, 1.0, 1.0}     // [ one,   zero,  one,   one],
	pg.RandomGradients[9] = Vec4f{1.0, 0.0, 1.0, -1.0}    // [ one,   zero,  one,  -one],
	pg.RandomGradients[10] = Vec4f{1.0, 0.0, -1.0, 1.0}   // [ one,   zero, -one,   one],
	pg.RandomGradients[11] = Vec4f{1.0, 0.0, -1.0, -1.0}  // [ one,   zero, -one,  -one],
	pg.RandomGradients[12] = Vec4f{-1.0, 0.0, 1.0, 1.0}   // [-one,   zero,  one,   one],
	pg.RandomGradients[13] = Vec4f{-1.0, 0.0, 1.0, -1.0}  // [-one,   zero,  one,  -one],
	pg.RandomGradients[14] = Vec4f{-1.0, 0.0, -1.0, 1.0}  // [-one,   zero, -one,   one],
	pg.RandomGradients[15] = Vec4f{-1.0, 0.0, -1.0, -1.0} // [-one,   zero, -one,  -one],
	pg.RandomGradients[16] = Vec4f{1.0, 1.0, 0.0, 1.0}    // [ one,   one,   zero,  one],
	pg.RandomGradients[17] = Vec4f{1.0, 1.0, 0.0, -1.0}   // [ one,   one,   zero, -one],
	pg.RandomGradients[18] = Vec4f{1.0, -1.0, 0.0, 1.0}   // [ one,  -one,   zero,  one],
	pg.RandomGradients[19] = Vec4f{0.0, -1.0, 0.0, -1.0}  // [ one,  -one,   zero, -one],
	pg.RandomGradients[20] = Vec4f{-1.0, 1.0, 0.0, 1.0}   // [-one,   one,   zero,  one],
	pg.RandomGradients[21] = Vec4f{-1.0, 1.0, 0.0, -1.0}  // [-one,   one,   zero, -one],
	pg.RandomGradients[22] = Vec4f{-1.0, -1.0, 0.0, 1.0}  // [-one,  -one,   zero,  one],
	pg.RandomGradients[23] = Vec4f{-1.0, -1.0, 0.0, -1.0} // [-one,  -one,   zero, -one],
	pg.RandomGradients[24] = Vec4f{1.0, 1.0, 1.0, 0.0}    // [ one,   one,   one,   zero],
	pg.RandomGradients[25] = Vec4f{1.0, 1.0, -1.0, 0.0}   // [ one,   one,  -one,   zero],
	pg.RandomGradients[26] = Vec4f{1.0, -1.0, 1.0, 0.0}   // [ one,  -one,   one,   zero],
	pg.RandomGradients[27] = Vec4f{1.0, -1.0, -1.0, 0.0}  // [ one,  -one,  -one,   zero],
	pg.RandomGradients[28] = Vec4f{-1.0, 1.0, 1.0, 0.0}   // [-one,   one,   one,   zero],
	pg.RandomGradients[29] = Vec4f{-1.0, 1.0, -1.0, 0.0}  // [-one,   one,  -one,   zero],
	pg.RandomGradients[30] = Vec4f{-1.0, -1.0, 1.0, 0.0}  // [-one,  -one,   one,   zero],
	pg.RandomGradients[31] = Vec4f{-1.0, -1.0, -1.0, 0.0} // [-one,  -one,  -one,   zero],

	return
}

func (pg *PerlinGenerator) getGradient2(whole Vec2i) Vec2f {
	x := whole.X & 0xFF
	xv := pg.Permutations[x]

	y := whole.Y & 0xFF
	yv := pg.Permutations[xv^y]

	i := yv % 32
	return Vec2f{pg.RandomGradients[i].X, pg.RandomGradients[i].Y}
}

func (pg *PerlinGenerator) getGradient3(whole Vec3i) Vec3f {
	x := whole.X & 0xFF
	xv := pg.Permutations[x]

	y := whole.Y & 0xFF
	yv := pg.Permutations[xv^y]

	z := whole.Z & 0xFF
	zv := pg.Permutations[yv^z]

	i := zv % 32
	return Vec3f{pg.RandomGradients[i].X, pg.RandomGradients[i].Y, pg.RandomGradients[i].Z}
}

func vec3fDot(a, b Vec3f) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

func vec2fDot(a, b Vec2f) float64 {
	return a.X*b.X + a.Y*b.Y
}

// Get3D calculates the perlin noise at a given 3D coordinate
func (pg *PerlinGenerator) Get3D(x, y, z float64) float64 {
	gradient3 := func(whole Vec3i, frac Vec3f) float64 {
		attn := 1.0 - vec3fDot(frac, frac)
		if attn > 0.0 {
			return (attn * attn) * vec3fDot(frac, pg.getGradient3(whole))
		} else {
			return 0.0
		}
	}

	floored := Vec3f{math.Floor(x), math.Floor(y), math.Floor(z)}
	whole0 := Vec3i{int(floored.X), int(floored.Y), int(floored.Z)}
	whole1 := Vec3i{whole0.X + 1, whole0.Y + 1, whole0.Z + 1}
	frac0 := Vec3f{x - floored.X, y - floored.Y, z - floored.Z}
	frac1 := Vec3f{frac0.X - 1, frac0.Y - 1, frac0.Z - 1}

	f000 := gradient3(Vec3i{whole0.X, whole0.Y, whole0.Z}, Vec3f{frac0.X, frac0.Y, frac0.Z})
	f100 := gradient3(Vec3i{whole1.X, whole0.Y, whole0.Z}, Vec3f{frac1.X, frac0.Y, frac0.Z})
	f010 := gradient3(Vec3i{whole0.X, whole1.Y, whole0.Z}, Vec3f{frac0.X, frac1.Y, frac0.Z})
	f110 := gradient3(Vec3i{whole1.X, whole1.Y, whole0.Z}, Vec3f{frac1.X, frac1.Y, frac0.Z})
	f001 := gradient3(Vec3i{whole0.X, whole0.Y, whole1.Z}, Vec3f{frac0.X, frac0.Y, frac1.Z})
	f101 := gradient3(Vec3i{whole1.X, whole0.Y, whole1.Z}, Vec3f{frac1.X, frac0.Y, frac1.Z})
	f011 := gradient3(Vec3i{whole0.X, whole1.Y, whole1.Z}, Vec3f{frac0.X, frac1.Y, frac1.Z})
	f111 := gradient3(Vec3i{whole1.X, whole1.Y, whole1.Z}, Vec3f{frac1.X, frac1.Y, frac1.Z})

	// Arbitrary values to shift and scale noise to -1..1
	return (f000 + f100 + f010 + f110 + f001 + f101 + f011 + f111 + 0.053179) * 1.056165
}

// Get2D calculates the perlin noise at a given 2D coordinate
func (pg *PerlinGenerator) Get2D(x, y float64) float64 {
	gradient2 := func(whole Vec2i, frac Vec2f) float64 {
		attn := 1.0 - vec2fDot(frac, frac)
		if attn > 0.0 {
			return (attn * attn) * vec2fDot(frac, pg.getGradient2(whole))
		} else {
			return 0.0
		}
	}

	floored := Vec2f{math.Floor(x), math.Floor(y)}
	whole0 := Vec2i{int(floored.X), int(floored.Y)}
	whole1 := Vec2i{whole0.X + 1, whole0.Y + 1}
	frac0 := Vec2f{x - floored.X, y - floored.Y}
	frac1 := Vec2f{frac0.X - 1, frac0.Y - 1}

	f00 := gradient2(Vec2i{whole0.X, whole0.Y}, Vec2f{frac0.X, frac0.Y})
	f10 := gradient2(Vec2i{whole1.X, whole0.Y}, Vec2f{frac1.X, frac0.Y})
	f01 := gradient2(Vec2i{whole0.X, whole1.Y}, Vec2f{frac0.X, frac1.Y})
	f11 := gradient2(Vec2i{whole1.X, whole1.Y}, Vec2f{frac1.X, frac1.Y})

	// Arbitrary values to shift and scale noise to -1..1
	return (f00 + f10 + f01 + f11 + 0.053179) * 1.056165
}
