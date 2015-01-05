package noisey

/* Copyright 2014, Timothy Bogdala <tdb@animal-machine.com>
See the LICENSE file for more details. */

/*

This module performs fractal Brownian motion which combines mulitple steps
of a coherent noise generator, each with different frequency and amplitude.

Reference material:
* Overview: https://code.google.com/p/fractalterraingeneration/wiki/Fractional_Brownian_Motion
* Libnoise's glossary: http://libnoise.sourceforge.net/glossary/

*/

// CoherentRandomGen2D is a generic interface for a noise generator
// that makes coherent random noise.
type CoherentRandomGen2D interface {
	GetValue2D(float64, float64) float64
}

// FBMGenerator2D takes noise and makes fractal Brownian motion values.
type FBMGenerator2D struct {
	NoiseMaker  CoherentRandomGen2D
	Octaves     int     // the number of octaves to calculate on each Get()
	Persistence float64 // a multiplier that determines how quickly the amplitudes diminish for each successive octave
	Lacunarity  float64 // a multiplier that determines how quickly the frequency increases for each successive octave
	Frequency   float64 // the number of cycles per unit length
}

// NewFBMGenerator2D creates a new fractal Brownian motion generator state.
func NewFBMGenerator2D(noise CoherentRandomGen2D) (fbm FBMGenerator2D) {
	fbm.NoiseMaker = noise
	fbm.Octaves = 1
	fbm.Persistence = 0.5
	fbm.Lacunarity = 2.0
	fbm.Frequency = 1.0
	return
}

// Get2D calculates the noise value over the number of Octaves and other parameters
// that scale the coordinates over each octave.
func (fbm *FBMGenerator2D) Get2D(x float64, y float64) (v float64) {
	curPersistence := 1.0

	x *= fbm.Frequency
	y *= fbm.Frequency

	for o := 0; o < fbm.Octaves; o++ {
		signal := fbm.NoiseMaker.GetValue2D(x, y)
		v += signal * curPersistence

		x *= fbm.Lacunarity
		y *= fbm.Lacunarity
		curPersistence *= fbm.Persistence
	}

	return
}
