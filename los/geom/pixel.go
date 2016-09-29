package geom

import (
	"math"
)

// DiskPixel returns the pixel index of the point at (r, phi) within a unit
// disk using the method of Gringorten & Yepez (1992).
func DiskPixel(r, theta float64, lvl int) int {
	dr := 1 / float64(2*lvl + 1)
	ir := int(r / dr)
	if ir == 0 {
		return 0
	}

	ir = (ir + 1) / 2
	if ir >= lvl {
		ir = lvl  // To fix floating point messiness.
	}
	ith := int(float64(8*ir) * theta / (2 * math.Pi))
	return (2*(ir - 1) + 1)*(2*(ir - 1) + 1) + ith
}

func DiskPixelNum(lvl int) int {
	return (lvl*2 + 1)*(lvl*2 + 1)
}

func SphericalPixel(phi, theta float64, lvl int) int {
	panic("NYI")
}

func SphericalPixelNum(lvl int) int {
	panic("NYI")
}