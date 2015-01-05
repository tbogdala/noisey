package main

/* Copyright 2014, Timothy Bogdala <tdb@animal-machine.com>
See the LICENSE file for more details. */

import (
	"fmt"
	"github.com/tbogdala/noisey"
	"math/rand"
	"runtime"
	"time"
)

func main() {
	const benchSize = 10240
	const totalBenchSize = benchSize * benchSize
	fmt.Println("Do some psuedo-benchmarking of the noise modules.")

	// make a test generator seeded to 1
	rngPerlin := rand.New(rand.NewSource(int64(1)))
	perlin := noisey.NewPerlinGenerator2D(rngPerlin, noisey.StandardQuality)

	// how fast can Perlin generate 10240x10240 numbers and add them together
	var sum float64 = 0
	perlinStart := time.Now()
	for y := 0; y < benchSize; y++ {
		for x := 0; x < benchSize; x++ {
			sum += perlin.GetValue2D(float64(x), float64(y))
		}
	}
	perlinFinish := time.Now()
	totalTime := perlinFinish.Sub(perlinStart)
	fmt.Printf("\n\nPerlin resulting sum = %f\n", sum)
	fmt.Printf("Total time for %d randoms = %s\n", totalBenchSize, totalTime)
	fmt.Printf("Time per random number = %dns\n", totalTime.Nanoseconds()/totalBenchSize)

	// clear the GC if needed -- for pseudo-science!
	runtime.GC()

	// make a test generator seeded to 1
	rngOpenSimplex := rand.New(rand.NewSource(int64(1)))
	openSimplex := noisey.NewOpenSimplexGenerator2D(rngOpenSimplex)

	// how fast can OpenSimplex generate 10240x10240 numbers and add them together
	sum = 0.0
	osgStart := time.Now()
	for y := 0; y < benchSize; y++ {
		for x := 0; x < benchSize; x++ {
			sum += openSimplex.GetValue2D(float64(x), float64(y))
		}
	}
	osgFinish := time.Now()
	totalTime = osgFinish.Sub(osgStart)
	fmt.Printf("\n\nOpenSimplex resulting sum = %f\n", sum)
	fmt.Printf("Total time for %d randoms = %s\n", totalBenchSize, totalTime)
	fmt.Printf("Time per random number = %dns\n", totalTime.Nanoseconds()/totalBenchSize)
}
