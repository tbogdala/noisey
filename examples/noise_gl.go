package main

/* Copyright 2014, Timothy Bogdala <tdb@animal-machine.com>
   See the LICENSE file for more details. */

/*

This is a test module that does the following:

1) Creates an OpenGL window
2) Creates an RGB texture from perlin noise
3) Displays the noise as a texture on a plane in the window

It requires the GLFW3 and GLEW libraries as well as the Go wrappers
for them: go-gl/gl and go-gl/glfw3.

Basic build instructions are:

	go get github.com/go-gl/gl
	go get github.com/go-gl/glfw3
	go get github.com/tbogdala/noisey
	cd $GOHOME/src/github.com/tbogdala/noisey/examples
	go build noise_gl.go
	./noise_gl

Hit `esc` to quit the program.

*/

import (
	"errors"
	"fmt"
	gl "github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
	"github.com/tbogdala/noisey"
	"math"
	"math/rand"
	"unsafe"
)

var (
	// vertex shader
	unlitTextureVertShader = `#version 330
  in vec3 position;
  in vec2 vertex_uv;
  out vec3 out_pos;
  out vec2 out_tex_coord;
  void main()
  {
    out_pos = position;
    out_tex_coord = vertex_uv;
    gl_Position = vec4(position, 1.0);
  }`

	// fragment shader
	unlitTextureFragShader = `#version 330
  uniform sampler2D diffuse_img;
  in vec3 out_pos;
  in vec2 out_tex_coord;
  out vec4 colourOut;
  void main()
  {
    colourOut = vec4(texture(diffuse_img, out_tex_coord).rgb, 1.0);
  }`
)

type Vec3 struct {
	X, Y, Z float32
}
type RenderMaterial struct {
	Shader gl.Program
	Tex0   gl.Texture
}
type DrawRenderable func(r Renderable)
type Renderable struct {
	Material    RenderMaterial
	Vao         gl.VertexArray
	VertVBO     gl.Buffer
	UvVBO       gl.Buffer
	NormsVBO    gl.Buffer
	ElementsVBO gl.Buffer
	FaceCount   int
	Scale       Vec3
	Location    Vec3
	DrawFunc    DrawRenderable
}

func (r Renderable) Draw() {
	r.DrawFunc(r)
}
func (r Renderable) Destroy() {
	r.VertVBO.Delete()
	r.UvVBO.Delete()
	r.NormsVBO.Delete()
	r.ElementsVBO.Delete()
	r.Vao.Delete()
}

func createPlane2dXY(shader gl.Program, x0, y0, x1, y1 float32) (r Renderable) {
	r.Material.Shader = shader
	r.FaceCount = 2
	r.Location = Vec3{0.0, 0.0, 0.0}
	r.Scale = Vec3{1.0, 1.0, 1.0}

	r.Vao = gl.GenVertexArray()

	verts := [12]float32{
		x0, y0, 0.0,
		x1, y0, 0.0,
		x0, y1, 0.0,
		x1, y1, 0.0,
	}
	indexes := [6]uint32{
		0, 1, 2,
		1, 3, 2,
	}
	uvs := [8]float32{
		0.0, 0.0,
		1.0, 0.0,
		0.0, 1.0,
		1.0, 1.0,
	}
	normals := [12]float32{
		0.0, 0.0, 1.0,
		0.0, 0.0, 1.0,
		0.0, 0.0, 1.0,
		0.0, 0.0, 1.0,
	}

	// calculate the memory size of floats used to calculate total memory size of float arrays
	floatSize := int(unsafe.Sizeof(gl.GLfloat(1.0)))
	uintSize := int(unsafe.Sizeof(gl.GLuint(1)))

	// create a VBO to hold the vertex data
	r.VertVBO = gl.GenBuffer()
	r.VertVBO.Bind(gl.ARRAY_BUFFER)
	gl.BufferData(gl.ARRAY_BUFFER, floatSize*len(verts), &verts, gl.STATIC_DRAW)

	// create a VBO to hold the normals data
	r.NormsVBO = gl.GenBuffer()
	r.NormsVBO.Bind(gl.ARRAY_BUFFER)
	gl.BufferData(gl.ARRAY_BUFFER, floatSize*len(normals), &normals, gl.STATIC_DRAW)

	// create a VBO to hold the uv data
	r.UvVBO = gl.GenBuffer()
	r.UvVBO.Bind(gl.ARRAY_BUFFER)
	gl.BufferData(gl.ARRAY_BUFFER, floatSize*len(uvs), &uvs, gl.STATIC_DRAW)

	// create a VBO to hold the face indexes
	r.ElementsVBO = gl.GenBuffer()
	r.ElementsVBO.Bind(gl.ELEMENT_ARRAY_BUFFER)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, uintSize*len(indexes), &indexes, gl.STATIC_DRAW)

	// customize the draw function
	r.DrawFunc = func(r Renderable) {
		r.Vao.Bind()

		gl.ActiveTexture(gl.TEXTURE0)
		r.Material.Tex0.Bind(gl.TEXTURE_2D)
		shaderTex1 := r.Material.Shader.GetUniformLocation("diffuse_img")
		shaderTex1.Uniform1i(0)

		shaderPosition := r.Material.Shader.GetAttribLocation("position")
		r.VertVBO.Bind(gl.ARRAY_BUFFER)
		shaderPosition.EnableArray()
		shaderPosition.AttribPointer(3, gl.FLOAT, false, 0, nil)

		shaderVertUv := r.Material.Shader.GetAttribLocation("vertex_uv")
		r.UvVBO.Bind(gl.ARRAY_BUFFER)
		shaderVertUv.EnableArray()
		shaderVertUv.AttribPointer(2, gl.FLOAT, false, 0, nil)

		r.ElementsVBO.Bind(gl.ELEMENT_ARRAY_BUFFER)
		gl.DrawElements(gl.TRIANGLES, r.FaceCount*3, gl.UNSIGNED_INT, nil)
		r.Vao.Unbind()
	}
	return
}

func errorCallback(err glfw.ErrorCode, desc string) {
	fmt.Printf("%v: %v\n", err, desc)
}

func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyEscape && action == glfw.Press {
		w.SetShouldClose(true)
	}
}

// loads shader objects and then attaches them to a program
func loadShaderProgram(vertShader, fragShader string) (gl.Program, error) {
	// create the program
	prog := gl.CreateProgram()

	// create the vertex shader
	vs := gl.CreateShader(gl.VERTEX_SHADER)
	vs.Source(vertShader)
	vs.Compile()
	vsCompiled := vs.Get(gl.COMPILE_STATUS)
	if vsCompiled != gl.TRUE {
		error := fmt.Sprintf("Failed to compile the vertex shader!\n%s", vs.GetInfoLog())
		fmt.Println(error)
		return prog, errors.New(error)
	}
	fmt.Println("Compiled the vertex shader ...")

	// create the fragment shader
	fs := gl.CreateShader(gl.FRAGMENT_SHADER)
	fs.Source(fragShader)
	fs.Compile()
	fsCompiled := fs.Get(gl.COMPILE_STATUS)
	if fsCompiled != gl.TRUE {
		error := fmt.Sprintf("Failed to compile the fragment shader!\n%s", fs.GetInfoLog())
		fmt.Println(error)
		return prog, errors.New(error)
	}
	fmt.Println("Compiled the fragment shader ...")

	// attach the shaders to the program and link
	prog.AttachShader(vs)
	prog.AttachShader(fs)
	prog.Link()
	progLinked := prog.Get(gl.LINK_STATUS)
	if progLinked != gl.TRUE {
		error := fmt.Sprintf("Failed to link the program!\n%s", prog.GetInfoLog())
		fmt.Println(error)
		return prog, errors.New(error)
	}
	fmt.Println("Shader program linked ...")

	// at this point the shaders can be deleted
	vs.Delete()
	fs.Delete()

	return prog, nil
}

// createTextureFromRGB makes an OpenGL texture and buffers the RGB data into it
func createTextureFromRGB(rgb []byte, imageSize int) gl.Texture {
	tex := gl.GenTexture()
	tex.Bind(gl.TEXTURE_2D)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, imageSize, imageSize, 0, gl.RGB, gl.UNSIGNED_BYTE, rgb)
	return tex
}

func generateNoiseImage(imageSize int, r noisey.RandomSource) []byte {
	// create a new perlin noise generator with HighQuality noise smoothing
	perlin := noisey.NewPerlinGenerator2D(r, noisey.HighQuality)

	// create the fractal Brownian motion generator based on perlin
	fbmPerlin := noisey.NewFBMGenerator2D(&perlin, 1, 0.5, 2.0, 1.0)

	// make an pixel image by calculating random noise and creating
	// an RGB byte triplet array based off the scaled noise value
	colors := make([]byte, imageSize*imageSize*3)
	for y := 0; y < imageSize; y++ {
		for x := 0; x < imageSize; x++ {
			v := fbmPerlin.Get2D(float64(x)*0.1, float64(y)*0.1)
			b := byte(math.Floor((v*0.5 + 0.5) * 255)) // normalize 0..1 then scale by 255
			colorIndex := y*imageSize*3 + x*3
			colors[colorIndex] = b
			colors[colorIndex+1] = b
			colors[colorIndex+2] = b
		}
	}

	return colors
}

func main() {
	// make sure that we display any errors that are encountered
	glfw.SetErrorCallback(errorCallback)

	if !glfw.Init() {
		panic("Can't init glfw!")
	}
	defer glfw.Terminate()

	// request a OpenGL 3.3 core context
	glfw.WindowHint(glfw.Samples, 4)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenglForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenglProfile, glfw.OpenglCoreProfile)

	// do the actual window creation
	window, err := glfw.CreateWindow(768, 768, "Noisey Perlin Test", nil, nil)
	if err != nil {
		panic(err)
	}

	window.SetKeyCallback(keyCallback)
	window.MakeContextCurrent()

	// make sure that GLEW initializes all of the GL functions
	if gl.Init() != 0 {
		panic("Failed to initialize GL and GLEW!")
	}

	// compile our shader
	prog, err := loadShaderProgram(unlitTextureVertShader, unlitTextureFragShader)
	if err != nil {
		panic("Failed to compile and link the shader program!")
	}
	prog.Use()

	// create the plane to draw as a test
	plane := createPlane2dXY(prog, -0.75, -0.75, 0.75, 0.75)

	// generate the noise and make a image
	r := rand.New(rand.NewSource(int64(1)))
	randomPixels := generateNoiseImage(512, r)
	noiseTex := createTextureFromRGB(randomPixels, 512)
	plane.Material.Tex0 = noiseTex

	// while there's no request to close the window
	for !window.ShouldClose() {
		width, height := window.GetFramebufferSize()

		gl.Viewport(0, 0, width, height)
		gl.ClearColor(0.0, 0.0, 0.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		plane.Draw()

		window.SwapBuffers()
		glfw.PollEvents()
	}

	plane.Material.Tex0.Delete()
	plane.Destroy()
}
