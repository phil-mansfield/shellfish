package analyze

import (
	"math"
	"math/rand"

	"github.com/gonum/matrix/mat64"
	grid "github.com/phil-mansfield/shellfish/los/analyze/ellipse_grid"
	intr "github.com/phil-mansfield/shellfish/math/interpolate"
	"github.com/phil-mansfield/shellfish/math/sort"
)

// Shell is a function that returns the radius of a shell at a given set of
// angles.
//
// Unless otherwise specified, all quantities are calculated through Monte
// Carlo solid angle sampling.
type Shell func(phi, theta float64) float64

// randomAngle returns and angle chosen uniformly at random.
func randomAngle() (phi, theta float64) {
	u, v := rand.Float64(), rand.Float64()
	return 2 * math.Pi * u, math.Acos(2*v - 1)
}

// cartesian converts a tuple of radial coordinates to cartesian coordinates.
func cartesian(phi, theta, r float64) (x, y, z float64) {
	sinP, cosP := math.Sincos(phi)
	sinT, cosT := math.Sincos(theta)
	return r * sinT * cosP, r * sinT * sinP, r * cosT
}

// CartesianSampledVolume returns the volume of a Shell. The volume is
// calculated by Monte Carlo sampling of a sphere of radius rMax.
//
// This is slower than Volume for most shell shapes.
func (s Shell) CartesianSampledVolume(samples int, rMax float64) float64 {
	inside := 0
	for i := 0; i < samples; i++ {
		x := rand.Float64()*(2*rMax) - rMax
		y := rand.Float64()*(2*rMax) - rMax
		z := rand.Float64()*(2*rMax) - rMax

		r := math.Sqrt(x*x + y*y + z*z)
		phi := math.Atan2(y, x)
		th := math.Acos(z / r)

		rs := s(phi, th)
		if r < rs {
			inside++
		}
	}

	return float64(inside) / float64(samples) * (rMax * rMax * rMax * 8)
}

// Volume returns the volume of Shell.
func (s Shell) Volume(samples int) float64 {
	sum := 0.0
	for i := 0; i < samples; i++ {
		phi, theta := randomAngle()
		r := s(phi, theta)
		sum += r * r * r
	}
	r3 := sum / float64(samples)
	return r3 * 4 * (math.Pi / 3)
}

// MeanRadius returns the angle-weighted mean radius of a Shell.
func (s Shell) MeanRadius(samples int) float64 {
	sum := 0.0
	for i := 0; i < samples; i++ {
		phi, th := randomAngle()
		r := s(phi, th)
		sum += r
	}
	return sum / float64(samples)
}

// MedianRadius returns the angle-weighted median radius of a Shell.
func (s Shell) MedianRadius(samples int) float64 {
	rs := make([]float64, samples)
	for i := range rs {
		phi, th := randomAngle()
		rs[i] = s(phi, th)
	}
	return sort.Median(rs, rs)
}

// triSort returns x, y, and z in sorted order and returns the argument
// index of the largest value.
func trisort(x, y, z float64) (a, b, c float64, aIdx int) {
	var p, q float64
	switch {
	case x > y && x > z:
		a, p, q, aIdx = x, y, z, 0
	case y > x && y > z:
		a, p, q, aIdx = y, z, x, 1
	default:
		a, p, q, aIdx = z, x, y, 2
	}

	if p > q {
		return a, p, q, aIdx
	} else {
		return a, q, p, aIdx
	}
}

// Axes calculates the moment of inertia-equivalent axes of a Shell as well
// as the direction of the major axis.
func (s Shell) Axes(samples int) (a, b, c float64, aVec [3]float64) {

	// Temporarily approximate a constant-density ellipsoidal shell as
	// a homoeoid.

	nxx, nyy, nzz := 0.0, 0.0, 0.0
	nxy, nyz, nzx := 0.0, 0.0, 0.0
	nx, ny, nz := 0.0, 0.0, 0.0
	norm := 0.0

	for i := 0; i < samples; i++ {
		phi, theta := randomAngle()
		r := s(phi, theta)
		area := r * r / cosNorm(s, phi, theta)
		x, y, z := cartesian(phi, theta, r)

		nxx += area * x * x
		nyy += area * y * y
		nzz += area * z * z
		nxy += area * x * y
		nyz += area * y * z
		nzx += area * z * x
		nx += area * x
		ny += area * y
		nz += area * z

		norm += area
	}

	nxx, nyy, nzz = nxx/norm, nyy/norm, nzz/norm
	nxy, nyz, nzx = nxy/norm, nyz/norm, nzx/norm
	nx, ny, nz = nx/norm, ny/norm, nz/norm

	mat := mat64.NewDense(3, 3, []float64{
		nyy + nzz - ny*ny - nz*nz, -nxy + nx*ny, -nzx + nz*nx,
		-nxy + nx*ny, nxx + nzz - nx*nx - nz*nz, -nyz + ny*nz,
		-nzx + nz*nx, -nyz + ny*nz, nxx + nyy - nx*nx - ny*ny,
	})

	eigen := &mat64.Eigen{}
	ok := eigen.Factorize(mat, false)
	if !ok {
		panic("Could not factorize inertia tensor.")
	}

	vals := eigen.Values(nil)
	vecs := eigen.Vectors()

	Ix, Iy, Iz := real(vals[0]), real(vals[1]), real(vals[2])
	ax2 := 3 * (Iy + Iz - Ix) / 2
	ay2 := 3*Iy - ax2
	az2 := 3*Iz - ax2

	// Correct the axis ratios via empirically derived tables.

	// TODO: Fix naming conventions.

	c, b, a, aIdx := trisort(math.Sqrt(ax2), math.Sqrt(ay2), math.Sqrt(az2))
	ac, bc := a/c, b/c

	// TODO: This function is just barely not thread safe.

	acRatio := axisInterpolators.acRatio.Eval(ac, bc)
	bcRatio := axisInterpolators.bcRatio.Eval(ac, bc)
	cRatio := axisInterpolators.cRatio.Eval(ac, bc)

	aVec = [3]float64{
		vecs.At(0, aIdx), vecs.At(1, aIdx), vecs.At(2, aIdx),
	}
	norm = math.Sqrt(aVec[0]*aVec[0] + aVec[1]*aVec[1] + aVec[2]*aVec[2])
	aVec[0] = aVec[0] / norm
	aVec[1] = aVec[1] / norm
	aVec[2] = aVec[2] / norm

	c = cRatio * c
	return c, bcRatio * bc * c, acRatio * ac * c, aVec
}

// cosNorm reutrns the cosine of the angle between \hat{r} and the normal
// vector of the Shell surface at a particular angle.
func cosNorm(s Shell, phi, theta float64) float64 {
	dp, dt := 1e-3, 1e-3
	r00 := s(phi-dp, theta-dt)
	x00, y00, z00 := cartesian(phi-dp, theta-dt, r00)
	r01 := s(phi-dp, theta+dt)
	x01, y01, z01 := cartesian(phi-dp, theta+dt, r01)
	r10 := s(phi+dp, theta-dt)
	x10, y10, z10 := cartesian(phi+dp, theta-dt, r10)
	r11 := s(phi+dp, theta+dt)
	x11, y11, z11 := cartesian(phi+dp, theta+dt, r11)

	dxa, dya, dza := x00-x11, y00-y11, z00-z11
	dxb, dyb, dzb := x01-x10, y01-y10, z01-z10

	// normal vector
	xn := dya*dzb - dza*dyb
	yn := dza*dxb - dxa*dzb
	zn := dxa*dyb - dya*dxb
	norm := math.Sqrt(xn*xn + yn*yn + zn*zn)

	xn, yn, zn = xn/norm, yn/norm, zn/norm
	xl, yl, zl := cartesian(phi, theta, 1)
	return xl*xn + yl*yn + zl*zn
}

// SurfaceArea returns the surface area of a shell.
func (s Shell) SurfaceArea(samples int) float64 {
	sum := 0.0
	for i := 0; i < samples; i++ {
		phi, theta := randomAngle()
		r := s(phi, theta)
		sum += r * r / cosNorm(s, phi, theta)
	}
	return sum / float64(samples) * 4 * math.Pi
}

// DiffVolume returns the volume of the space between two Shells, s1 and s2.
func (s1 Shell) DiffVolume(s2 Shell, samples int) float64 {
	sum := 0.0
	for i := 0; i < samples; i++ {
		phi, theta := randomAngle()
		r1, r2 := s1(phi, theta), s2(phi, theta)
		r := (r1 + r2) / 2
		dr := math.Abs(r1 - r2)
		sum += dr * r * r
	}
	return sum / float64(samples) * (4 * math.Pi) / 3
}

// MaxDiff returns the maximum radial distance between two Shells along
// any line of sight.
func (s1 Shell) MaxDiff(s2 Shell, samples int) float64 {
	max := 0.0
	for i := 0; i < samples; i++ {
		phi, theta := randomAngle()
		r1, r2 := s1(phi, theta), s2(phi, theta)
		dr := math.Abs(r1 - r2)
		if dr > max {
			max = dr
		}
	}
	return max
}

// RadialRange returns the maximum and minimum radius of a Shell.
func (s Shell) RadialRange(samples int) (low, high float64) {
	phi, theta := randomAngle()
	low = s(phi, theta)
	high = low
	for i := 0; i < samples; i++ {
		phi, theta := randomAngle()
		r := s(phi, theta)
		if r > high {
			high = r
		}
		if r < low {
			low = r
		}
	}
	return low, high
}

// RadiusHistogram returns a normalized angle-weighted histogram of the radii
// of a Shell.
func (s Shell) RadiusHistogram(
	samples, bins int, rMin, rMax float64,
) (rs, ns []float64) {
	rs, ns = make([]float64, bins), make([]float64, bins)
	dr := (rMax - rMin) / float64(bins)
	for i := range rs {
		rs[i] = rMin + dr*(float64(i)+0.5)
	}

	count := 0
	for i := 0; i < samples; i++ {
		phi, theta := randomAngle()
		r := s(phi, theta)
		ri := (r - rMin) / dr
		if ri < 0 {
			continue
		}
		idx := int(ri)
		if idx >= bins {
			continue
		}
		ns[idx]++
		count++
	}

	for i := range ns {
		ns[i] /= float64(count) * dr
	}

	return rs, ns
}

func (s Shell) AngularFractionProfile(
	samples, bins int, rMin, rMax float64,
) (rs, fs []float64) {
	rs, fs = make([]float64, bins), make([]float64, bins)
	ns := make([]int, bins)

	lrMin, lrMax := math.Log(rMin), math.Log(rMax)
	dlr := (lrMax + lrMin) / float64(bins)
	for i := range rs {
		rs[i] = math.Exp(lrMin + (float64(i) + 0.5) * dlr)
	}

	for i := 0; i < samples; i++ {
		phi, theta := randomAngle()
		lr := math.Log(s(phi, theta))
		lri := int((lr - lrMin) / dlr)
		if lri < 0 || lri >= bins {
			continue
		}
		ns[lri]++
	}

	// reverse cumulative sum
	for i := bins - 2; i >= 0; i-- {
		ns[i] += ns[i+1]
	}

	for i := 0; i < bins; i++ {
		fs[i] = float64(ns[i]) / float64(ns[0])
	}

	return rs, fs
}

// Contains returns true if a Shell contains a point and false otherwise. The
// point must be in a coordinate system in which the Shell is at (0, 0, 0).
func (s Shell) Contains(x, y, z float64) bool {
	r := math.Sqrt(x*x + y*y + z*z)
	phi := math.Atan2(y, x)
	theta := math.Acos(z / r)
	return s(phi, theta) > r
}

// axisInterpolators contains state needed for Shell.Axes().
var axisInterpolators = struct {
	acRatio, bcRatio, cRatio intr.BiInterpolator
}{}

func init() {
	axisInterpolators.acRatio = intr.NewBiCubic(
		grid.ACRatios, grid.BCRatios,
		grid.ACCorrectionGrid,
	)
	axisInterpolators.bcRatio = intr.NewBiCubic(
		grid.ACRatios, grid.BCRatios,
		grid.BCCorrectionGrid,
	)
	axisInterpolators.cRatio = intr.NewBiCubic(
		grid.ACRatios, grid.BCRatios,
		grid.CCorrectionGrid,
	)
}
