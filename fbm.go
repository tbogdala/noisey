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

// FBMGenerator2D takes noise and makes fractal Brownian motion values.
type FBMGenerator2D struct {
	NoiseMaker  NoiseyGet2D // the interface FBMGenerator2D uses gets noise values
	Octaves     int         // the number of octaves to calculate on each Get()
	Persistence float64     // a multiplier that determines how quickly the amplitudes diminish for each successive octave
	Lacunarity  float64     // a multiplier that determines how quickly the frequency increases for each successive octave
	Frequency   float64     // the number of cycles per unit length
}

// NewFBMGenerator2D creates a new fractal Brownian motion generator state. A 'default' fBm
// would have 1 octave, 0.5 persistence, 2.0 lacunarity and 1.0 frequency.
func NewFBMGenerator2D(noise NoiseyGet2D, octaves int, persistence float64, lacunarity float64, frequency float64) (fbm FBMGenerator2D) {
	fbm.NoiseMaker = noise
	fbm.Octaves = octaves
	fbm.Persistence = persistence
	fbm.Lacunarity = lacunarity
	fbm.Frequency = frequency
	return
}

// Get2D calculates the noise value over the number of Octaves and other parameters
// that scale the coordinates over each octave.
func (fbm *FBMGenerator2D) Get2D(x float64, y float64) (v float64) {
	curPersistence := 1.0

	x *= fbm.Frequency
	y *= fbm.Frequency

	for o := 0; o < fbm.Octaves; o++ {
		signal := fbm.NoiseMaker.Get2D(x, y)
		v += signal * curPersistence

		x *= fbm.Lacunarity
		y *= fbm.Lacunarity
		curPersistence *= fbm.Persistence
	}

	return
}

// FBMGenerator3D takes noise and makes fractal Brownian motion values.
type FBMGenerator3D struct {
	NoiseMaker  NoiseyGet3D // the interface FBMGenerator3D uses gets noise values
	Octaves     int         // the number of octaves to calculate on each Get()
	Persistence float64     // a multiplier that determines how quickly the amplitudes diminish for each successive octave
	Lacunarity  float64     // a multiplier that determines how quickly the frequency increases for each successive octave
	Frequency   float64     // the number of cycles per unit length
}

// NewFBMGenerator3D creates a new fractal Brownian motion generator state. A 'default' fBm
// would have 1 octave, 0.5 persistence, 2.0 lacunarity and 1.0 frequency.
func NewFBMGenerator3D(noise NoiseyGet3D, octaves int, persistence float64, lacunarity float64, frequency float64) (fbm FBMGenerator3D) {
	fbm.NoiseMaker = noise
	fbm.Octaves = octaves
	fbm.Persistence = persistence
	fbm.Lacunarity = lacunarity
	fbm.Frequency = frequency
	return
}

// Get3D calculates the noise value over the number of Octaves and other parameters
// that scale the coordinates over each octave.
func (fbm *FBMGenerator3D) Get3D(x float64, y float64, z float64) (v float64) {
	curPersistence := 1.0

	x *= fbm.Frequency
	y *= fbm.Frequency
	z *= fbm.Frequency

	for o := 0; o < fbm.Octaves; o++ {
		signal := fbm.NoiseMaker.Get3D(x, y, z)
		v += signal * curPersistence

		x *= fbm.Lacunarity
		y *= fbm.Lacunarity
		z *= fbm.Lacunarity
		curPersistence *= fbm.Persistence
	}

	return v
}
