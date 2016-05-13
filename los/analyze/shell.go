package analyze

import (
	"math"
	"math/rand"

	"github.com/phil-mansfield/shellfish/math/sort"
	"github.com/gonum/matrix/mat64"
)

type Shell func(phi, theta float64) float64

func randomAngle() (phi, theta float64) {
	u, v := rand.Float64(), rand.Float64()
	return 2 * math.Pi * u, math.Acos(2*v - 1)
}

func cartesian(phi, theta, r float64) (x, y, z float64) {
	sinP, cosP := math.Sincos(phi)
	sinT, cosT := math.Sincos(theta)
	return r * sinT * cosP, r * sinT * sinP, r * cosT
}

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

func (s Shell) MeanRadius(samples int) float64 {
	sum := 0.0
	for i := 0; i < samples; i++ {
		phi, th := randomAngle()
		r := s(phi, th)
		sum += r
	}
	return sum / float64(samples)
}

func (s Shell) MedianRadius(samples int) float64 {
	rs := make([]float64, samples)
	for i := range rs {
		phi, th := randomAngle()
		rs[i] = s(phi, th)
	}
	return sort.Median(rs, rs)
}

func trisort(x, y, z float64) (a, b, c float64) {
	var p, q float64
	switch {
	case x > y && x > z: a, p, q = x, y, z
	case y > x && y > z: a, p, q = y, z, x
	default: a, p, q = z, x, y
	}

	if p > q {
		return a, p, q
	} else {
		return a, q, p
	}
}

func (s Shell) Axes(samples int) (a, b, c float64) {
	nxx, nyy, nzz := 0.0, 0.0, 0.0
	nxy, nyz, nzx := 0.0, 0.0, 0.0
	nx, ny, nz := 0.0, 0.0, 0.0
	norm := 0.0

	for i := 0; i < samples; i++ {
		phi, theta := randomAngle()
		r := s(phi, theta)
		area := r*r / cosNorm(s, phi, theta)
		x, y, z := cartesian(phi, theta, r)

		nxx += area*x*x
		nyy += area*y*y
		nzz += area*z*z
		nxy += area*x*y
		nyz += area*y*z
		nzx += area*z*x
		nx += area*x
		ny += area*y
		nz += area*z

		norm += area
	}

	nxx, nyy, nzz = nxx/norm, nyy/norm, nzz/norm
	nxy, nyz, nzx = nxy/norm, nyz/norm, nzx/norm
	nx, ny, nz = nx/norm, ny/norm, nz/norm

	mat := mat64.NewDense(3, 3, []float64{
		nyy + nzz - ny*ny - nz*nz, -nxy + nx*ny, -nzx + nz*nx,
		-nxy + nx*ny, nxx + nzz - nx*nx - ny*ny, -nyz + ny*nz,
		-nzx + nz*nx, -nyz + ny*nz, nxx + nyy - nx*nx - ny*ny,
	})
	
	eigen := &mat64.Eigen{}
	ok := eigen.Factorize(mat, false)
	if !ok { panic("Could not factorize inertia tensor.") }

	vals := eigen.Values(nil)

	Ix, Iy, Iz := real(vals[0]), real(vals[1]), real(vals[2])
	ax2 := 3 * (Iy + Iz - Ix) / 2
	ay2 := 3*Iy - ax2
	az2 := 3*Iz - ax2

	return trisort(math.Sqrt(ax2), math.Sqrt(ay2), math.Sqrt(az2))
}

func cosNorm(s Shell, phi, theta float64) float64 	{
	dp, dt := 1e-3, 1e-3
	r00 := s(phi - dp, theta - dt)
	x00, y00, z00 := cartesian(phi - dp, theta - dt, r00)
	r01 := s(phi - dp, theta + dt)
	x01, y01, z01 := cartesian(phi - dp, theta + dt, r01)
	r10 := s(phi + dp, theta - dt)
	x10, y10, z10 := cartesian(phi + dp, theta - dt, r10)
	r11 := s(phi + dp, theta + dt)
	x11, y11, z11 := cartesian(phi + dp, theta + dt, r11)

	dxa, dya, dza := x00 - x11, y00 - y11, z00 - z11
	dxb, dyb, dzb := x01 - x10, y01 - y10, z01 - z10

	// normal vector
	xn := dya*dzb - dza*dyb
	yn := dza*dxb - dxa*dzb
	zn := dxa*dyb - dya*dxb
	norm := math.Sqrt(xn*xn + yn*yn + zn*zn)

	xn, yn, zn = xn/norm, yn/norm, zn/norm
	xl, yl, zl := cartesian(phi, theta, 1)
	return xl*xn + yl*yn + zl*zn
}

func (s Shell) SurfaceArea(samples int) float64 {
	sum := 0.0
	for i := 0; i < samples; i++ {
		phi, theta := randomAngle()
		r := s(phi, theta)
		sum += r*r / cosNorm(s, phi, theta)
	}
	return sum / float64(samples) * 4 * math.Pi
}

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

func  (s Shell) RadiusHistogram(
	samples, bins int, rMin, rMax float64,
) (rs, ns []float64) {
	rs, ns = make([]float64, bins), make([]float64, bins)
	dr := (rMax - rMin) / float64(bins)
	for i := range rs {
		rs[i] = rMin + dr*(float64(i) + 0.5)
	}

	count := 0
	for i := 0; i < samples; i++ {
		phi, theta := randomAngle()
		r := s(phi, theta)
		ri := (r - rMin) / dr
		if ri < 0 { continue }
		idx := int(ri)
		if idx >= bins { continue }
		ns[idx]++
		count++
	}

	for i := range ns {
		ns[i] /= float64(count) * dr
	}

	return rs, ns
}

func (s Shell) Contains(x, y, z float64) bool {
	r := math.Sqrt(x*x + y*y + z*z)
	phi := math.Atan2(y, x)
	theta := math.Acos(z / r)
	return s(phi, theta) > r
}
