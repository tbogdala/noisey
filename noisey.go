/*
Package noisey is a library that implements coherent noise algorithms.

The selection is currently very limited and consists of: 2D Perlin noise (64bit).

An interface called 'RandomSource' is also exported so that a client can implement
a different random number generator and pass it to the noise generators.

Sample programs can be found in the 'examples' directory.

Here's a quick fragment of how to use the 2D perlin noise generator:

  import "github.com/tbogdala/noisey"

  // ... yadda yadda ...

  // create a new RNG from Go's built in library with a seed of '1'
  r := rand.New(rand.NewSource(int64(1)))

  // create a new perlin noise generator using the RNG created above
  perlin := noisey.NewPerlinGenerator2D(r, 256)

  // get the noise value at point (0.4, 0.2) and use 'fast' smoothing
  v := perlin.Get(0.4, 0.2, noisey.FastQuality)

*/
package noisey

// RandomSource is a generic interface for a random number generator
// allowing the user to use the built-in RNG or a custom one that implements
// this interface.
type RandomSource interface {
  Float64() float64
  Perm(int) []int
}

// Vec2f is a simple 2D vector of 64 bit floats
type Vec2f struct {
  X, Y float64
}
