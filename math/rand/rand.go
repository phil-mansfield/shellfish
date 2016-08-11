/*package rand provides implementations of various types of pseudo random
number generators. Due to the similarity in form, it also provides an
implementation of Sobol sequences.

Here are some usage examples for these generators.

	// Generate a single value
    gen := New(Xorshift, 1337)
    x := gen.Uniform(3, 7)

	// Multiple random floats (faster)
	xs := make([]float64, 100)
	gen.UniformAt(3, 7, xs)

    // Random int
    y := gen.UniformInt(3, 7)

    // Use the time as a seed
    gen2 := NewTimeSeed(Xorshift)

Currently three types of generators are provided. The first is Xorshift, which
is very fast, the second is Tausworthe which is slower (especially at start
up), but which is a better generator, and the last is Goland, which is a
wrapper around Go's standard library generator.
*/
package rand

import (
	"math"
	"time"
)

// generatorBackend is an interface which is used by the generators to supply
// the functionality needed for top-level functions like Uniform().
type generatorBackend interface {
	Init(seed uint64)
	Next() float64
	NextSequence(target []float64)
}

// Generator is a random number generator.
type Generator struct {
	backend        generatorBackend
}

// Generator type is a flag used to indicate the desired algorithm
// for a random number generator.
type GeneratorType uint8

const (
	Xorshift GeneratorType = iota
	Golang
	Tausworthe
)

// NewTimeSeed returns a new random number generator that uses the current
// time as the seed.
func NewTimeSeed(gt GeneratorType) *Generator {
	return New(gt, uint64(time.Now().UnixNano()))
}

// New returns a new random number generator.
func New(gt GeneratorType, seed uint64) *Generator {
	var backend generatorBackend

	switch gt {
	case Xorshift:
		backend = new(xorshiftGenerator)
	case Golang:
		backend = new(golangGenerator)
	case Tausworthe:
		backend = new(tauswortheGenerator)
	default:
		panic("Unrecognized GeneratorType")
	}

	backend.Init(seed)
	gen := &Generator{backend}
	return gen
}

// UniformInt returns an integer uniformly at random within in the
// range [low, high).
func (gen *Generator) UniformInt(low, high int) int {
	f := gen.backend.Next()
	return int(math.Floor(float64(high-low)*f + float64(low)))

}

// Uniform returns a float uniformly at random within the range [low, high).
func (gen *Generator) Uniform(low, high float64) float64 {
	if low == 0.0 && high == 1.0 {
		return gen.backend.Next()
	}
	return (gen.backend.Next() * (high - low)) + low
}

// UniformAt writes floats generated uniformly at random in the range
// [low, high) to every element in a target slice. This is generally faster
// than calling Uniform the corresponding number of times.
func (gen *Generator) UniformAt(low, high float64, target []float64) {
	gen.backend.NextSequence(target)
	if low == 0.0 && high == 1.0 {
		return
	}
	for i := range target {
		target[i] = target[i]*(high-low) + low
	}
}
