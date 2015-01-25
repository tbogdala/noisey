package noisey

/* Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
See the LICENSE file for more details. */

/*

This module provides a way to save/load noise setting to/from a JSON file,
as well as providing creation methods to create objects for all of the seeds, coherent
random value sources and builder/selectors.

A sample JSON file would contain something like this:

{
  "Seeds": {
    "Default": 1
  },
  "Sources": {
    "perlin": {
      "SourceType": "perlin2d",
      "Quality": 2,
      "Seed": "Default"
    }
  },
  "Generators": {
    "basic": {
      "GeneratorType": "fBm2d",
      "Sources": [
        "perlin"
      ],
      "Octaves": 5,
      "Persistence": 0.25,
      "Lacunarity": 2,
      "Frequency": 1.13
    }
  }
}


A quick sample of what this looks like is here:

  import (
    "bytes"
    "io/ioutil"
  )

  bytes, err := ioutil.ReadFile(configFilename)
  if err != nil {
    panic(err)
  }

  noiseBank, err := noisey.LoadNoiseJSON(bytes)
  if err != nil {
    panic(err)
  }

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

This loads the JSON file into the structures in this module and then calls
BuildSources() and BuildGenerators() so that the seeds, sources and generator
modules are all created.

At this point you can get the generator from the noiseBank variable and
use it to get random numbers or put it inside a builder module to make
something else:

  fbmPerlin := noiseBank.GetGenerator("basic")
  builder := noisey.NewBuilder2D(fbmPerlin, imageSize, imageSize)
  builder.Bounds = noisey.Builder2DBounds{0.0, 0.0, 6.0, 6.0}
  builder.Build()


*/

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// RandomSeedBuilder is a type used to construct RandomSource interfaces
// from a seed listed in the configuration JSON
type RandomSeedBuilder func(s int64) RandomSource

// GeneratorJSON is a generic data structure for noise generators in
// noisey. Something like a C union, not all of the fields may be applicable
// to every generator type.
type GeneratorJSON struct {
	// GeneratorType is a type string to identify what generator to create
	// on BuildGenerators()
	GeneratorType string

	// Sources is an array of strings that are names in the NoiseJSON.Sources
	// map that are to be used in this generator.
	Sources []string

	// Generators is an array of strings that are names in the NoiseJSON.Generators
	// map that are to be used in this generator.
	Generators []string

	Octaves     int     // Octaves is generator specific ...
	Persistence float64 // Persistence is generator specific ...
	Lacunarity  float64 // Lacunarity is generator specific ...
	Frequency   float64 // Frequency is generator specific ...
	LowerBound float64 // LowerBound is generator specific ...
	UpperBound float64 // LowerBound is generator specific ...
	Scale float64 // Scale is generator specific ...
	Bias float64 // Scale is generator specific ...
}

// SourceJSON describes the source of the random information, like perlin2d.
type SourceJSON struct {
	// SourceType is a type string used to identify what source module to create
	// on BuildSources().
	SourceType string

	// Quality is source module specific ...
	Quality int

	// Seed is a string that needs to be a name in the NoiseJSON.Seeds map that
	// is to be used in this generator.
	Seed string
}

// NoiseJSON is a structure that facilities the saving and loading of JSON
// representations of a system of seeds, sources and generators of noise.
type NoiseJSON struct {
	// Seeds uses a name string as a key that can be referenced in SourceJSON
	// structures and can have predefined seed values. When calling BuildSources(),
	// a client may pass a function to build the actual RandomSource interface
	// and is therefore not bound to use this ...
	Seeds map[string]int64

	// Sources uses a name string as a key that can be referenced in other structures
	// and maps to a SoruceJSON structure that describes how the noise source
	// should be built.
	Sources map[string]SourceJSON

	// Generators uses a name string as a key that can be referenced in other
	// structures and maps to a GeneratorJSON structre that describes how the noise
	// generator should be built.
	Generators map[string]GeneratorJSON

	// builtSeeds are cached RandomSource random number sources built after BuildSources()
	builtSeeds map[string]RandomSource

	// builtSources are cached noise providers built after BuildSources()
	builtSources map[string]SourceGet2D

	// builtGenerators are cached noise generators built after BuildGenerators()
	builtGenerators map[string]BuilderGet2D
}

// NewNoiseJSON creates a new structure that can be used to save noise settings
// out to JSON or to load noise settings in from a JSON byte array.
func NewNoiseJSON() *NoiseJSON {
	nj := new(NoiseJSON)
	nj.Seeds = make(map[string]int64)
	nj.Sources = make(map[string]SourceJSON)
	nj.Generators = make(map[string]GeneratorJSON)

	nj.builtSeeds = make(map[string]RandomSource)
	nj.builtSources = make(map[string]SourceGet2D)
	nj.builtGenerators = make(map[string]BuilderGet2D)

	return nj
}

// LoadNoiseJSON unmarshals the JSON from the byte array and returns a NoiseJSON
// object on success; error otherwise.
func LoadNoiseJSON(bytes []byte) (*NoiseJSON, error) {
	var cfg *NoiseJSON = NewNoiseJSON()
	err := json.Unmarshal(bytes, cfg)
	if err != nil {
		return nil, fmt.Errorf("Unable to read json into the configuration structure.\n%v\n", err)
	}

	return cfg, nil
}

// GetGenerator returns a cached generator BuilderGet2D object. This function
// Must be called after both BuildSources() and BuildGenerators().
func (cfg *NoiseJSON) GetGenerator(name string) BuilderGet2D {
	s, ok := cfg.builtGenerators[name]
	if ok == false {
		return nil
	}
	return s
}

// SaveNoiseJSON marshals the structure into a JSON byte array that is indented nicely.
func (cfg *NoiseJSON) SaveNoiseJSON() ([]byte, error) {
	rawBytes, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("Unable to encode the configuration structure into JSON.\n%v\n", err)
	}

	// format them nicely
	var b bytes.Buffer
	json.Indent(&b, rawBytes, "", "\t")

	return b.Bytes(), nil
}

// BuildSources takes a RandomSeedbuilder function as a parameter to create
// the actual random number generators from the seed provided and then
// creates the SourceGet2D interface objects based off of settings from
// SourceJSON structures in NoiseJSON.Sources. This method should be
// called before BuildGenerators().
func (cfg *NoiseJSON) BuildSources(seedBuilder RandomSeedBuilder) error {
	// loop through all configured sources
	for sourceName, source := range cfg.Sources {
		// try to get the same seede if it's already built
		r, ok := cfg.builtSeeds[source.Seed]
		if ok != true {
			seed, ok := cfg.Seeds[source.Seed]
			if ok != true {
				seed = 1
			}

			// get the random source by taking the referenced seed and calling
			// the seedBuilder() function with it that was passed in.
			r = seedBuilder(seed)

			// store the result for future generators to share
			cfg.builtSeeds[source.Seed] = r
		}

		var s SourceGet2D
		switch source.SourceType {
		case "perlin2d":
			p2d := NewPerlinGenerator2D(r, source.Quality)
			s = SourceGet2D(&p2d)
		case "opensimplex2d":
			os2d := NewOpenSimplexGenerator2D(r)
			s = SourceGet2D(&os2d)
		default:
			return fmt.Errorf("Undefined source type (%s) for source %s.\n", source.SourceType, sourceName)
		}

		// store the result
		cfg.builtSources[sourceName] = s
	}

	return nil
}

// BuildGenerators creates BuilderGet2D interface objects based off of the settings
// in the GeneratorJSON objects in NoiseJSON.Gnerators. This method should be
// called after BuildSources().
func (cfg *NoiseJSON) BuildGenerators() error {
	// loop through all configured generators
	for genName, gen := range cfg.Generators {
		var sourceArray []SourceGet2D
		var genArray []BuilderGet2D

		// build the array of sources and if one's not found, then return an error
		if gen.Sources != nil {
			sourceArray = make([]SourceGet2D, len(gen.Sources))
			for i, ss := range gen.Sources {
				builtSource, ok := cfg.builtSources[ss]
				if ok != true {
					return fmt.Errorf("Generator \"%s\" creation failed: couldn't find built source \"%s\".\n", genName, ss)
				}
				sourceArray[i] = builtSource
			}
		}

		// build the array of generators and if one's not found, then return an error
		if gen.Generators != nil {
			genArray = make([]BuilderGet2D, len(gen.Generators))
			for i, ss := range gen.Generators {
				builtGen, ok := cfg.builtGenerators[ss]
				if ok != true {
					return fmt.Errorf("Generator \"%s\" creation failed: couldn't find built source \"%s\".\n", genName, ss)
				}
				genArray[i] = builtGen
			}
		}

		var g BuilderGet2D
		switch gen.GeneratorType {
		case "fBm2d":
			fmt.Printf("fBm2d name : %s\n", genName)
			fbm := NewFBMGenerator2D(sourceArray[0])
			fbm.Octaves = gen.Octaves
			fbm.Persistence = gen.Persistence
			fbm.Lacunarity = gen.Lacunarity
			fbm.Frequency = gen.Frequency
			g = BuilderGet2D(&fbm)
		case "select2d":
			sel := NewSelect2D(genArray[0], genArray[1], genArray[2], gen.LowerBound, gen.UpperBound)
			g = BuilderGet2D(&sel)
		case "scale2d":
			scale := NewScale2D(genArray[0], gen.Scale, gen.Bias)
			g = BuilderGet2D(&scale)
		default:
			return fmt.Errorf("Undefined generator type (%s) for generator %s.\n", gen.GeneratorType, genName)
		}

		// store the result
		cfg.builtGenerators[genName] = g
	}

	return nil
}
