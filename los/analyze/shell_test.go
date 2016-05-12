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
	s := ellipsoid(3, 2, 1)
	samples := 100 * 1000
	fmt.Printf("Volume:  %8.3g\n", s.Volume(samples))
	Ix, Iy, Iz := s.Axes(samples)
	fmt.Printf("Moments: %8.3g %8.3g %8.3g\n", Ix, Iy, Iz)
	fmt.Printf("Area:    %8.3g\n", s.SurfaceArea(samples))
}
