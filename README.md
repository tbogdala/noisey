noisey: a go library for coherent random noise
==============================================

This library natively implements coherent noise algorithms in Go. No 3rd party libraries are required.

Currently it supports the following:

* 2D (64-bit) [perlin noise][link1]


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

Full examples can be found in the `examples` folder, but this fragment will illustrate basic usage of perlin noise:

```go
import "github.com/tbogdala/noisey"

// ... yadda yadda ...

// create a new RNG from Go's built in library with a seed of '1'
r := rand.New(rand.NewSource(int64(1)))

// create a new perlin noise generator using the RNG created above
perlin := noisey.NewPerlinGenerator2D(r, 256)

// get the noise value at point (0.4, 0.2) and use 'fast' smoothing
v := perlin.Get(0.4, 0.2, noisey.StandardQuality)
```


To Do
-----

* fractal Brownian motion
* 3D perlin noise
* a way to combine noise generators and modifiers to support
things like libnoise's [complex planetary surface][link2]


License
-------

Noisey is released under the BSD license. See the `LICENSE` file for more details.



[link1]: http://webstaff.itn.liu.se/~stegu/TNM022-2005/perlinnoiselinks/perlin-noise-math-faq.html
[link2]: http://libnoise.sourceforge.net/examples/complexplanet/index.html
