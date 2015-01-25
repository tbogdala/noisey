package noisey

/* Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
See the LICENSE file for more details. */

import (
  "bytes"
  "encoding/json"
  "fmt"
)

type RandomSeedBuilder func(s int64)RandomSource

type GeneratorJSON struct {
  GeneratorType string
  Sources []string
  Octaves int
  Persistence float64
  Lacunarity float64
  Frequency float64
}

// SourceJSON describes the source of the random information. The key string
// used to define this in JSON will be the module type
type SourceJSON struct {
  SourceType string
  Quality int
  Seed string
}

type NoiseJSON struct {
  Seeds map[string]int64
  Sources map[string]SourceJSON
  Generators map[string]GeneratorJSON

  builtSeeds map[string]RandomSource
  builtSources map[string]CoherentRandomGen2D
  builtGenerators map[string]BuilderSource2D
}

func NewNoiseJSON() *NoiseJSON {
  nj := new(NoiseJSON)
  nj.Seeds = make(map[string]int64)
  nj.Sources = make(map[string]SourceJSON)
  nj.Generators = make(map[string]GeneratorJSON)

  nj.builtSeeds = make(map[string]RandomSource)
  nj.builtSources = make(map[string]CoherentRandomGen2D)
  nj.builtGenerators = make(map[string]BuilderSource2D)

  return nj
}

func LoadNoiseJSON(bytes []byte) (*NoiseJSON, error) {
  var cfg *NoiseJSON = NewNoiseJSON()
  err := json.Unmarshal(bytes, cfg)
  if err != nil {
    return nil, fmt.Errorf("Unable to read json into the configuration structure.\n%v\n", err)
  }

  return cfg, nil
}


func (cfg *NoiseJSON) GetGenerator(name string) BuilderSource2D {
  s, ok := cfg.builtGenerators[name]
  if ok == false {
    return nil
  }
  return s
}

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

func (cfg *NoiseJSON) BuildSources(seedBuilder RandomSeedBuilder) error {
  // loop through all configured sources
  for sourceName,source := range cfg.Sources {
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

    var s CoherentRandomGen2D
    switch source.SourceType {
      case "perlin2d":
        p2d := NewPerlinGenerator2D(r, source.Quality)
        s = CoherentRandomGen2D(&p2d)
      default:
        return fmt.Errorf("Undefined source type (%s) for source %s.\n", source.SourceType, sourceName)
    }

    // store the result
    cfg.builtSources[sourceName] = s
  }

  return nil
}


func (cfg *NoiseJSON) BuildGenerators() error {
  // loop through all configured generators
  for genName,gen := range cfg.Generators {
    // build the array of sources and if one's not found, error out
    sourceArray := make([]CoherentRandomGen2D, len(gen.Sources))
    for i,ss := range gen.Sources {
      builtSource, ok := cfg.builtSources[ss]
      if ok != true {
        return fmt.Errorf("Generator %s creation failed: couldn't find built source %s.\n", genName, ss)
      }
      sourceArray[i] = builtSource
    }

    var g BuilderSource2D
    switch gen.GeneratorType {
      case "fBm2d":
        fbm := NewFBMGenerator2D(sourceArray[0])
        fbm.Octaves = gen.Octaves
        fbm.Persistence = gen.Persistence
        fbm.Lacunarity = gen.Lacunarity
        fbm.Frequency = gen.Frequency
        g = BuilderSource2D(&fbm)
      default:
        return fmt.Errorf("Undefined generator type (%s) for generator %s.\n", gen.GeneratorType, genName)
    }

    // store the result
    cfg.builtGenerators[genName] = g
  }

  return nil
}
