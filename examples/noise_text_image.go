package main

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

	// create a new perlin noise generator with a 'standard' 256 int
	// table of random permutations
	perlin := noisey.NewPerlinGenerator2D(r, 256)

	// make an ascii pixel image by calculating random noise
	pixels := make([]float64, imageSize*imageSize)
	for i := 0; i < 100; i++ {
		for y := 0; y < imageSize; y++ {
			for x := 0; x < imageSize; x++ {
				v := perlin.Get(float64(x)*0.1, float64(y)*0.1, noisey.FastQuality)
				v = v*0.5 + 0.5
				pixels[y*imageSize+x] = v
			}
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
