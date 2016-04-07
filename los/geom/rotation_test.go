package geom

import (
	"math"
	"math/rand"
	"testing"
)

func vecEpsEq(v1, v2 *[3]float32, eps float32) bool {
	for i := 0; i < 3; i++ {
		diff := v1[i] - v2[i]
		if diff > eps || diff < -eps {
			return false
		}
	}
	return true
}

func TestRotate(t *testing.T) {
	eps := float32(1e-4)
	pi := float32(math.Pi)
	pi2 := float32(math.Pi/2)
	sqrt2 := 1/float32(math.Sqrt(2))

	table := []struct{
		phi, theta, psi float32
		start, end [3]float32
	} {
		{0, 0, 0, [3]float32{1, 2, 3}, [3]float32{1, 2, 3}},
		{pi, 0, 0, [3]float32{1, 0, 0}, [3]float32{-1, 0, 0}},
		{-pi, 0, 0, [3]float32{1, 0, 0}, [3]float32{-1, 0, 0}},
		{0, 0, pi, [3]float32{1, 0, 0}, [3]float32{-1, 0, 0}},
		{0, 0, -pi, [3]float32{1, 0, 0}, [3]float32{-1, 0, 0}},
		{0, pi, 0, [3]float32{0, 1, 0}, [3]float32{0, -1, 0}},
		{0, -pi, 0, [3]float32{0, 1, 0}, [3]float32{0, -1, 0}},

		{pi2, 0, 0, [3]float32{1, 0, 0}, [3]float32{0, 1, 0}},
		{pi2, 0, 0, [3]float32{0, 1, 0}, [3]float32{-1, 0, 0}},
		{0, 0, pi2, [3]float32{1, 0, 0}, [3]float32{0, 1, 0}},
		{0, 0, pi2, [3]float32{0, 1, 0}, [3]float32{-1, 0, 0}},
		{0, pi2, 0, [3]float32{0, 1, 0}, [3]float32{0, 0, 1}},
		{0, pi2, 0, [3]float32{0, 0, 1}, [3]float32{0, -1, 0}},

		{pi2, pi2/2, 0, [3]float32{1, 0, 0}, [3]float32{0, sqrt2, sqrt2}},
	}

	for i, test := range table {
		m := EulerMatrix(test.phi, test.theta, test.psi)
		v := test.start
		RotateVec(v, m)
		if !vecEpsEq(&v, &test.end, eps) {
			t.Errorf(
				"%d) %v.Rotate(%.4g %.4g %.4g) -> %v instead of %v",
				i+1, test.start, test.phi, test.theta, test.psi, v, test.end,
			)
		}
	}
}

func TestEulerMatrixBetween(t *testing.T) {
	sqrt2 := 1/float32(math.Sqrt(2))

	eps := float32(1e-4)
	table := []struct{
		v1, v2 [3]float32
	} {
		{[3]float32{1, 0, 0}, [3]float32{1, 0, 0}},
		{[3]float32{1, 0, 0}, [3]float32{0, 1, 0}},
		{[3]float32{0, 1, 0}, [3]float32{0, sqrt2, sqrt2}},
		{[3]float32{1, 0, 0}, [3]float32{0, sqrt2, sqrt2}},
 		{[3]float32{1, 0, 0}, [3]float32{0, 0, 1}},
		{[3]float32{0, 1, 0}, [3]float32{1, 0, 0}},
		{[3]float32{0, 1, 0}, [3]float32{0, 0, 1}},
		{[3]float32{0, 0, 1}, [3]float32{1, 0, 0}},


		{[3]float32{2, 3, 4}, [3]float32{4, 2, 3}},
		{[3]float32{2, 3, 4}, [3]float32{4, 2, 3}},
		{[3]float32{2, 3, 4}, [3]float32{4, 2, 3}},
	}

	for i, test := range table {
		m := EulerMatrixBetween(&test.v1, &test.v2)
		v := test.v1
		RotateVec(v, m)
		if !vecEpsEq(&v, &test.v2, eps) {
			t.Errorf(
				"%d) %v.Rotate(EulerMatrixBetween(%v %v)) -> %v instead of %v",
				i+1, test.v1, test.v1, test.v2, v, test.v2,
			)
		}
	}
}


func TestAngularDistance(t *testing.T) {
	n := 1000
	for i := 0; i < n; i++ {
		phi1, phi2 := rand.Float32() * 2 * math.Pi, rand.Float32() * 2 * math.Pi
		dist := AngularDistance(phi1, phi2)
		
		if dist > math.Pi || dist < -math.Pi {
			t.Errorf(
				"%d) AnguleDistance(%g, %g) -> %g out of range [-pi, pi].",
				i + 1, phi1, phi2, dist,
			)
		} else if !almostEq(
			float32(math.Mod(float64(dist+phi1+2*math.Pi), 2*math.Pi)),
			phi2,
			1e-4) {

			t.Errorf(
				"%d) AngularDistance(%g, %g) -> %g doesn't add up.",
				i + 1, phi1, phi2, dist,
			)
		}
	}
}

func BenchmarkVecRotate(b *testing.B) {
	v := [3]float32{1, 1, 1}
	m := EulerMatrix(1, 2, 3)
	for i := 0; i < b.N; i++ {
		RotateVec(v, m)
	}
}

func BenchmarkTetraRotate(b *testing.B) {
	t := Tetra{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}, {1, 1, 1}}
	m := EulerMatrix(1, 2, 3)
	for i := 0; i < b.N; i++ {
		t.Rotate(m)
	}
}

func BenchmarkEulerMatrix(b *testing.B) {
	n := 1000
	phis := make([]float32, n)
	thetas := make([]float32, n)
	psis := make([]float32, n)
	for i := 0; i < n; i++ {
		phis[i] = rand.Float32()
		thetas[i] = rand.Float32()
		psis[i] = rand.Float32()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EulerMatrix(phis[i % n], thetas[i % n], psis[i % n])
	}
}
