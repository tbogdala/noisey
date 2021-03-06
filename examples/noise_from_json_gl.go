/* Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
   See the LICENSE file for more details. */

package main

/*

This is a test module that does the following:

1) Creates an OpenGL window
2) Creates an RGB texture from noise described in a JSON config file
3) Displays the noise as a texture on a plane in the window

It requires the GLFW3 and GLEW libraries as well as the Go wrappers
for them: go-gl/gl and go-gl/glfw3.

Basic build instructions are:

	go get github.com/go-gl/gl/v3.3-core/gl
	go get github.com/go-gl/glfw/v3.1/glfw
	go get github.com/tbogdala/noisey
	cd $GOHOME/src/github.com/tbogdala/noisey/examples
	./build.sh
	./noise_from_json_gl

Hit `esc` to quit the program.
Hit `r` to reload the JSON file and compute the noise again!
Hit `c` to toggle the colorize effect.

*/

import (
	"fmt"
	gl "github.com/go-gl/gl/v3.3-core/gl"
	glfw "github.com/go-gl/glfw/v3.1/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/tbogdala/noisey"
	"io/ioutil"
	"math"
	"math/rand"
)

var (
	configFilename                    = "noise.json"
	noiseBank       *noisey.NoiseJSON = nil
	noiseTex        uint32
	colorizeEnabled bool = true
	imageSize            = int32(512)

	app   *ExampleApp
	plane *Renderable
)

func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyEscape && action == glfw.Press {
		w.SetShouldClose(true)
	}
	if key == glfw.KeyR && action == glfw.Press {
		fmt.Println("Reloading noise bank from JSON file...")
		loadJSONFile()
		randomPixels := generateNoiseImage(imageSize)
		gl.BindTexture(gl.TEXTURE_2D, noiseTex)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, imageSize, imageSize, 0, gl.RGB, gl.UNSIGNED_BYTE, gl.Ptr(randomPixels))
	}
	if key == glfw.KeyC && action == glfw.Press {
		colorizeEnabled = !colorizeEnabled
		if colorizeEnabled {
			fmt.Println("Colorizing the noise according to a gradient ...")
		} else {
			fmt.Println("Displaying noise as a grayscale image ...")
		}
		randomPixels := generateNoiseImage(imageSize)
		gl.BindTexture(gl.TEXTURE_2D, noiseTex)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, imageSize, imageSize, 0, gl.RGB, gl.UNSIGNED_BYTE, gl.Ptr(randomPixels))

	}
}

// createTextureFromRGB makes an OpenGL texture and buffers the RGB data into it
func createTextureFromRGB(rgb []byte, imageSize int32) (tex uint32) {
	gl.GenTextures(1, &tex)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, tex)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, imageSize, imageSize, 0, gl.RGB, gl.UNSIGNED_BYTE, gl.Ptr(rgb))

	return tex
}

func generateNoiseImage(imageSize int32) []byte {
	// create the fractal Brownian motion generator based on perlin
	fbmPerlin := noiseBank.GetGenerator("basic")

	// make an pixel image by calculating random noise and creating
	// an RGB byte triplet array based off the scaled noise value
	builder := noisey.NewBuilder2D(fbmPerlin, int(imageSize), int(imageSize))
	builder.Bounds = noisey.Builder2DBounds{0.0, 0.0, float64(imageSize) * 0.01, float64(imageSize) * 0.01}
	builder.Build()

	colors := make([]byte, imageSize*imageSize*3)
	for y := 0; y < builder.Height; y++ {
		for x := 0; x < builder.Width; x++ {
			v := builder.Values[(y*builder.Width)+x]
			b := byte(math.Floor((v*0.5 + 0.5) * 255)) // normalize 0..1 then scale by 255
			colorIndex := y*int(imageSize)*3 + x*3

			if colorizeEnabled {
				if b > 250 { // snow
					colors[colorIndex] = 255
					colors[colorIndex+1] = 255
					colors[colorIndex+2] = 255
				} else if b > 190 { // rock
					colors[colorIndex] = 128
					colors[colorIndex+1] = 128
					colors[colorIndex+2] = 128
				} else if b > 160 { // dirt
					colors[colorIndex] = 224
					colors[colorIndex+1] = 224
					colors[colorIndex+2] = 0
				} else if b > 130 { // grass
					colors[colorIndex] = 32
					colors[colorIndex+1] = 160
					colors[colorIndex+2] = 0
				} else if b > 125 { // sand
					colors[colorIndex] = 240
					colors[colorIndex+1] = 240
					colors[colorIndex+2] = 64
				} else if b > 120 { // shore
					colors[colorIndex] = 0
					colors[colorIndex+1] = 128
					colors[colorIndex+2] = 255
				} else if b > 32 { // shallow
					colors[colorIndex] = 0
					colors[colorIndex+1] = 0
					colors[colorIndex+2] = 255
				} else { // deeps
					colors[colorIndex] = 0
					colors[colorIndex+1] = 0
					colors[colorIndex+2] = 128
				}
			} else {
				colors[colorIndex] = b
				colors[colorIndex+1] = b
				colors[colorIndex+2] = b
			}
		}
	}

	return colors
}

func loadJSONFile() {
	// load the actual JSON configuration file
	fmt.Printf("Loading JSON configuration file bytes...\n")
	bytes, err := ioutil.ReadFile(configFilename)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Parsing the JSON ...\n")
	noiseBank, err = noisey.LoadNoiseJSON(bytes)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Parsing complete!\n")

	// build the sources from the JSON file
	err = noiseBank.BuildSources(func(s int64) noisey.RandomSource {
		return rand.New(rand.NewSource(int64(s)))
	})
	if err != nil {
		panic(err)
	}
	err = noiseBank.BuildGenerators()
	if err != nil {
		panic(err)
	}
}

func renderCallback(delta float64) {
	gl.Viewport(0, 0, int32(app.Width), int32(app.Height))
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	// make the projection and view matrixes
	projection := mgl.Ident4()
	view := mgl.Ident4()

	plane.Draw(projection, view)
}

func main() {
	app = NewApp()
	app.InitGraphics("Noisey Perlin Test", 768, 768)
	app.SetKeyCallback(keyCallback)
	app.OnRender = renderCallback

	// compile our shader
	var err error
	textureShader, err := LoadShaderProgram(UnlitTextureVertShader, UnlitTextureFragShader)
	if err != nil {
		panic("Failed to compile the shader! " + err.Error())
	}

	// load the JSON configuration file
	loadJSONFile()

	// create the plane to draw as a test
	plane = CreatePlaneXY(-0.75, -0.75, 0.75, 0.75, 1.0)
	plane.Shader = textureShader

	// generate the noise and make a image
	randomPixels := generateNoiseImage(imageSize)
	noiseTex = createTextureFromRGB(randomPixels, imageSize)
	plane.Tex0 = noiseTex

	app.RenderLoop()
}
