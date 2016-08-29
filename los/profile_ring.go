package los

import (
	"math"

	"github.com/phil-mansfield/shellfish/los/geom"
)

// ProfileRing is a ring of uniformly spaced profiles which allows step
// functions to be added quickly to every profile.
//
// Here is an example usage of profile rings where a ring of 100 profiles with
// 200 bins each and with R ranges of [0, 1] are initialized. Then a hundred
// random step functions of height 3.5 are inserted to the 42nd profile.
//
//     ring := &ProfileRing{}
//     ring.Init(0, 1, 200, 100)
//     for i := 0; i < 100; i++ {
//         low, high := rand.Float64(), rand.Float64()
//         if low > high { lo, high = high, low }
//         ring.Insert(low, high, 3.5, 42)
//     }
//     out := make([]float64, 200)
//     ring.Retrieve(42, out)
//
// The methods Join() and Split() have also been provided which allow the state
// of a single profile ring to be split across many rings so that multiple
// threads may work on the same ring at the same time.
//
// The current implementation of ProfileRing maintains the derivatives of
// each profile in the ring and integrating only when a profile is requested
// for analysis.
type ProfileRing struct {
	derivs          []float64 // Contiguous block of pofile data. Column-major.
	Lines           []geom.Line
	bins            int // Length of an individual profile.
	n               int // Number of profiles.
	lowR, highR, dr float64
}

// Join adds two ProfileRings, p1 and p2, together and puts the results in the
// p1.
func (p1 *ProfileRing) Join(p2 *ProfileRing) {
	if p1.n != p2.n || p1.bins != p2.bins {
		panic("ProfileRing sizes do not match.")
	} else {
		for i, val := range p2.derivs {
			p1.derivs[i] += val
		}
	}
}

// Split splits the state of a profile ring, p1, so that it is shared by a
// second profile ring, p2. The state can later be rejoined by the Join()
// method.
func (p1 *ProfileRing) Split(p2 *ProfileRing) {
	if p1.n != p2.n || p1.bins != p2.bins {
		p2.Init(p1.lowR, p1.highR, p1.bins, p1.n)
	} else {
		p2.bins = p1.bins
		p2.lowR = p1.lowR
		p2.highR = p1.highR
		p2.dr = p2.dr

		for i := range p2.derivs {
			p2.derivs[i] = 0
		}
	}
}

// Init initializes a profile ring made up of n profiles each of which consist
// of the given number of radial bins and extend between the two specified
// radii.
func (p *ProfileRing) Init(lowR, highR float64, bins, n int) {
	p.derivs = make([]float64, bins*n)
	p.bins = bins
	p.n = n
	p.lowR = lowR
	p.highR = highR
	p.dr = (highR - lowR) / float64(bins)

	p.Lines = make([]geom.Line, n)
	for i := 0; i < n; i++ {
		sin, cos := math.Sincos(p.Angle(i))
		p.Lines[i].Init(0, 0, float32(cos), float32(sin))
	}
}

// Insert inserts a plateau with the given radial extent and density to the
// profile.
func (p *ProfileRing) Insert(start, end, rho float64, i int) {
	if end <= p.lowR || start >= p.highR {
		return
	}

	// One could be a bit more careful with floating point ops here, if
	// push comes to shove. In particular, Modf calls can be avoided trhough
	// trickery. (However, most recent benchmarks reveal that very little time
	// is spent in this method call anymore, so I wouldn't bother.)
	if start > p.lowR {
		fidx, rem := math.Modf((start - p.lowR) / p.dr)
		idx := int(fidx)
		p.derivs[i*p.bins+idx] += rho * (1 - rem)
		if idx < p.bins-1 {
			p.derivs[i*p.bins+idx+1] += rho * rem
		}
	} else {
		p.derivs[i*p.bins] += rho
	}

	if end < p.highR {
		fidx, rem := math.Modf((end - p.lowR) / p.dr)
		idx := int(fidx)
		p.derivs[i*p.bins+idx] -= rho * (1 - rem)
		if idx < p.bins-1 {
			p.derivs[i*p.bins+idx+1] -= rho * rem
		}
	}
}

// Retrieve does any neccessary post-processing on the specified profile and
// writes in to an out buffer.
func (p *ProfileRing) Retrieve(i int, out []float64) {
	sum := float64(0)
	for j := 0; j < p.bins; j++ {
		sum += p.derivs[j+p.bins*i]
		out[j] = sum
	}
}

// Angle returns the angle that the line with the specified index points in.
func (p *ProfileRing) Angle(i int) float64 {
	return math.Pi * 2 * float64(i) / float64(p.n)
}
