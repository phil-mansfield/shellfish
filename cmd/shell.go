package cmd

import (
	"fmt"
	"math"
	"runtime"
	"sort"


	"github.com/phil-mansfield/shellfish/parse"
	"github.com/phil-mansfield/shellfish/cmd/env"
	"github.com/phil-mansfield/shellfish/cmd/catalog"
	"github.com/phil-mansfield/shellfish/cmd/memo"

	"github.com/phil-mansfield/shellfish/los"
	"github.com/phil-mansfield/shellfish/los/analyze"

	"github.com/phil-mansfield/shellfish/io"

	"github.com/phil-mansfield/shellfish/math/rand"

)

type ShellConfig struct {
	radialBins, spokes, rings int64
	rMaxMult, rMinMult float64
	rKernelMult float64

	eta float64
	order, smoothingWindow, levels, subsampleFactor int64
	losSlopeCutoff, backgroundRhoMult float64
}

var _ Mode = &ShellConfig{}

func (config *ShellConfig) ExampleConfig() string {
	return `[shell.config]

# Note: All fields in this file are optional. The value the field is set to
# is the default value. It is unlikely that the average user will want to
# change any of these values.
#
# If you find that the default parameters are behaving suboptimally in some
# significant way on halos with more than 50,000 particles, please submit a
# report on https://github.com/phil-mansfield/shellfish/issues.

# SubsampleFactor is the factor by which the points should be subsampled. This
# parameter is chiefly useful for convergence testing.
SubsampleFactor = 1

# RadialBins is the number of radial bins used for each line of sight.
RadialBins = 256

# Spokes is the number of lines of sight per ring.
Spokes = 256

# Rings is the number of rungs per halo.
Rings = 24

# RMaxMult is the maximum radius of a line of sight as a multiplier of R200m.
RMaxMult = 3.0

# RMaxMult is the minimum radius of a line of sight as a multiplier of R200m.
RMinMult = 0.5

# KernelRadiusMult is the radius of the spherical kernels around every
# particle as a multiplier of R200m.
RKernelMult = 0.2

# Eta is the tuning paramter of the point-filtering routines and determines the
# characteristic scale used in this step. Exactly what this entails is too
# complicated to be described here and can be found in the Shellfish paper.
Eta = 10.0

# Order indicates the order of the Penna function used to represent the
# splashback shell.
Order = 3

# Levels is the number of recursive angular splittings that should be done when
# filtering points.
Levels = 3

# SmoothingWindow is the width of the Savitzky-Golay smoothing window used
# when finding the point of steepest slope along lines of sight. Must be an odd
# number.
SmoothingWindow = 121

# Cutoff is the minimum slope allowed when finding the point of steepest slope
# for individual lines of sight. Of all the parameters that should not be
# changed, this is the one which should not be changed the most.
LOSSlopeCutoff = 0.0

# BackgroundRhoMult is the density assigned to points which do not intersect
# with any kernels as a multiple of the kernel density.
BackgroundRhoMult = 0.5`
}

func (config *ShellConfig) ReadConfig(fname string) error {
	vars := parse.NewConfigVars("shell.config")

	vars.Int(&config.subsampleFactor, "SubsampleFactor", 1)
	vars.Int(&config.radialBins, "RadialBins", 256)
	vars.Int(&config.spokes, "Spokes", 256)
	vars.Int(&config.rings, "Rings", 24)
	vars.Float(&config.rMaxMult, "RMaxMult", 3)
	vars.Float(&config.rMinMult, "RMinMult", 0.5)
	vars.Float(&config.rKernelMult, "RKernelMult", 0.2)
	vars.Float(&config.eta, "Eta", 10)
	vars.Int(&config.order, "Order", 3)
	vars.Int(&config.levels, "Levels", 3)
	vars.Int(&config.smoothingWindow, "SmoothingWindow", 121)
	vars.Float(&config.losSlopeCutoff, "LOSSlopeCutoff", 0.0)
	vars.Float(&config.backgroundRhoMult, "BackgroundRhoMult", 0.5)

	if fname == "" { return nil }
	if err := parse.ReadConfig(fname, vars); err != nil { return err }
	return config.validate()
}

func (config *ShellConfig) validate() error {
	switch {
	case config.subsampleFactor <= 0:
		return fmt.Errorf("The variable '%s' was set to %d.",
			"SubsampleFactor", config.subsampleFactor)
	case config.radialBins <= 0:
		return fmt.Errorf("The variable '%s' was set to %d.",
			"RadialBins", config.radialBins)
	case config.spokes <= 0:
		return fmt.Errorf("The variable '%s' was set to %d.",
			"Spokes", config.spokes)
	case config.rings <= 0:
		return fmt.Errorf("The variable '%s' was set to %d.",
			"Rings", config.rings)
	case config.rMaxMult <= 0:
		return fmt.Errorf("The variable '%s' was set to %g.",
			"RMaxMult", config.rMaxMult)
	case config.rMinMult <= 0:
		return fmt.Errorf("The variable '%s' was set to %g.",
			"RMinMult", config.rMinMult)
	case config.rKernelMult <= 0:
		return fmt.Errorf("The variable '%s' was set to %g.",
			"RKernelMult", config.rKernelMult)
	case config.eta <= 0:
		return fmt.Errorf("The variable '%s' was set to %g.",
			"Eta", config.eta)
	case config.order <= 0:
		return fmt.Errorf("The variable '%s' was set to %d.",
			"Order", config.order)
	case config.levels <= 0:
		return fmt.Errorf("The variable '%s' was set to %d.",
			"Levels", config.levels)
	case config.smoothingWindow <= 0:
		return fmt.Errorf("The variable '%s' was set to %d.",
			"SmoothingWindow", config.smoothingWindow)
	}

	if config.rMinMult >= config.rMaxMult {
		return fmt.Errorf("The variable '%s' was set to %g, but the " +
			"variable '%s' was set to %g.", "RMinMult", config.rMinMult,
			"RMaxMult", config.rMaxMult)
	}

	return nil
}

func (config *ShellConfig) Run(
	flags []string, gConfig *GlobalConfig, e *env.Environment, stdin []string,
) ([]string, error) {

	// Parse.
	intCols, coords, err := catalog.ParseCols(
		stdin, []int{0, 1}, []int{2, 3, 4, 5},
	)
	if err != nil { return nil, err }
	ids, snaps := intCols[0], intCols[1]
	
	if len(ids) == 0 { return nil, fmt.Errorf("No input IDs.") }

	// Compute coefficients.
	out := make([][]float64, len(ids))
	rowLength := config.order*config.order*2

	for i := range out {
		out[i] = make([]float64, rowLength)
	}

	err = loop(ids, snaps, coords, config, e, out)
	if err != nil { return nil, err }

	intNames := []string{"ID", "Snapshot"}
	floatNames := []string{"X", "Y", "Z", "R_200m"}
	for k := 0; k < 2; k++ {
		for j := 0; j < int(config.order); j++ {
			for i := 0; i < int(config.order); i++ {
				floatNames = append(
					floatNames, fmt.Sprintf("P_i=%d,j=%d,k=%d", i, j, k),
				)
			}
		}
	}


	colOrder := make([]int, 2 + 4 + 2*config.order*config.order)
	for i := range colOrder { colOrder[i] = i }

	lines := catalog.FormatCols(
		[][]int{ids, snaps}, append(coords, transpose(out)...), colOrder,
	)

	cString := catalog.CommentString(intNames, floatNames, colOrder)
	return append([]string{cString}, lines...), nil
}

func transpose(in [][]float64) [][]float64 {
	rows, cols := len(in), len(in[0])
	out := make([][]float64, cols)
	for i := range out { out[i] = make([]float64, rows) }

	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			out[x][y] = in[y][x]
		}
	}
	
	return out
}

func loop(
	ids, snaps []int, coords [][]float64,
	c *ShellConfig, e *env.Environment, out [][]float64,
) error {
	snapBins, idxBins := binBySnap(snaps, ids)
	ringBuf := make([]analyze.RingBuffer, c.rings)
	for i := range ringBuf { ringBuf[i].Init(int(c.spokes), int(c.radialBins)) }

	sortedSnaps := []int{}
	for snap := range snapBins {
		sortedSnaps = append(sortedSnaps, snap)
	}
	sort.Ints(sortedSnaps)

	hds, _, err := memo.ReadHeaders(sortedSnaps[0], e)
	if err != nil { return err }

	workers := runtime.NumCPU()
	sphBuf := &sphBuffers{
		intr: make([]bool, hds[0].N),
		vecs: make([][3]float32, hds[0].N),
		sphWorkers: make([]los.Halo, workers - 1),
	}

	for _, snap := range sortedSnaps {
		if snap == -1 { continue }
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

		// Create Halos
		runtime.GC()
		halos, err := createHalos(snapCoords, &hds[0], c, e)
		if err != nil { return err }

		if err = sphereLoop(snap, ids, idxs, halos, c, e, sphBuf, out);
			err != nil { return err }

		// Analysis
		if err = haloAnalysis(halos, idxs, c, ringBuf, out); err != nil {
			return err
		}
	}

	return nil
}

func sphereLoop(
	snap int, IDs, ids []int, halos []*los.Halo, c *ShellConfig,
	e *env.Environment, sphBuf *sphBuffers, out [][]float64,
) error {
	hds, files, err := memo.ReadHeaders(snap, e)
	if err != nil { return err }
	intrBins := binIntersections(hds, halos)
	
	for i := range hds {
		runtime.GC()
		if len(intrBins[i]) == 0 { continue }
		
		err := io.ReadSheetPositionsAt(files[i], sphBuf.vecs)
		if err != nil { return err }

		binHs := intrBins[i]
		for j := range binHs {
			loadSphereVecs(binHs[j], sphBuf, &hds[i], c)
		}

	}

	return nil
}

type sphBuffers struct {
	sphWorkers []los.Halo
	vecs [][3]float32
	intr []bool
}

func loadSphereVecs(
	h *los.Halo, sphBuf *sphBuffers, hd *io.SheetHeader, c *ShellConfig,
) {
	workers := runtime.NumCPU()
	runtime.GOMAXPROCS(workers)
	sphWorkers, vecs, intr := sphBuf.sphWorkers, sphBuf.vecs, sphBuf.intr
	if len(sphWorkers) + 1 != workers { panic("impossible")}

	sync := make(chan bool, workers)

	h.Transform(vecs, hd.TotalWidth)
	rad := h.RMax() * c.rKernelMult / c.rMaxMult
	h.Intersect(vecs, rad, intr)
	numIntr := 0
	for i := range intr {
		if intr[i] { numIntr++ }
	}

	h.Split(sphWorkers)

	for i := range sphWorkers {
		wh := &sphBuf.sphWorkers[i]
		go chanLoadSphereVec(wh, vecs, intr, i, workers, hd, c, sync)
	}
	chanLoadSphereVec(h, vecs, intr, workers - 1, workers, hd, c, sync)

	for i := 0; i < workers; i++ { <-sync }

	h.Join(sphWorkers)
}

func chanLoadSphereVec(
	h *los.Halo, vecs [][3]float32, intr []bool,
	offset, workers int,
	hd *io.SheetHeader, c *ShellConfig, sync chan bool,
) {

	rad := h.RMax() * c.rKernelMult / c.rMaxMult
	sphVol := 4*math.Pi/3*rad*rad*rad

	sf := c.subsampleFactor
	pl := hd.TotalWidth / float64(hd.CountWidth / sf)
	pVol := pl*pl*pl
	
	rho := pVol / sphVol

	skip := workers*int(sf*sf*sf)
	for i := offset*int(sf*sf*sf); i < int(hd.N); i += skip {
		if intr[i] { h.Insert(vecs[i], rad, rho) }
	}

	sync <- true
}

func haloAnalysis(
	halos []*los.Halo, idxs []int, c *ShellConfig,
	ringBuf []analyze.RingBuffer, out [][]float64,
) error {
	// Calculate Penna coefficients.
	for i := range halos {
		runtime.GC()

		var ok bool
		out[idxs[i]], ok = calcCoeffs(halos[i], ringBuf, c)
		if !ok {
			fmt.Errorf("Shell coefficients undetermined. The most likely " +
				"explanation is that there is corruption in your particle " +
				"snapshots.")
		}
	}
	return nil
}

func createHalos(
	coords [][]float64, hd *io.SheetHeader, c *ShellConfig, e *env.Environment,
) ([]*los.Halo, error) {
	halos := make([]*los.Halo, len(coords[0]))
	for i, _ := range coords[0] {
		x, y, z, r := coords[0][i], coords[1][i], coords[2][i], coords[3][i]

		// This happens sometimes...
		if r <= 0 { continue }

		norms := normVecs(int(c.rings))
		origin := [3]float64{x, y, z}
		rMax, rMin := r*c.rMaxMult, r*c.rMinMult
		rad := r*c.rKernelMult

		sphVol := 4*math.Pi/3*rad*rad*rad
		pl := hd.TotalWidth/float64(int(hd.CountWidth)/int(c.subsampleFactor))
		pVol := pl*pl*pl
		rho := pVol / sphVol
		defaultRho := rho * c.backgroundRhoMult

		halo := &los.Halo{}
		halo.Init(norms, origin, rMin, rMax, int(c.radialBins),
			int(c.spokes), defaultRho)

		halos[i] = halo
	}

	return halos, nil
}

func normVecs(n int) [][3]float32 {
	var vecs [][3]float32
	gen := rand.NewTimeSeed(rand.Xorshift)
	switch n {
	case 3:
		vecs = [][3]float32{{0, 0, 1}, {0, 1, 0}, {1, 0, 0}}
	default:
		vecs = make([][3]float32, n)
		for i := range vecs {
			for {
				x := gen.Uniform(-1, +1)
				y := gen.Uniform(-1, +1)
				z := gen.Uniform(-1, +1)
				r := math.Sqrt(x*x + y*y + z*z)

				if r < 1 {
					vecs[i] = [3]float32{
						float32(x/r), float32(y/r), float32(z/r),
					}
					break
				}
			}
		}
	}

	return vecs
}

type profileRange struct {
	rMin, rMax float64
	v0 [3]float32
}

func binIntersections(
	hds []io.SheetHeader, halos []*los.Halo,
) [][]*los.Halo {
	bins := make([][]*los.Halo, len(hds))
	for i := range hds {
		for hi := range halos {
			if halos[hi].SheetIntersect(&hds[i]) {
				bins[i] = append(bins[i], halos[hi])
			}
		}
	}
	return bins
}

func calcCoeffs(
	halo *los.Halo, buf []analyze.RingBuffer, c *ShellConfig,
) ([]float64, bool) {
	for i := range buf {
		buf[i].Clear()
		buf[i].Splashback(halo, i, int(c.smoothingWindow), c.losSlopeCutoff)
	}
	pxs, pys, ok := analyze.FilterPoints(buf, int(c.levels), c.eta)
	if !ok { return nil, false }
	cs, _ := analyze.PennaVolumeFit(pxs, pys, halo, int(c.order), int(c.order))
	return cs, true
}

func binBySnap(snaps, ids []int) (snapBins, idxBins map[int][]int) {
	snapBins = make(map[int][]int)
	idxBins = make(map[int][]int)
	for i, snap := range snaps {
		id := ids[i]
		snapBins[snap] = append(snapBins[snap], id)
		idxBins[snap] = append(idxBins[snap], i)
	}
	return snapBins, idxBins
}
