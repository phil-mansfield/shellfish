package analyze

import (
	"math"

	"github.com/phil-mansfield/shellfish/los"
	"github.com/phil-mansfield/shellfish/los/geom"
)

// RingBuffer contains the locations of candidate splashback radii for a single
// ring of a LOS profiles. They are stored in multiple forms: x and y within
// the plane of the ring, x, y, and z in the coordinate system of the
// simulation, and r and phi within the plane of the ring. It also contains a
// bunch of buffers to prevent useless allocations in various operations.
type RingBuffer struct {
	PlaneXs, PlaneYs []float64 // x and y coords in the plane of the ring.
	Xs, Ys, Zs       []float64 // Cartesian coords in the simulation box.
	Rs, Phis         []float64 // r and phi coords in the plane of the ring.
	Oks              []bool // Corresponds to a valid splashback point.

	profRs, profRhos []float64
	smoothRhos, smoothDerivs []float64

	N, Bins int
}

// Init initializes
func (r *RingBuffer) Init(n, bins int) {
	r.N, r.Bins = n, bins

	r.PlaneXs, r.PlaneYs = make([]float64, n), make([]float64, n)
	r.Xs, r.Ys, r.Zs = make([]float64, n), make([]float64, n), make([]float64, n)
	r.Phis, r.Rs = make([]float64, n), make([]float64, n)
	r.Oks = make([]bool, n)

	r.smoothRhos, r.smoothDerivs = make([]float64, bins), make([]float64, bins)
	r.profRs, r.profRhos = make([]float64, bins), make([]float64, bins)
}

// Clear clears all the values contained within a RingBuffer.
func (r *RingBuffer) Clear() {
	for i := 0; i < r.N; i++ {
		r.PlaneXs[i], r.PlaneYs[i] = 0, 0
		r.Xs[i], r.Ys[i], r.Zs[i] = 0, 0, 0
		r.Phis[i], r.Rs[i] = 0, 0
		r.Oks[i] = false
	}

	for i := 0; i < r.Bins; i++ {
		r.smoothRhos[i], r.smoothDerivs[i] = 0, 0
		r.profRs[i], r.profRhos[i] = 0, 0
	}
}

// Splashback calculates the candidate splashback radius for a given line of
// sight and stores the relevant information in the RingBuffer.
func (r *RingBuffer) Splashback(
	h *los.Halo, ring int, window int, dLim float64,
) {
	h.GetRs(r.profRs)
	ls := new(geom.LineSegment)
	for i := 0; i < r.N; i++ {
		h.GetRhos(ring, i, r.profRhos)

		_, _, r.Oks[i] = Smooth(
			r.profRs, r.profRhos, window,
			Vals(r.smoothRhos), Derivs(r.smoothDerivs),
		)

		if !r.Oks[i] {
			continue
		}
		r.Rs[i], r.Oks[i] = SplashbackRadius(
			r.profRs, r.smoothRhos, r.smoothDerivs, DLim(dLim),
		)

		if !r.Oks[i] {
			continue
		}

		r.Phis[i] = float64(h.Phi(i))
		if r.Phis[i] < 0 {
			r.Phis[i] += math.Pi
		}
		sin, cos := math.Sincos(r.Phis[i])
		r.PlaneXs[i], r.PlaneYs[i] = cos*r.Rs[i], sin*r.Rs[i]

		h.LineSegment(ring, i, ls)
		r.Xs[i] = r.Rs[i] * float64(ls.Dir[0])
		r.Ys[i] = r.Rs[i] * float64(ls.Dir[1])
		r.Zs[i] = r.Rs[i] * float64(ls.Dir[2])
	}
}

// OkPlaneCoords returns the within-plane x and y coordinates where r.Oks
// is true.
func (r *RingBuffer) OkPlaneCoords(xs, ys []float64) (okXs, okYs []float64) {
	xs, ys = xs[:0], ys[:0]
	for i, ok := range r.Oks {
		if ok {
			xs = append(xs, r.PlaneXs[i])
			ys = append(ys, r.PlaneYs[i])
		}
	}
	return xs, ys
}

// OkPlaneCoords returns the within-plane r and phi coordinates where r.Oks
// is true.
func (r *RingBuffer) OkPolarCoords(rs, phis []float64) (okRs, okPhis []float64) {
	rs, phis = rs[:0], phis[:0]
	for i, ok := range r.Oks {
		if ok {
			rs = append(rs, r.Rs[i])
			phis = append(phis, r.Phis[i])
		}
	}
	return rs, phis
}
