noisey: a Go library for coherent random noise
==============================================

This library natively implements coherent noise algorithms in Go. No 3rd party libraries are required.

Currently it supports the following:

### Sources

* 2D (64-bit) [Perlin noise][link1]
* 2D (64-bit) [OpenSimplex noise][link3]

### Generators and Modifiers

* FBMGenerator2D - fractal Brownian Motion
* Select2D - choose from source A or B depending on control source
* Scale2D - modify output by multiplying by a scale and adding a bias constant

Additionally, noisey can load settings from a JSON configuration file and create
sources and generators from that.

**IMPORTANT: This is a new library and API stability is not guaranteed (though 1.0 is getting close).**


Installation
------------

You can get the latest copy of the library using this command:

```bash
go get github.com/tbogdala/noisey
```

You can then use it in your code by importing it:

```go
import "github.com/tbogdala/noisey"
```


Usage
-----

Full examples can be found in the `examples` folder, but this fragment will illustrate basic usage of Perlin noise:

```go
import "github.com/tbogdala/noisey"

// ... yadda yadda yadda ...

// create a new RNG from Go's built in library with a seed of '1'
r := rand.New(rand.NewSource(int64(1)))

// create a new perlin noise generator using the RNG created above
perlin := noisey.NewPerlinGenerator2D(r, noisey.StandardQuality)

// get the noise value at point (0.4, 0.2)
v := perlin.GetValue2D(0.4, 0.2)
```

If you want "smooth" noise, or fractal Brownian motion, then you use
the noise generator with another structure:

```go
import "github.com/tbogdala/noisey"

// ... yadda yadda yadda ...

// create a new RNG from Go's built in library with a seed of '1'
r := rand.New(rand.NewSource(int64(1)))

// create a new Perlin noise generator using the RNG created above
noiseGen := noisey.NewPerlinGenerator2D(r, noisey.HighQuality)

// create the fractal Brownian motion modifier based on Perlin
fbmPerlin := noisey.NewFBMGenerator2D(&noiseGen)
fbmPerlin.Octaves = 5
fbmPerlin.Persistence = 0.25
fbmPerlin.Lacunarity = 2.0
fbmPerlin.Frequency = 1.13

// get the noise value at point (0.4, 0.2)
v := fbmPerlin.Get2D(0.4, 0.2)
```

Samples that display noise to console or OpenGL windows are included. If the
OpenGL examples are desired, you must also install the Go libraries
`github.com/go-gl/gl` and `github.com/go-gl/glfw3`

Other noise generators like OpenSimplex can be used in similar manner. Just
create the noise generator with the constructor by passing a random number generator:

```go
// create a new RNG from Go's built in library with a seed of '1'
r := rand.New(rand.NewSource(int64(1)))

// create a new OpenSimplex noise generator using the RNG created above
opensimplex := noisey.NewOpenSimplexGenerator2D(r)
```

Benchmarks
----------

Benchmarks can be run with Go's built-in test tool by executing the following:

```bash
cd $GOPATH/src/github.com/tbogdala/noisey
go test -cpu 4 -bench .
```

The cpu flag can be adjusted accordingly, but shouldn't make a difference since
the generators don't improve in parallel operation.

To Do
-----

* 3D noise algorithms


License
-------

Noisey is released under the BSD license. See the `LICENSE` file for more details.



[link1]: http://webstaff.itn.liu.se/~stegu/TNM022-2005/perlinnoiselinks/perlin-noise-math-faq.html
[link2]: http://libnoise.sourceforge.net/examples/complexplanet/index.html
[link3]: http://uniblock.tumblr.com/post/97868843242/noise
