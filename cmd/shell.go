package cmd

import (
	"fmt"
	"log"
	"math"
	"runtime"
	"sort"

	"github.com/phil-mansfield/shellfish/cmd/catalog"
	"github.com/phil-mansfield/shellfish/cmd/env"
	"github.com/phil-mansfield/shellfish/cmd/memo"
	"github.com/phil-mansfield/shellfish/cosmo"
	"github.com/phil-mansfield/shellfish/los"
	"github.com/phil-mansfield/shellfish/los/analyze"
	"github.com/phil-mansfield/shellfish/parse"
	"github.com/phil-mansfield/shellfish/io"
	"github.com/phil-mansfield/shellfish/math/rand"
	"github.com/phil-mansfield/shellfish/logging"
)

type ShellConfig struct {
	radialBins, spokes, rings int64
	rMaxMult, rMinMult        float64
	rKernelMult               float64

	eta                                             float64
	order, smoothingWindow, levels, subsampleFactor int64
	losSlopeCutoff, backgroundRhoMult               float64
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

# Rings is the number of rings per halo.
Rings = 100

# RMaxMult is the maximum radius of a line of sight as a multiplier of R200m.
RMaxMult = 3.0

# RMinMult is the minimum radius of a line of sight as a multiplier of R200m.
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
	vars.Int(&config.rings, "Rings", 100)
	vars.Float(&config.rMaxMult, "RMaxMult", 3)
	vars.Float(&config.rMinMult, "RMinMult", 0.5)
	vars.Float(&config.rKernelMult, "RKernelMult", 0.2)
	vars.Float(&config.eta, "Eta", 10)
	vars.Int(&config.order, "Order", 3)
	vars.Int(&config.levels, "Levels", 3)
	vars.Int(&config.smoothingWindow, "SmoothingWindow", 121)
	vars.Float(&config.losSlopeCutoff, "LOSSlopeCutoff", 0.0)
	vars.Float(&config.backgroundRhoMult, "BackgroundRhoMult", 0.5)

	if fname == "" {
		return nil
	}
	if err := parse.ReadConfig(fname, vars); err != nil {
		return err
	}
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
		return fmt.Errorf("The variable '%s' was set to %g, but the "+
			"variable '%s' was set to %g.", "RMinMult", config.rMinMult,
			"RMaxMult", config.rMaxMult)
	}

	return nil
}

func (config *ShellConfig) Run(
	flags []string, gConfig *GlobalConfig, e *env.Environment, stdin []string,
) ([]string, error) {

	if logging.Mode != logging.Nil {
		log.Println(`
#####################
## shellfish shell ##
#####################`,
		)
	}

	// Parse.
	intCols, coords, err := catalog.ParseCols(
		stdin, []int{0, 1}, []int{2, 3, 4, 5},
	)
	if err != nil {
		return nil, err
	}
	ids, snaps := intCols[0], intCols[1]

	if len(ids) == 0 {
		return nil, fmt.Errorf("No input IDs.")
	}

	// Compute coefficients.
	out := make([][]float64, len(ids))
	rowLength := config.order * config.order * 2

	for i := range out {
		out[i] = make([]float64, rowLength)
	}

	buf, err := getVectorBuffer(
		e.ParticleCatalog(snaps[0], 0),
		gConfig.SnapshotType, gConfig.Endianness,
	)

	err = loop(ids, snaps, coords, config, buf, e, out, gConfig.Threads)
	if err != nil {
		return nil, err
	}

	intNames := []string{"ID", "Snapshot"}
	floatNames := []string{"X", "Y", "Z", "R200m", "P_ijk"}

	colOrder := make([]int, 2+4+2*config.order*config.order)
	for i := range colOrder {
		colOrder[i] = i
	}

	lines := catalog.FormatCols(
		[][]int{ids, snaps}, append(coords, transpose(out)...), colOrder,
	)

	cString := catalog.CommentString(
		intNames, floatNames, []int{0, 1, 2, 3, 4, 5, 6},
		[]int{1, 1, 1, 1, 1, 1, 2*int(config.order*config.order)},
	)
	return append([]string{cString}, lines...), nil
}

func transpose(in [][]float64) [][]float64 {
	rows, cols := len(in), len(in[0])
	out := make([][]float64, cols)
	for i := range out {
		out[i] = make([]float64, rows)
	}

	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			out[x][y] = in[y][x]
		}
	}

	return out
}

func loop(
	ids, snaps []int, coords [][]float64, c *ShellConfig,
	buf io.VectorBuffer, e *env.Environment, out [][]float64,
	threads int64,
) error {
	snapBins, idxBins := binBySnap(snaps, ids)
	ringBuf := make([]analyze.RingBuffer, c.rings)
	for i := range ringBuf {
		ringBuf[i].Init(int(c.spokes), int(c.radialBins))
	}

	sortedSnaps := []int{}
	for snap := range snapBins {
		sortedSnaps = append(sortedSnaps, snap)
	}
	sort.Ints(sortedSnaps)

	hds, _, err := memo.ReadHeaders(sortedSnaps[0], buf, e)
	if err != nil {
		return err
	}
	minMass := buf.MinMass()

	workers := runtime.NumCPU()
	if threads > 0 {
		workers = int(threads)
	}
	sphBuf := &sphBuffers{
		intr:       make([]bool, hds[0].N),
		xs:         [][3]float32{},
		ms:         []float32{},
		sphWorkers: make([]los.Halo, workers-1),
	}

	for _, snap := range sortedSnaps {
		if snap == -1 {
			continue
		}
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
		halos, err := createHalos(snapCoords, &hds[0], c, e, minMass)
		if err != nil {
			return err
		}

		// I'm so sorry about having ten arguments to this function.
		if err = sphereLoop(snap, ids, idxs, halos, c,
			buf, e, sphBuf, threads, out); err != nil {

			return err
		}

		// Analysis
		if err = haloAnalysis(halos, idxs, c, ringBuf, out); err != nil {
			return err
		}
	}

	return nil
}

// TODO: Refactor this monstrosity of a call signature.

func sphereLoop(
	snap int, IDs, ids []int, halos []*los.Halo, c *ShellConfig,
	buf io.VectorBuffer, e *env.Environment, sphBuf *sphBuffers,
	threads int64, out [][]float64,
) error {
	hds, files, err := memo.ReadHeaders(snap, buf, e)
	if err != nil {
		return err
	}
	intrBins := binIntersections(hds, halos)

	for i := range hds {
		runtime.GC()
		if len(intrBins[i]) == 0 {
			continue
		}

		sphBuf.xs, sphBuf.ms, err = buf.Read(files[i])
		if err != nil {
			return err
		}

		binHs := intrBins[i]
		for j := range binHs {
			loadSphereVecs(binHs[j], sphBuf, &hds[i], c, threads)
		}

		buf.Close()
	}

	return nil
}

type sphBuffers struct {
	sphWorkers []los.Halo
	xs         [][3]float32
	ms         []float32
	intr       []bool
}

func loadSphereVecs(
	h *los.Halo, sphBuf *sphBuffers, hd *io.Header, c *ShellConfig,
	threads int64,
) {
	workers := runtime.NumCPU()
	if threads > 0 {
		workers = int(threads)
	}
	runtime.GOMAXPROCS(workers)
	sphWorkers, xs := sphBuf.sphWorkers, sphBuf.xs
	sphBuf.intr = expandBools(sphBuf.intr[:0], len(xs))
	ms, intr := sphBuf.ms, sphBuf.intr
	if len(sphWorkers)+1 != workers {
		panic("impossible")
	}

	sync := make(chan bool, workers)

	h.Transform(xs, hd.TotalWidth)
	rad := h.RMax() * c.rKernelMult / c.rMaxMult
	h.Intersect(xs, rad, intr)
	numIntr := 0
	for i := range intr {
		if intr[i] {
			numIntr++
		}
	}

	h.Split(sphWorkers)
	
	for i := range sphWorkers {
		wh := &sphBuf.sphWorkers[i]
		go chanLoadSphereVec(wh, xs, ms, intr, i, workers, hd, c, sync)
	}
	chanLoadSphereVec(h, xs, ms, intr, workers-1, workers, hd, c, sync)

	for i := 0; i < workers; i++ {
		<-sync
	}

	h.Join(sphWorkers)
}

func expandBools(scalars []bool, n int) []bool {
	switch {
	case cap(scalars) >= n:
		return scalars[:n]
	case int(float64(cap(scalars))*1.5) > n:
		return append(scalars[:cap(scalars)],
			make([]bool, n-cap(scalars))...)
	default:
		return make([]bool, n)
	}
}

func chanLoadSphereVec(
	h *los.Halo, xs [][3]float32, ms []float32,
	intr []bool, offset, workers int,
	hd *io.Header, c *ShellConfig, sync chan bool,
) {
	rad := h.RMax() * c.rKernelMult / c.rMaxMult
	sphVol := 4 * math.Pi / 3 * rad * rad * rad

	rhoM := cosmo.RhoAverage(hd.Cosmo.H100*100,
		hd.Cosmo.OmegaM, hd.Cosmo.OmegaL, hd.Cosmo.Z)

	sf := c.subsampleFactor
	skip := workers * int(sf*sf*sf)
	for i := offset * int(sf*sf*sf); i < int(hd.N); i += skip {
		if intr[i] {
			h.Insert(xs[i], rad, (float64(ms[i])*float64(sf*sf*sf)/
				sphVol)/rhoM)
		}
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

		if logging.Mode == logging.Debug {
			log.Printf("Halo %3d", i)
		}
		
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
	coords [][]float64, hd *io.Header, c *ShellConfig, e *env.Environment,
	minMass float32,
) ([]*los.Halo, error) {

	halos := make([]*los.Halo, len(coords[0]))
	for i, _ := range coords[0] {
		x, y, z, r := coords[0][i], coords[1][i], coords[2][i], coords[3][i]

		// This happens sometimes...
		if r <= 0 {
			continue
		}

		norms := normVecs(int(c.rings))
		origin := [3]float64{x, y, z}
		rMax, rMin := r*c.rMaxMult, r*c.rMinMult
		rad := r * c.rKernelMult

		sphVol := 4 * math.Pi / 3 * rad * rad * rad
		sf := c.subsampleFactor
		rhoM := cosmo.RhoAverage(hd.Cosmo.H100*100,
			hd.Cosmo.OmegaM, hd.Cosmo.OmegaL, hd.Cosmo.Z)
		rho := ((float64(minMass) * float64(sf*sf*sf)) / sphVol) / rhoM
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
						float32(x / r), float32(y / r), float32(z / r),
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
	v0         [3]float32
}

func binIntersections(
	hds []io.Header, halos []*los.Halo,
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
	if !ok {
		return nil, false
	}
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
