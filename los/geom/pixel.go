package geom

import (
	"math"
)

var cosPi4 = math.Cos(math.Pi/4)

// DiskPixel returns the pixel index of the point at (r, phi) within a unit
// disk using the method of Gringorten & Yepez (1992).
func DiskPixel(r, theta float64, lvl int) int {
	ir := int(r * float64(2*lvl + 1))
	if ir == 0 { return 0 }

	ir = (ir + 1) / 2
	if ir >= lvl { ir = lvl }
	ith := int(float64(8*ir) * theta / (2 * math.Pi))
	return (2*(ir - 1) + 1)*(2*(ir - 1) + 1) + ith
}

// SpherePixel returns the pixel index at (phi, theta) (i.e. azimuthal, polar)
// using a two-hemisphere variation on the method of Gringoten & Yepez (1992).
func SpherePixel(phi, theta float64, lvl int) int {
	if lvl == 0 { return 0 }
	if theta > math.Pi/2 {
		return DiskPixel(math.Cos(theta/2)/cosPi4, phi, lvl - 1)
	}
	return DiskPixelNum(lvl-1) +
		DiskPixel(math.Cos((math.Pi-theta)/2)/cosPi4, phi, lvl - 1)
}

// DiskPixelNum returns the number of disk pixels at the specified level.
func DiskPixelNum(lvl int) int {
	return (lvl*2 + 1)*(lvl*2 + 1)
}

// SpherePixelNum returns the number of sphere pixels at the specified level.
func SpherePixelNum(lvl int) int {
	if lvl == 0 { return 1 }
	return DiskPixelNum(lvl - 1)*2
}