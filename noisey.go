/*
Package noisey is a library that implements coherent noise algorithms.

The selection is currently very limited and consists of:

	* 2D Perlin noise (64bit)
	* 2D OpenSimplex noise (64bit)

An interface called 'RandomSource' is also exported so that a client can implement
a different random number generator and pass it to the noise generators.

Sample programs can be found in the 'examples' directory.

*/
package noisey

// RandomSource is a generic interface for a random number generator
// allowing the user to use the built-in RNG or a custom one that implements
// this interface.
type RandomSource interface {
	Float64() float64
	Perm(int) []int
}

// BuilderSource2D is an interface defining how the Builder* types get noise.
type BuilderSource2D interface {
	Get2D(float64, float64) float64
}

// CoherentRandomGen2D is a generic interface for a noise generator
// that makes coherent random noise.
type CoherentRandomGen2D interface {
	GetValue2D(float64, float64) float64
}

// Vec2f is a simple 2D vector of 64 bit floats
type Vec2f struct {
	X, Y float64
}
