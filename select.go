package noisey

/* Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
See the LICENSE file for more details. */

// Select2D is a module that uses SourcesA or SourceB depending
// on the value coming from Control. If the value from control is between
// LowerBound and UpperBound then it uses SourceB, but otherwise it will
// use SourceA.
type Select2D struct {
	// the first channel of noise that the select module uses
	SourceA NoiseyGet2D

	// the second channel of noise that the select module uses
	SourceB NoiseyGet2D

	// a channel of noise that determines whether SourceA or SourceB gets
	// used as an output for a given coordinate
	Control NoiseyGet2D

	// if the value of Control is above LowerBound but below Upper Bound
	// then SourceB is output, otherwise SourceA is output
	LowerBound float64

	// if the value of Control is above LowerBound but below Upper Bound
	// then SourceB is output, otherwise SourceA is output
	UpperBound float64

	// the folloff value at the edge of a transition -- if <=0.0 there is an
	// abrupt transition from SourceA to SourceB, otherwise it specifies
	// the width of the transition range where values are blended between
	// the two sources
	EdgeFalloff float64
}

// NewSelect2D creates a new selector 2d module.
func NewSelect2D(a, b, c NoiseyGet2D, lower float64, upper float64, edge float64) (selector Select2D) {
	selector.SourceA = a
	selector.SourceB = b
	selector.Control = c
	selector.LowerBound = lower
	selector.UpperBound = upper
	selector.EdgeFalloff = edge
	return
}

// Get2D calculates the noise value using SourceA or SourceB depending on Control.
func (selector *Select2D) Get2D(x float64, y float64) (v float64) {
	control := selector.Control.Get2D(x, y)

	// process the lack of edge falloff first
	if selector.EdgeFalloff <= 0.0 {
		if selector.LowerBound < control && control < selector.UpperBound {
			return selector.SourceB.Get2D(x, y)
		}
		return selector.SourceA.Get2D(x, y)
	}

	// if we got here, then edge falloff is positive
	if control < selector.LowerBound-selector.EdgeFalloff {
		return selector.SourceA.Get2D(x, y)
	}
	if control < selector.LowerBound+selector.EdgeFalloff {
		lower := selector.LowerBound - selector.EdgeFalloff
		upper := selector.LowerBound + selector.EdgeFalloff
		v := (control - lower) / (upper - lower)
		lerpControl := calcCubicSCurve(v)
		a := selector.SourceA.Get2D(x, y)
		b := selector.SourceB.Get2D(x, y)
		return lerp(a, b, lerpControl)
	}
	if control < selector.UpperBound-selector.EdgeFalloff {
		return selector.SourceB.Get2D(x, y)
	}
	if control < selector.UpperBound+selector.EdgeFalloff {
		lower := selector.UpperBound - selector.EdgeFalloff
		upper := selector.UpperBound + selector.EdgeFalloff
		v := (control - lower) / (upper - lower)
		lerpControl := calcCubicSCurve(v)
		a := selector.SourceA.Get2D(x, y)
		b := selector.SourceB.Get2D(x, y)
		return lerp(b, a, lerpControl)
	}
	return selector.SourceA.Get2D(x, y)
}

// Select3D is a module that uses SourcesA or SourceB depending
// on the value coming from Control. If the value from control is between
// LowerBound and UpperBound then it uses SourceB, but otherwise it will
// use SourceA.
type Select3D struct {
	// the first channel of noise that the select module uses
	SourceA NoiseyGet3D

	// the second channel of noise that the select module uses
	SourceB NoiseyGet3D

	// a channel of noise that determines whether SourceA or SourceB gets
	// used as an output for a given coordinate
	Control NoiseyGet3D

	// if the value of Control is above LowerBound but below Upper Bound
	// then SourceB is output, otherwise SourceA is output
	LowerBound float64

	// if the value of Control is above LowerBound but below Upper Bound
	// then SourceB is output, otherwise SourceA is output
	UpperBound float64

	// the folloff value at the edge of a transition -- if <=0.0 there is an
	// abrupt transition from SourceA to SourceB, otherwise it specifies
	// the width of the transition range where values are blended between
	// the two sources
	EdgeFalloff float64
}

// NewSelect3D creates a new selector 3d module.
func NewSelect3D(a, b, c NoiseyGet3D, lower float64, upper float64, edge float64) (selector Select3D) {
	selector.SourceA = a
	selector.SourceB = b
	selector.Control = c
	selector.LowerBound = lower
	selector.UpperBound = upper
	selector.EdgeFalloff = edge
	return
}

// Get3D calculates the noise value using SourceA or SourceB depending on Control.
func (selector *Select3D) Get3D(x, y, z float64) (v float64) {
	control := selector.Control.Get3D(x, y, z)

	// process the lack of edge falloff first
	if selector.EdgeFalloff <= 0.0 {
		if selector.LowerBound < control && control < selector.UpperBound {
			return selector.SourceB.Get3D(x, y, z)
		}
		return selector.SourceA.Get3D(x, y, z)
	}

	// if we got here, then edge falloff is positive
	if control < selector.LowerBound-selector.EdgeFalloff {
		return selector.SourceA.Get3D(x, y, z)
	}
	if control < selector.LowerBound+selector.EdgeFalloff {
		lower := selector.LowerBound - selector.EdgeFalloff
		upper := selector.LowerBound + selector.EdgeFalloff
		v := (control - lower) / (upper - lower)
		lerpControl := calcCubicSCurve(v)
		a := selector.SourceA.Get3D(x, y, z)
		b := selector.SourceB.Get3D(x, y, z)
		return lerp(a, b, lerpControl)
	}
	if control < selector.UpperBound-selector.EdgeFalloff {
		return selector.SourceB.Get3D(x, y, z)
	}
	if control < selector.UpperBound+selector.EdgeFalloff {
		lower := selector.UpperBound - selector.EdgeFalloff
		upper := selector.UpperBound + selector.EdgeFalloff
		v := (control - lower) / (upper - lower)
		lerpControl := calcCubicSCurve(v)
		a := selector.SourceA.Get3D(x, y, z)
		b := selector.SourceB.Get3D(x, y, z)
		return lerp(b, a, lerpControl)
	}
	return selector.SourceA.Get3D(x, y, z)
}
