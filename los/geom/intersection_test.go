package geom

import (
	"math"
	"math/rand"
	"testing"

	"github.com/phil-mansfield/shellfish/render/io"
)

var (
	hx   = float32(58.68211)
	hy   = float32(59.48198)
	hz   = float32(8.70855)
	rMax = float32(0.6318755421407911)
	rMin = float32(0)

	L = [3]float32{0, 0, 1}

	hd  io.SheetHeader
	xs  [][3]float32
	ts  []Tetra
	pts []PluckerTetra
)

func randomizeTetra(t *Tetra, low, high float32) {
	for v := 0; v < 4; v++ {
		for i := 0; i < 3; i++ {
			t[v][i] = (high-low)*rand.Float32() + low
		}
	}
	t.Orient(+1)
}

func BenchmarkZPlaneSliceHit(b *testing.B) {
	n := 1000
	ts := make([]Tetra, n)
	pts := make([]PluckerTetra, n)
	poly := new(TetraSlice)
	for i := range ts {
		randomizeTetra(&ts[i], -1, +1)
		ts[i].Orient(+1)
		pts[i].Init(&ts[i])
	}

	for i := 0; i < b.N; i++ {
		idx := i % n
		ts[idx].ZPlaneSlice(&pts[idx], 0, poly)
	}
}

func BenchmarkZPlaneSliceMIss(b *testing.B) {
	n := 1000
	ts := make([]Tetra, n)
	pts := make([]PluckerTetra, n)
	poly := new(TetraSlice)
	for i := range ts {
		randomizeTetra(&ts[i], -1, +1)
		ts[i].Orient(+1)
		pts[i].Init(&ts[i])
	}

	for i := 0; i < b.N; i++ {
		idx := i % n
		ts[idx].ZPlaneSlice(&pts[idx], 1, poly)
	}
}

func BenchmarkAngleRange(b *testing.B) {
	n := 1000
	polys := make([]TetraSlice, 1)
	t, pt := new(Tetra), new(PluckerTetra)
	for len(polys) <= n {
		randomizeTetra(t, -1, +1)
		t.Orient(+1)
		pt.Init(t)

		if t.ZPlaneSlice(pt, 0, &polys[len(polys)-1]) {
			polys = append(polys, TetraSlice{})
		}
	}

	for i := 0; i < b.N; i++ {
		idx := i % n
		polys[idx].AngleRange()
	}
}

func randomAnchoredPluckerVec() *AnchoredPluckerVec {
	x := rand.Float32()
	y := rand.Float32()
	z := rand.Float32()
	norm := float32(math.Sqrt(float64(x*x + y*y + z*z)))

	P := &[3]float32{0, 0, 0}
	L := &[3]float32{x / norm, y / norm, z / norm}
	p := new(AnchoredPluckerVec)
	p.Init(P, L)
	return p
}

func BenchmarkPluckerTetraInit(b *testing.B) {
	ts := make([]Tetra, 1<<10)
	pts := make([]PluckerTetra, len(ts))
	for i := range ts {
		randomizeTetra(&ts[i], 0, 1)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		i := n % len(ts)
		pts[i].Init(&ts[i])
	}
}

func BenchmarkIntersectionBary(b *testing.B) {
	ts := make([]Tetra, 1<<10)
	pts := make([]PluckerTetra, len(ts))
	for i := range ts {
		randomizeTetra(&ts[i], 0, 1)
		pts[i].Init(&ts[i])
	}

	w := new(IntersectionWorkspace)
	ap := randomAnchoredPluckerVec()

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		i := n % len(ts)
		w.IntersectionBary(&pts[i], &ap.PluckerVec)
	}
}

func BenchmarkIntersectionDistance(b *testing.B) {
	ts := make([]Tetra, 1<<10)
	pts := make([]PluckerTetra, len(ts))
	for i := range ts {
		randomizeTetra(&ts[i], 0, 1)
		pts[i].Init(&ts[i])
	}

	w := new(IntersectionWorkspace)
	ap := randomAnchoredPluckerVec()

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		i := n % len(ts)
		w.IntersectionDistance(&pts[i], &ts[i], ap)
	}
}

func almostEq(x1, x2, eps float32) bool {
	return x1+eps > x2 && x1-eps < x2
}

func TestIntersectionDistance(t *testing.T) {
	P, L := [3]float32{-0.5, 1, 0.5}, [3]float32{1, 0, 0}
	eps := float32(1e-4)

	ap := new(AnchoredPluckerVec)
	pt := new(PluckerTetra)
	w := new(IntersectionWorkspace)

	ap.Init(&P, &L)

	table := []struct {
		t           Tetra
		enter, exit float32
		ok          bool
	}{
		{Tetra{{1, 0, 4}, {1, 4, 0}, {1, 0, 0}, {5, 0, 0}}, 1.5, 4, true},
		{Tetra{{1, 4, 0}, {1, 0, 4}, {3, 0, 0}, {5, 0, 0}}, 2.75, 4, true},
		{Tetra{{1, 0, 4}, {1, 4, 0}, {1, 0, 0}, {5, 0, 0}}, 1.5, 4, true},

		{Tetra{{-1, 4, 0}, {-1, 0, 4}, {-1, 0, 0}, {3, 0, 0}}, -0.5, 2, true},
		{Tetra{{9, 4, 0}, {9, 0, 4}, {9, 0, 0}, {13, 0, 0}}, 9.5, 12, true},

		{Tetra{{1, 0, 4}, {1, 0, 0}, {5, 0, 0}, {1, 4, 0}}, 1.5, 4, true},
		{Tetra{{1, 0, 0}, {5, 0, 0}, {1, 4, 0}, {1, 0, 4}}, 1.5, 4, true},
		{Tetra{{5, 0, 0}, {1, 4, 0}, {1, 0, 4}, {1, 0, 0}}, 1.5, 4, true},

		{Tetra{{1, 6, 0}, {1, 2, 4}, {1, 2, 0}, {5, 2, 0}}, 0, 0, false},
	}

	for i, test := range table {
		test.t.Orient(+1)
		pt.Init(&test.t)
		enter, exit, ok := w.IntersectionDistance(pt, &test.t, ap)
		if ok != test.ok {
			t.Errorf("%d) Expected ok = %v, but got %v.", i+1, test.ok, ok)
		} else if !almostEq(enter, test.enter, eps) {
			t.Errorf(
				"%d) Expected enter = %g, but got %g", i+1, test.enter, enter,
			)
		} else if !almostEq(exit, test.exit, eps) {
			t.Errorf(
				"%d) Expected leave = %g, but got %g", i+1, test.exit, exit,
			)
		}
	}
}

func almostContains(x, y float32, xs, ys []float32) bool {
	for i := range xs {
		if almostEq(xs[i], x, 1e-4) && almostEq(ys[i], y, 1e-4) {
			return true
		}
	}
	return false
}

func almostEqTetraSlice(poly *TetraSlice, xs, ys []float32) bool {
	if len(xs) != poly.Points {
		return false
	}
	for i := range xs {
		if !almostContains(poly.Xs[i], poly.Ys[i], xs, ys) {
			return false
		}
	}
	return true
}

func TestZPlaneSlice(t *testing.T) {
	tet := Tetra{{0, 4, 3}, {0, 4, -1}, {4, 4, -1}, {0, 8, -1}}
	xs, ys := []float32{0, 0, 3}, []float32{7, 4, 4}

	tet.Orient(+1)
	pt := new(PluckerTetra)
	pt.Init(&tet)
	poly := new(TetraSlice)

	if !tet.ZPlaneSlice(pt, 0, poly) {
		t.Errorf("z=0 did not slice a tetrahedron it intersected.")
	} else if !almostEqTetraSlice(poly, xs, ys) {
		t.Errorf(
			"Expected xs = %v, ys = %v, but got xs = %v, ys = %v.",
			xs, ys, poly.Xs, poly.Ys,
		)
	}

	tet.Translate(&[3]float32{0, 0, 10})
	pt.Init(&tet)
	if tet.ZPlaneSlice(pt, 0, poly) {
		t.Errorf("z=0 sliced a tetrahedron it did not intersect.")
	}
}

func vectorRing(n int) (vecs []AnchoredPluckerVec, phis []float32) {
	phis = make([]float32, n)
	vecs = make([]AnchoredPluckerVec, n)
	P := [3]float32{0, 0, 0}

	for i := 0; i < n; i++ {
		phi := float32(i) * math.Pi * 2 / float32(n)
		phis[i] = phi
		x := float32(math.Cos(float64(phi)))
		y := float32(math.Sin(float64(phi)))
		L := [3]float32{x, y, 0}
		vecs[i].Init(&P, &L)
	}

	return vecs, phis
}

func tetXs(t *Tetra) []float32 {
	xs := make([]float32, 4)
	for i := 0; i < 4; i++ {
		xs[i] = t[i][0]
	}
	return xs
}

func tetYs(t *Tetra) []float32 {
	ys := make([]float32, 4)
	for i := 0; i < 4; i++ {
		ys[i] = t[i][1]
	}
	return ys
}

func TestAngleRange(t *testing.T) {
	n, m := 100, 1000
	vecs, phis := vectorRing(n)
	tet := new(Tetra)
	poly := new(TetraSlice)
	pt := new(PluckerTetra)
	w := new(IntersectionWorkspace)

	for i := 0; i < m; i++ {
		randomizeTetra(tet, -1, +1)
		pt.Init(tet)

		if !tet.ZPlaneSlice(pt, 0, poly) {
			continue
		}
		low, width := poly.AngleRange()

		for j, phi := range phis {

			_, exit, ok := w.IntersectionDistance(pt, tet, &vecs[j])
			ok = ok && exit > 0
			inRange := AngleInRange(phi, low, width)
			if inRange && !ok {
				t.Errorf("Expected vec %d to be in range for tet %d.",
					j+1, i+1)
			} else if !inRange && ok {
				t.Errorf("Expected vec %d to not be in range for tet %d.",
					j+1, i+1)
			}
		}
	}
}

func BenchmarkSolveLine(b *testing.B) {
	y01, m1 := 2.0, -7.0
	y02, m2 := 5.0, 1.0
	for i := 0; i < b.N; i++ {
		x := (y02 - y01) / (m2 - m1)
		y := y01 + x*m1
		math.Sqrt(x*x + y*y)
	}
}

func BenchmarkSphereLineSegmentIntersection(b *testing.B) {
	n := 1000
	ls := make([]LineSegment, n)
	for i := range ls {
		ls[i] = LineSegment{[3]float32{float32(rand.Float64()),
			float32(rand.Float64()),
			float32(rand.Float64())}, [3]float32{0, 0, 1}, 0, 1}
	}
	// Want most of the lines to intersect.
	sphere := Sphere{[3]float32{0.5, 0.5, 0.5}, 0.5}

	b.ResetTimer()
	idx := 0
	for i := 0; i < b.N; i++ {
		sphere.LineSegmentIntersect(&ls[idx])
		idx++
		if idx == n {
			idx = 0
		}
	}
}
