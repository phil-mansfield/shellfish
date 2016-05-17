package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/phil-mansfield/shellfish/los/analyze"
)

const (
	samples = 1000*1000
)

func ellipsoid(a, b, c float64) analyze.Shell {
	return func(phi, theta float64) float64 {
		sp, cp := math.Sincos(phi)
		st, ct := math.Sincos(theta)

		return 1 / math.Sqrt((cp*cp*st*st/(a*a) +
		sp*sp*st*st/(b*b) + ct*ct/(c*c)))
	}
}

func formatGrid(nx, ny int, vals []float64) string {
	tokens := make([]string, len(vals))
	for i := range tokens { tokens[i] = fmt.Sprintf("%8.4g", vals[i]) }

	lines := make([]string, ny)
	for y := 0; y < ny; y++ {
		tLine := tokens[y*nx: (y+1)*nx]
		lines[y] = fmt.Sprintf("    %s,\n", strings.Join(tLine, ","))
	}

	return strings.Join(lines, "")
}

func main() {
	rand.Seed(time.Now().UnixNano())

	aLow, aHigh := 0.2, 1.0
	bLow, bHigh := 0.2, 1.0
	na, nb := 30, 30

	as, bs := make([]float64, na), make([]float64, nb)
	gas, gbs := make([]float64, na*nb), make([]float64, na*nb)
	cs := make([]float64, na*nb)

	for ia := 0; ia < na; ia++ {
		a := aLow + (aHigh - aLow)*float64(ia)/float64(na-1)
		as[ia] = a
		for ib := 0; ib < nb; ib++ {
			b := bLow + (bHigh - bLow)*float64(ib)/float64(nb-1)
			bs[ib] = b

			if ia > ib {
				cs[ia*nb + ib] = 0
				gas[ia*nb + ib] = 0
				gbs[ia*nb + ib] = 0
				continue
			}

			shell := ellipsoid(1, a, b)
			oc, ob, oa := shell.Axes(samples)



			cs[ia*nb + ib] = 1 / oc
			gas[ia*nb + ib] = a / (oa/oc)
			gbs[ia*nb + ib] = b / (ob/oc)
			_, _ = ob, oa
		}
	}

	fmt.Println("import numpy as np")

	fmt.Println("acs = np.array([")
	fmt.Print(formatGrid(na, 1, as))
	fmt.Println("])\n")

	fmt.Println("bcs = np.array([")
	fmt.Print(formatGrid(nb, 1, bs))
	fmt.Println("])\n")

	fmt.Println("ac_grid = np.array([")
	fmt.Print(formatGrid(nb, na, gas))
	fmt.Println("])\n")

	fmt.Println("bc_grid = np.array([")
	fmt.Print(formatGrid(nb, na, gbs))
	fmt.Println("])\n")

	fmt.Println("c_grid = np.array([")
	fmt.Print(formatGrid(nb, na, cs))
	fmt.Println("])\n")

	fmt.Printf("ac_grid = np.reshape(ac_grid, (%d, %d))\n", na, nb)
	fmt.Printf("bc_grid = np.reshape(bc_grid, (%d, %d))\n", na, nb)
	fmt.Printf("c_grid = np.reshape(c_grid, (%d, %d))\n", na, nb)
}
