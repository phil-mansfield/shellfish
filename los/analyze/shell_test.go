package analyze

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"
)

func sphere(r float64) Shell {
	return func(phi, theta float64) float64 {
		return r
	}
}

func brokenSphere(r1, r2 float64) Shell {
	return func(phi, theta float64) float64 {
		//if theta > math.Pi/2 {
		//	return r1
		//}
		//return r2

		if phi > math.Pi {
			return r1
		}
		return r2
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
	s := ellipsoid(2, 4, 3)
	//s := brokenSphere(2, 1)
	samples := 1000 * 1000
	fmt.Printf("Volume: %8.4g\n", s.Volume(samples))
	a, b, c, aVec := s.Axes(samples)
	fmt.Printf("Axes: %8.4g %8.4g %8.4g\n", a, b, c)
	fmt.Printf("Printiple Axis: %8.4g\n", aVec)
	fmt.Printf("Area: %8.4g\n", s.SurfaceArea(samples))
}
