package noisey

/* Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
See the LICENSE file for more details. */


// Select2D is a module that uses SourcesA or SourceB depending
// on the value coming from Control. If the value from control is between
// LowerBound and UpperBound then it uses SourceB, but otherwise it will
// use SourceA.
type Select2D struct {
  // the first channel of noise that the select module uses
  SourceA  BuilderGet2D

  // the second channel of noise that the select module uses
  SourceB  BuilderGet2D

  // a channel of noise that determines whether SourceA or SourceB gets
  // used as an output for a given coordinate
  Control  BuilderGet2D

  // if the value of Control is above LowerBound but below Upper Bound
  // then SourceB is output, otherwise SourceA is output
  LowerBound float64

  // if the value of Control is above LowerBound but below Upper Bound
  // then SourceB is output, otherwise SourceA is output
  UpperBound float64
}

// NewSelect2D creates a new selector 2d module.
func NewSelect2D(a, b, c BuilderGet2D, lower float64, upper float64) (selector Select2D) {
  selector.SourceA = a
  selector.SourceB = b
  selector.Control = c
  selector.LowerBound = lower
  selector.UpperBound = upper
  return
}

// Get2D calculates the noise value using SourceA or SourceB depending on Control.
func (selector *Select2D) Get2D(x float64, y float64) (v float64) {
  control := selector.Control.Get2D(x, y)

  if selector.LowerBound < control && control < selector.UpperBound {
    return selector.SourceB.Get2D(x, y)
  }
  return selector.SourceA.Get2D(x, y)
}
