package cmd

import (
	"fmt"
	"math"
	"runtime"
	"sort"

	"github.com/phil-mansfield/shellfish/cosmo"
	"github.com/phil-mansfield/shellfish/io"
	"github.com/phil-mansfield/shellfish/los/geom"
	"github.com/phil-mansfield/shellfish/los/analyze"

	"github.com/phil-mansfield/shellfish/parse"
	"github.com/phil-mansfield/shellfish/cmd/catalog"
	"github.com/phil-mansfield/shellfish/cmd/memo"
	"github.com/phil-mansfield/shellfish/cmd/env"
)


type StatsConfig struct {
	values []string
	histogramBins int64
	monteCarloSamples int64
	exclusionStrategy string
	order int64
}

var _ Mode = &StatsConfig{}

func (config *StatsConfig) ExampleConfig() string {
	return `[stats.config]

#####################
## Required Fields ##
#####################

# Values determines what columns will be written to stdout. If one of the
# elements of the list corresponds to a histogram, then HistogramBins x values
# will be written starting at that column, HistogramBins y values will be
# written after that, and the specified columns will continue from there.
#
# The supported columns are:
# id       - The ID of the halo, as initially supplied.
# snap     - The snapshot index of the halo, as initially supplied.
# r-sp     - The volume-weighted splashback radius of the halo.
# m-sp     - The total mass contained within the splashback shell of the halo.
# r-sp-max - The maximum radius of the splashback shell.
# r-sp-min - The minimum radius of the splashback shell.
Values = snap, id, r-sp

#####################
## Optional Fields ##
#####################

# HistogramBins is the number of bins to use for histogramed quantities.
HistogramBins = 50

# MonteCarloSamplings The number of Monte Carlo samplings done when calculating
# properties of shells.
MonteCarloSamples = 10000

# ExclustionStrategy is the strategy for removing halos contained within a
# larger halo's splashback shell.
#
# The supported strategies are:
# none    - Don't try to do this.
# contain - Only halos which have a center inside a larger halo's splashback are
#           excluded.
# overlap - Halos which have a splashback shell that overlaps the splashback
#           shell of a larger halo are excluded.
#
# The default value is none.
ExclusionStrategy = none

# Order is the order of the Penna shell constructed around the halos. It must be
# the same value used by the shell.config file. By default both are set to 3.
Order = 3
`
}

func (config *StatsConfig) ReadConfig(fname string) error {
	vars := parse.NewConfigVars("stats.config")

	vars.Strings(&config.values, "Values", []string{})
	vars.Int(&config.histogramBins, "HistogramBins", 50)
	vars.Int(&config.monteCarloSamples, "MonteCarloSamples", 10 * 1000)
	vars.String(&config.exclusionStrategy, "ExclusionStrategy", "none")
	vars.Int(&config.order, "Order", 3)

	if fname == "" { return nil }
	if err := parse.ReadConfig(fname, vars); err != nil { return err }
	return config.validate()
}

func (config *StatsConfig) validate() error {
	for i, val := range config.values {
		switch val {
		case "snap", "id", "m-sp", "r-sp", "r-sp-min", "r-sp-max":
		default:
			return fmt.Errorf("Item %d of variable 'Values' is set to '%s', " +
				"which I don't recognize.", i, val)
		}
	}

	switch config.exclusionStrategy {
	case "none", "contain", "overlap":
	default:
		return fmt.Errorf("variable 'ExclusionStrategy' set to '%s', which " +
			"I don't recognize.", config.exclusionStrategy)
	}

	switch {
	case config.histogramBins <= 0:
		return fmt.Errorf("The variable '%s' was set to %g",
			"HistogramBins", config.histogramBins)
	case config.monteCarloSamples <= 0:
		return fmt.Errorf("The variable '%s' was set to %g",
			"MonteCarloSamples", config.monteCarloSamples)
	}

	return nil
}

func (config *StatsConfig) Run(
	flags []string, gConfig *GlobalConfig, e *env.Environment, stdin []string,
) ([]string, error) {

	intColIdxs := []int{0, 1}
	floatColIdxs := make([]int, 4 + 2*config.order*config.order)
	for i := range floatColIdxs { floatColIdxs[i] = i + len(intColIdxs) }
	intCols, floatCols, err := catalog.ParseCols(
		stdin, intColIdxs, floatColIdxs,
	)

	if err != nil { return nil, err }
	ids, snaps := intCols[0], intCols[1]
	coords, coeffs := floatCols[:4], transpose(floatCols[4:])
	
	snapBins, coeffBins, idxBins := binCoeffsBySnap(snaps, ids, coeffs)

	masses := make([]float64, len(ids))
	rads := make([]float64, len(ids))
	rmins := make([]float64, len(ids))
	rmaxes := make([]float64, len(ids))

	sortedSnaps := []int{}
	for snap := range snapBins {
		sortedSnaps = append(sortedSnaps, snap)
	}
	sort.Ints(sortedSnaps)

	for _, snap := range sortedSnaps {
		if snap == -1 { continue }
		snapCoeffs := coeffBins[snap]
		idxs := idxBins[snap]

		snapCoords := [][]float64{
			make([]float64, len(idxs)), make([]float64, len(idxs)),
			make([]float64, len(idxs)), make([]float64, len(idxs)),
		}
		for i, idx := range idxs {
			snapCoords[0][i] = coords[0][idx]
			snapCoords[1][i] = coords[1][idx]
			snapCoords[2][i] = coords[2][idx]
			snapCoords[3][i] = coords[3][idx]
		}

		for j := range idxs {
			rads[idxs[j]] = rSp(snapCoeffs[j])
			rmins[idxs[j]], rmaxes[idxs[j]] = rangeSp(snapCoeffs[j])
		}

		hds, files, err := memo.ReadHeaders(snap, e)
		if err != nil { return nil, err }
		hBounds, err := boundingSpheres(snapCoords, &hds[0], config, e)
		if err != nil { return nil, err }
		intrBins := binSphereIntersections(hds, hBounds)

		rLows := make([]float64, len(snapCoeffs))
		rHighs := make([]float64, len(snapCoeffs))
		for i := range snapCoeffs {
			order := findOrder(snapCoeffs[i])
			shell := analyze.PennaFunc(snapCoeffs[i], order, order, 2)
			rLows[i], rHighs[i] = shell.RadialRange(10 * 1000)
		}

		buf, err := io.NewGotetraBuffer(files[0])
		if err != nil { return nil, err }

		for i := range hds {
			if len(intrBins[i]) == 0 { continue }

			xs, err := buf.Read(files[i])
			if err != nil { return nil, err }

			for j := range idxs {
				masses[idxs[j]] += massContained(
					&hds[i], xs, snapCoeffs[j],
					hBounds[j], rLows[j], rHighs[j],
				)
			}
			
			buf.Close()
		}
	}

	lines := catalog.FormatCols(
		[][]int{ids, snaps},
		[][]float64{masses, rads, rmins, rmaxes},
		[]int{0, 1, 2, 3, 4, 5},
	)
	cString := catalog.CommentString(
		[]string{"ID", "Snapshot"},
		[]string{"M_sp [M_sun/h]", "R_sp [Mpc/h]",
			"R_sp,min [Mpc/h]", "R_sp,max [Mpc/h]"},
		[]int{0, 1, 2, 3, 4, 5},
	)

	return append([]string{cString}, lines...), nil
}

func wrapDist(x1, x2, width float32) float32 {
	dist := x1 - x2
	if dist > width / 2 {
		return dist - width
	} else if dist < width / -2 {
		return dist + width
	} else {
		return dist
	}
}

func inRange(x, r, low, width, tw float32) bool {
	return wrapDist(x, low, tw) > -r && wrapDist(x, low + width, tw) < r
}

// SheetIntersect returns true if the given halo and sheet intersect one another
// and false otherwise.
func sheetIntersect(s geom.Sphere, hd *io.SheetHeader) bool {
	tw := float32(hd.TotalWidth)
	return inRange(s.C[0], s.R, hd.Origin[0], hd.Width[0], tw) &&
	inRange(s.C[1], s.R, hd.Origin[1], hd.Width[1], tw) &&
	inRange(s.C[2], s.R, hd.Origin[2], hd.Width[2], tw)
}


func binCoeffsBySnap(
	snaps, ids []int, coeffs [][]float64,
) (snapBins map[int][]int,coeffBins map[int][][]float64,idxBins map[int][]int) {
	snapBins = make(map[int][]int)
	coeffBins = make(map[int][][]float64)
	idxBins = make(map[int][]int)
	for i, snap := range snaps {
		snapBins[snap] = append(snapBins[snap], ids[i])
		coeffBins[snap] = append(coeffBins[snap], coeffs[i])
		idxBins[snap] = append(idxBins[snap], i)
	}
	return snapBins, coeffBins, idxBins
}

func boundingSpheres(
	coords [][]float64, hd *io.SheetHeader, c *StatsConfig, e *env.Environment,
) ([]geom.Sphere, error) {
	xs, ys, zs, rs := coords[0], coords[1], coords[2], coords[3]

	spheres := make([]geom.Sphere, len(coords[0]))
	for i := range spheres {
		spheres[i].C = [3]float32{
			float32(xs[i]), float32(ys[i]), float32(zs[i]),
		}
		spheres[i].R = float32(rs[i])
	}

	return spheres, nil
}

func findOrder(coeffs []float64) int {
	i := 1
	for {
		if 2*i*i == len(coeffs) {
			return i
		} else if 2*i*i > len(coeffs) {
			panic("Impossible")
		}
		i++
	}
}

func wrap(x, tw2 float32) float32 {
	if x > tw2 {
		return x - tw2
	} else if x < -tw2 {
		return x + tw2
	}
	return x
}

func coords(idx, cells int64) (x, y, z int64) {
	x = idx % cells
	y = (idx % (cells * cells)) / cells
	z = idx / (cells * cells)
	return x, y, z
}

func rSp(coeffs []float64) float64 {
	order := findOrder(coeffs)
	shell := analyze.PennaFunc(coeffs, order, order, 2)
	vol := shell.Volume(10 * 1000)
	r := math.Pow(vol / (math.Pi * 4 / 3), 0.33333)
	return r
	//return shell.MedianRadius(10 * 1000)
}

func rangeSp(coeffs []float64) (rmin, rmax float64) {
	order := findOrder(coeffs)
	shell := analyze.PennaFunc(coeffs, order, order, 2)
	return shell.RadialRange(10 * 1000)
}

func massContained(
	hd *io.SheetHeader, xs [][3]float32, coeffs []float64,
	sphere geom.Sphere, rLow, rHigh float64,
) float64 {

	cpu := runtime.NumCPU()
	workers := int64(runtime.GOMAXPROCS(cpu))
	outChan := make(chan float64, workers)
	for i := int64(0); i < workers - 1; i++ {
		go massContainedChan(
			hd, xs, coeffs, sphere, rLow, rHigh, i, workers, outChan,
		)
	}

	massContainedChan(
		hd, xs, coeffs, sphere, rLow, rHigh, workers - 1, workers, outChan,
	)

	sum := 0.0
	for i := int64(0); i < workers; i++ {
		sum += <-outChan
	}

	return sum
}

func massContainedChan(
	hd *io.SheetHeader, xs [][3]float32, coeffs []float64,
	sphere geom.Sphere, rLow, rHigh float64,
	offset, workers int64, out chan float64,
) {
	c := &hd.Cosmo
	rhoM := cosmo.RhoAverage(c.H100 * 100, c.OmegaM, c.OmegaL, c.Z )
	dx := hd.TotalWidth / float64(hd.CountWidth) / (1 + c.Z)
	ptMass := rhoM * (dx*dx*dx)
	tw2 := float32(hd.TotalWidth) / 2

	order := findOrder(coeffs)
	shell := analyze.PennaFunc(coeffs, order, order, 2)
	low2, high2 := float32(rLow*rLow), float32(rHigh*rHigh)

	sum := 0.0
	for i := offset; i < hd.N; i += workers {
		x, y, z := xs[i][0], xs[i][1], xs[i][2]
		x, y, z = x - sphere.C[0], y - sphere.C[1], z - sphere.C[2]
		x = wrap(x, tw2)
		y = wrap(y, tw2)
		z = wrap(z, tw2)

		r2 := x*x + y*y +z*z

		if r2 < low2 || ( r2 < high2 &&
		shell.Contains(float64(x), float64(y), float64(z))) {
			sum += ptMass
		}
	}
	out <- sum
}

func binSphereIntersections(
	hds []io.SheetHeader, spheres []geom.Sphere,
) [][]geom.Sphere {
	bins := make([][]geom.Sphere, len(hds))
	for i := range hds {
		for si := range spheres {
			if sheetIntersect(spheres[si], &hds[i]) {
				bins[i] = append(bins[i], spheres[si])
			}
		}
	}
	return bins
}
