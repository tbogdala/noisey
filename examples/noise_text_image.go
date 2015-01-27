package main

/* Copyright 2014, Timothy Bogdala <tdb@animal-machine.com>
See the LICENSE file for more details. */

import (
	"bufio"
	"fmt"
	"math/rand"

	"github.com/tbogdala/noisey"
	"os"
)

func main() {
	fmt.Println("Doing a noisey test.")

	const imageSize int = 128

	// make a test generator seeded to 1
	r := rand.New(rand.NewSource(int64(1)))

	// create a new perlin noise generator with 'standard' quality
	perlin := noisey.NewPerlinGenerator2D(r, noisey.StandardQuality)

	// create the fractal Brownian motion generator based on perlin
	fbmPerlin := noisey.NewFBMGenerator2D(&perlin, 1.0, 0.5, 2.0, 1.0)

	// make an ascii pixel image by calculating random noise
	pixels := make([]float64, imageSize*imageSize)
	for y := 0; y < imageSize; y++ {
		for x := 0; x < imageSize; x++ {
			v := fbmPerlin.Get2D(float64(x)*0.1, float64(y)*0.1)
			v = v*0.5 + 0.5
			pixels[y*imageSize+x] = v
		}
	}

	// print the image out to the terminal using the symbols below
	symbols := []string{" ", "░", "▒", "▓", "█", "█"}
	out := bufio.NewWriter(os.Stdout)
	for y := 0; y < imageSize; y++ {
		for x := 0; x < imageSize; x++ {
			fmt.Fprint(out, symbols[int(pixels[y*imageSize+x]/0.2)])
		}
		fmt.Fprintln(out)
	}
	out.Flush()
}
