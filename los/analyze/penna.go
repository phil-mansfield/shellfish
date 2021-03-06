package analyze

import (
	"math"

	"github.com/phil-mansfield/shellfish/los"
	"github.com/phil-mansfield/shellfish/math/mat"

	"github.com/gonum/matrix/mat64"
)

// PennaCoeffs calculates the Penna-Dines coefficients corresponding to the
// parameters I, J, and K for a set of input points.
func PennaCoeffs(xs, ys, zs []float64, I, J, K int) []float64 {
	N := len(xs)
	// TODO: Pass buffers to the function.
	rs := make([]float64, N)
	cosths := make([]float64, N)
	sinths := make([]float64, N)
	cosphis := make([]float64, N)
	sinphis := make([]float64, N)
	cs := make([]float64, I*J*K)

	// Precompute trig functions.
	for i := range rs {
		rs[i] = math.Sqrt(xs[i]*xs[i] + ys[i]*ys[i] + zs[i]*zs[i])
		cosths[i] = zs[i] / rs[i]
		sinths[i] = math.Sqrt(1 - cosths[i]*cosths[i])
		cosphis[i] = xs[i] / rs[i] / sinths[i]
		sinphis[i] = ys[i] / rs[i] / sinths[i]
	}

	MVals := make([]float64, I*J*K*len(xs))
	M := mat.NewMatrix(MVals, len(rs), I*J*K)

	// Populate matrix.
	for n := 0; n < N; n++ {
		m := 0
		for k := 0; k < K; k++ {
			costh := math.Pow(cosths[n], float64(k))
			for j := 0; j < J; j++ {
				sinphi := math.Pow(sinphis[n], float64(j))
				cosphi := 1.0
				for i := 0; i < I; i++ {
					MVals[m*M.Width+n] =
						math.Pow(sinths[n], float64(i+j)) *
							cosphi * costh * sinphi
					m++
					cosphi *= cosphis[n]
				}
			}
		}
	}

	// Solve.
	mat.VecMult(rs, pinv(M, M.Transpose()), cs)
	return cs
}

// PennaFunc returns a shell function correpsonding to a particular set of
// Penna-Dines coefficients.
func PennaFunc(cs []float64, I, J, K int) Shell {
	return func(phi, th float64) float64 {
		idx, sum := 0, 0.0
		sinPhi, cosPhi := math.Sincos(phi)
		sinTh, cosTh := math.Sincos(th)

		for k := 0; k < K; k++ {
			cosK := math.Pow(cosTh, float64(k))
			for j := 0; j < J; j++ {
				sinJ := math.Pow(sinPhi, float64(j))
				for i := 0; i < I; i++ {
					cosI := math.Pow(cosPhi, float64(i))
					sinIJ := math.Pow(sinTh, float64(i+j))
					sum += cs[idx] * sinIJ * cosK * sinJ * cosI
					idx++
				}
			}
		}
		return sum
	}
}

// PennaVolumeFit fits a Penna-Dines shell to a set of points constrained to
// a collection of planes belong to an los.Halo object.
//
// This function is essentially just a wrapper around PennaCoeffs.
func PennaVolumeFit(
	xs, ys [][]float64, h *los.Halo, I, J int,
) (cs []float64, shell Shell) {
	n := 0
	for i := range xs {
		n += len(xs[i])
	}
	fXs, fYs, fZs := make([]float64, n), make([]float64, n), make([]float64, n)

	idx := 0
	for i := range xs {
		for j := range xs[i] {
			fXs[idx], fYs[idx], fZs[idx] =
				h.PlaneToVolume(i, xs[i][j], ys[i][j])
			idx++
		}
	}

	cs = PennaCoeffs(fXs, fYs, fZs, I, J, 2)
	return cs, PennaFunc(cs, I, J, 2)
}

// FilterPoints applies the filtering algorithm from section 2.2.3 of
// Mansfield, Kravtsov, & Diemer (2016) to the points contained in each of
// a collection of RingBuffers.
//
// This function is mostly just a wrapper around functions from kde.go.
func FilterPoints(
	rs []RingBuffer, levels int, h float64,
) (pxs, pys [][]float64, ok bool) {
	pxs, pys = [][]float64{}, [][]float64{}
	for ri := range rs {
		r := &rs[ri]
		validXs := make([]float64, 0, r.N)
		validYs := make([]float64, 0, r.N)

		for i := 0; i < r.N; i++ {
			if r.Oks[i] {
				validXs = append(validXs, r.PlaneXs[i])
				validYs = append(validYs, r.PlaneYs[i])
			}
		}

		validRs, validPhis := []float64{}, []float64{}
		for i := range r.Rs {
			if r.Oks[i] {
				validRs = append(validRs, r.Rs[i])
				validPhis = append(validPhis, r.Phis[i])
			}
		}

		// If
		factor := 1.0
		fRs, fThs := []float64{}, []float64{}
		var (
			kt *KDETree
			ok bool
		)
		for i := 0; i < 10 && len(fRs) == 0; i++ {
			kt, ok = NewKDETree(validRs, validPhis, levels, h*factor)
			if !ok {
				return nil, nil, false
			}
			fRs, fThs, _ = kt.FilterNearby(validRs, validPhis, levels, kt.H())
			factor *= 1.1
		}

		fXs, fYs := make([]float64, len(fRs)), make([]float64, len(fRs))
		for i := range fRs {
			sin, cos := math.Sincos(fThs[i])
			fXs[i], fYs[i] = fRs[i]*cos, fRs[i]*sin
		}

		pxs, pys = append(pxs, fXs), append(pys, fYs)
	}

	return pxs, pys, true
}

// pinv calculates the pseudoinverse of a matrix, m, and its transpose, t.
// Why doesn't this function just calculate the transpose, you ask? Because
// mistakes were made.
func pinv(m, t *mat.Matrix) *mat.Matrix {
	// I HATE THIS
	// TODO: Make this function not painfully slow.
	gm := mat64.NewDense(m.Height, m.Width, m.Vals)
	gmt := mat64.NewDense(m.Width, m.Height, t.Vals)

	out1 := mat64.NewDense(m.Height, m.Height,
		make([]float64, m.Height*m.Height))
	out2 := mat64.NewDense(m.Width, m.Height,
		make([]float64, m.Height*m.Width))
	out1.Mul(gm, gmt)

	r, c := out1.Dims()
	inv := mat64.NewDense(c, r, make([]float64, r*c))
	inv.Inverse(out1)
	out2.Mul(gmt, inv)

	vals := make([]float64, m.Width*m.Height)
	for y := 0; y < m.Width; y++ {
		for x := 0; x < m.Height; x++ {
			vals[y*m.Height+x] = out2.At(y, x)
		}
	}
	return mat.NewMatrix(vals, m.Height, m.Width)
}