package analyze

import (
	"math"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func sphere(r float64) Shell {
	return func(phi, theta float64) float64 {
		return r
	}
}

func ellipsoid(a, b, c float64) Shell {
	return func(phi, theta float64) float64 {
		sp, cp := math.Sincos(phi)
		st, ct := math.Sincos(theta)

		return 1 / math.Sqrt((cp*cp*st*st/(a*a) +
			sp*sp*st*st/(b*b) + ct*ct/(c*c)))
	}
}

func TestEverything(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	s := ellipsoid(1, 1, 2)
	samples := 100 * 1000
	fmt.Printf("Volume:  %8.5g\n", s.Volume(samples))
	a, b, c := s.Axes(samples)
	fmt.Printf("Moments: %8.5g %8.5g %8.5g\n", a, b, c)
	fmt.Printf("Area:    %8.5g\n", s.SurfaceArea(samples))
}
