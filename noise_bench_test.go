package noisey

/* Copyright 2014, Timothy Bogdala <tdb@animal-machine.com>
See the LICENSE file for more details. */

import (
	"math/rand"
	"testing"
)

func BenchmarkPerlin2D(b *testing.B) {
	var sum float64 = 0
	const benchSize = 100
	const totalBenchSize = benchSize * benchSize

	// make a test generator seeded to 1
	rngPerlin := rand.New(rand.NewSource(int64(1)))
	perlin := NewPerlinGenerator(rngPerlin)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for y := 0; y < benchSize; y++ {
			for x := 0; x < benchSize; x++ {
				sum += perlin.Get2D(float64(x), float64(y))
			}
		}
	}
	//fmt.Printf("\n\nPerlin resulting sum = %f\n", sum)
}

func BenchmarkPerlin3Das2D(b *testing.B) {
	var sum float64 = 0
	const benchSize = 100
	const totalBenchSize = benchSize * benchSize

	// make a test generator seeded to 1
	rngPerlin := rand.New(rand.NewSource(int64(1)))
	perlin := NewPerlinGenerator(rngPerlin)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for y := 0; y < benchSize; y++ {
			for x := 0; x < benchSize; x++ {
				sum += perlin.Get3D(float64(x), float64(y), 0.0)
			}
		}
	}
	//fmt.Printf("\n\nPerlin resulting sum = %f\n", sum)
}

func BenchmarkPerlin3D(b *testing.B) {
	var sum float64 = 0
	const benchSize = 100
	const totalBenchSize = benchSize * benchSize

	// make a test generator seeded to 1
	rngPerlin := rand.New(rand.NewSource(int64(1)))
	perlin := NewPerlinGenerator(rngPerlin)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for y := 0; y < benchSize; y++ {
			for x := 0; x < benchSize; x++ {
				for z := 0; x < benchSize; x++ {
					sum += perlin.Get3D(float64(x), float64(y), float64(z))
				}
			}
		}
	}
	//fmt.Printf("\n\nPerlin resulting sum = %f\n", sum)
}

func BenchmarkOpenSimplex2D(b *testing.B) {
	var sum float64 = 0
	const benchSize = 100
	const totalBenchSize = benchSize * benchSize

	// make a test generator seeded to 1
	rngOpenSimplex := rand.New(rand.NewSource(int64(1)))
	openSimplex := NewOpenSimplexGenerator(rngOpenSimplex)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for y := 0; y < benchSize; y++ {
			for x := 0; x < benchSize; x++ {
				sum += openSimplex.Get2D(float64(x), float64(y))
			}
		}
	}
	//	fmt.Printf("\n\nOpenSimplex resulting sum = %f\n", sum)
}

func BenchmarkOpenSimplex3Das2D(b *testing.B) {
	var sum float64 = 0
	const benchSize = 100
	const totalBenchSize = benchSize * benchSize

	// make a test generator seeded to 1
	rngOpenSimplex := rand.New(rand.NewSource(int64(1)))
	openSimplex := NewOpenSimplexGenerator(rngOpenSimplex)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for y := 0; y < benchSize; y++ {
			for x := 0; x < benchSize; x++ {
				sum += openSimplex.Get3D(float64(x), float64(y), 0.0)
			}
		}
	}
	//	fmt.Printf("\n\nOpenSimplex resulting sum = %f\n", sum)
}

func BenchmarkOpenSimplex3D(b *testing.B) {
	var sum float64 = 0
	const benchSize = 100
	const totalBenchSize = benchSize * benchSize

	// make a test generator seeded to 1
	rngOpenSimplex := rand.New(rand.NewSource(int64(1)))
	openSimplex := NewOpenSimplexGenerator(rngOpenSimplex)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for y := 0; y < benchSize; y++ {
			for x := 0; x < benchSize; x++ {
				for z := 0; x < benchSize; x++ {
					sum += openSimplex.Get3D(float64(x), float64(y), float64(z))
				}
			}
		}
	}
	//	fmt.Printf("\n\nOpenSimplex resulting sum = %f\n", sum)
}
