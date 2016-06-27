package cmd

import (
	"fmt"
	"log"
	"math"
	"runtime"
	"sort"
	"time"

	"github.com/phil-mansfield/shellfish/io"
	"github.com/phil-mansfield/shellfish/los/analyze"
	"github.com/phil-mansfield/shellfish/los/geom"

	"github.com/phil-mansfield/shellfish/cmd/catalog"
	"github.com/phil-mansfield/shellfish/cmd/env"
	"github.com/phil-mansfield/shellfish/cmd/memo"
	"github.com/phil-mansfield/shellfish/parse"
	"github.com/phil-mansfield/shellfish/logging"
)

type StatsConfig struct {
	values            []string
	monteCarloSamples int64
	exclusionStrategy string
	order             int64

	shellFilter bool
	shellParticleFile string
	shellWidth float64
}

var _ Mode = &StatsConfig{}

func (config *StatsConfig) ExampleConfig() string {
	return `[stats.config]

#####################
## Required Fields ##
#####################

# Values determines what columns will be written to stdout.
#
# The supported columns are:
# id    - The ID of the halo, as initially supplied.
# snap  - The snapshot index of the halo, as initially supplied.
# r_sp  - The volume-weighted splashback radius of the halo.
# m_sp  - The total mass contained within the splashback shell of the halo.
# V_sp  - Volume of splashback shell.
# SA_sp - Surface area of splashback shell.
# a_sp  - Largest axis of ellipsoidal fit to splashback shell.
# b_sp  - Largest axis of ellipsoidal fit to splashback shell.
# c_sp  - Largest axis of ellipsoidal fit to splashback shell.
#
# WARNING: THIS IS NOT FULLY IMPLEMENTED
Values = snap, id, m_sp, r_sp, SA_sp, V_sp, a_sp, b_sp, c_sp

#####################
## Optional Fields ##
#####################

# MonteCarloSamplings The number of Monte Carlo samplings done when calculating
# properties of shells.
MonteCarloSamples = 50000

# ExclustionStrategy is the strategy for removing halos contained within a
# larger halo's splashback shell.
#
# The supported strategies are:
# none    - Don't remove halos.
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

# ShellParticleFile and ShellWidth allow Shellfish to output a file containing
# the IDs of particles which are close to the edge of the halo.
# ShellParticlesFile is a file that the IDs will be written out to, and
# ShellWidth is how far away from the shell a particle can be (as a multiplier
# of R200m) while still being output to the file.
#
# The format of the file is the following:
#
# |- 1 -| |- 2_0 -| ... |- 2_i -| |- ... 3_0 ... -| ... |- ... 3_i ... -|
#
# 1) HaloCount: int64
#        The number of halos in the file.
# 2_i) HaloInfo: struct { ID, Snap, StartByte, Particles int64 }
#        Summarizing information for the ith halo in the file. In order, the
#        fields represent the halo ID, the snapshot number of the halo, the
#        index of the first byte of the particle array corresponding to this
#        halo (i.e. the 'pos' parameter for seek() when using the flag
#        SEEK_SET), and the number of particles that are in that halo's array.
# 3_i) ParticleIDs: []int64
#        The IDs of the particles near the shell of the ith halo.
#
# The file will be written with the byte ordering specified by the Endianness
# variable in the global config file.
#
# If ShellParticleFile = "" or if ShellWidth = 0, no such file will be
# created.
# ShellParticleFile = shell-particles.dat
# ShellWidth = 0.05`
}

func (config *StatsConfig) ReadConfig(fname string) error {
	vars := parse.NewConfigVars("stats.config")

	vars.Strings(&config.values, "Values", []string{})
	vars.Int(&config.monteCarloSamples, "MonteCarloSamples", 50*1000)
	vars.String(&config.exclusionStrategy, "ExclusionStrategy", "none")
	vars.Int(&config.order, "Order", 3)
	vars.String(&config.shellParticleFile, "ShellParticleFile", "")
	vars.Float(&config.shellWidth, "ShellWidth", 0)

	if fname == "" {
		return nil
	}
	if err := parse.ReadConfig(fname, vars); err != nil {
		return err
	}

	config.shellFilter = config.shellWidth > 0 &&
		config.shellParticleFile != ""

	return config.validate()
}

func (config *StatsConfig) validate() error {
	for i, val := range config.values {
		switch val {
		case "snap", "id", "m_sp", "r_sp", "a_sp", "b_sp", "c_sp",
			"SA_sp/V_sp":
		default:
			return fmt.Errorf("Item %d of variable 'Values' is set to '%s', "+
				"which I don't recognize.", i, val)
		}
	}

	switch config.exclusionStrategy {
	case "none", "contain", "overlap":
	default:
		return fmt.Errorf("variable 'ExclusionStrategy' set to '%s', which "+
			"I don't recognize.", config.exclusionStrategy)
	}

	switch {
	case config.monteCarloSamples <= 0:
		return fmt.Errorf("The variable '%s' was set to %g",
			"MonteCarloSamples", config.monteCarloSamples)
	}

	return nil
}

func (config *StatsConfig) Run(
	flags []string, gConfig *GlobalConfig, e *env.Environment, stdin []string,
) ([]string, error) {

	if logging.Mode != logging.Nil {
		log.Println(`
#####################
## shellfish stats ##
#####################`,
		)
	}
	var t time.Time
	if logging.Mode == logging.Performance { t = time.Now() }

	intColIdxs := []int{0, 1}
	floatColIdxs := make([]int, 4+2*config.order*config.order)
	for i := range floatColIdxs {
		floatColIdxs[i] = i + len(intColIdxs)
	}
	intCols, floatCols, err := catalog.ParseCols(
		stdin, intColIdxs, floatColIdxs,
	)

	if err != nil {
		return nil, err
	}
	if len(intCols) == 0 {
		return nil, fmt.Errorf("No input IDs.")
	}
	ids, snaps := intCols[0], intCols[1]
	coords, coeffs := floatCols[:4], transpose(floatCols[4:])
	snapBins, coeffBins, idxBins := binCoeffsBySnap(snaps, ids, coeffs)

	masses := make([]float64, len(ids))

	rads := make([]float64, len(ids))
	rmins := make([]float64, len(ids))
	rmaxes := make([]float64, len(ids))
	vols := make([]float64, len(ids))
	sas := make([]float64, len(ids))
	as := make([]float64, len(ids))
	bs := make([]float64, len(ids))
	cs := make([]float64, len(ids))
	aVecs := make([][3]float64, len(ids))
	shellParticles := make([][]int64, len(ids))

	sortedSnaps := []int{}
	for snap := range snapBins {
		sortedSnaps = append(sortedSnaps, snap)
	}
	sort.Ints(sortedSnaps)

	buf, err := getVectorBuffer(
		e.ParticleCatalog(snaps[0], 0),
		gConfig.SnapshotType, gConfig.Endianness,
	)
	if err != nil {
		return nil, err
	}

	for _, snap := range sortedSnaps {
		if snap == -1 {
			continue
		}
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

		samples := int(config.monteCarloSamples)
		for j := range idxs {
			order := findOrder(coeffs[idxs[j]])
			shell := analyze.PennaFunc(coeffs[idxs[j]], order, order, 2)

			vol := shell.Volume(samples)
			r := math.Pow(vol/(math.Pi*4/3), 0.33333)

			vols[idxs[j]] = vol
			rads[idxs[j]] = r
			sas[idxs[j]] = shell.SurfaceArea(samples)
			as[idxs[j]], bs[idxs[j]], cs[idxs[j]], aVecs[idxs[j]] =
				shell.Axes(samples)


			rmins[idxs[j]], rmaxes[idxs[j]] = rangeSp(snapCoeffs[j], config)
		}

		hds, files, err := memo.ReadHeaders(snap, buf, e)
		if err != nil {
			return nil, err
		}
		hBounds, err := boundingSpheres(snapCoords, &hds[0], config, e)
		if err != nil {
			return nil, err
		}
		intrBins := binSphereIntersections(hds, hBounds)

		rLows := make([]float64, len(snapCoeffs))
		rHighs := make([]float64, len(snapCoeffs))
		for i := range snapCoeffs {
			// TODO: Figure out what's going on here and refactor.
			rLows[i], rHighs[i] = rangeSp(snapCoeffs[i], config)
		}

		for i := range hds {
			if len(intrBins[i]) == 0 {
				continue
			}

			xs, ms, pIDs, err := buf.Read(files[i])

			if err != nil {
				return nil, err
			}

			for j := range idxs {
				masses[idxs[j]] += massContained(
					&hds[i], xs, ms, snapCoeffs[j],
					hBounds[j], rLows[j], rHighs[j],
					gConfig.Threads,
				)

				if config.shellFilter {
					// This isn't the correct way to handle this for
					// performance, but massContained is already gross enough as
					// it is.
					shellParticles[idxs[j]] = appendShellParticles(
						&hds[i], xs, pIDs, snapCoeffs[j],
						hBounds[j], rLows[j], rHighs[j],
						config.shellWidth,
						gConfig.Threads,
						shellParticles[idxs[j]],
					)
				}

			}

			buf.Close()
		}
	}

	// TODO: Remove debugging code.
	for i := range shellParticles {
		log.Printf("Halo %d: %d shell particles.",
			i, len(shellParticles[i]))
	}


	axs := make([]float64, len(ids))
	ays := make([]float64, len(ids))
	azs := make([]float64, len(ids))
	for i := range axs {
		axs[i], ays[i], azs[i] = aVecs[i][0], aVecs[i][1], aVecs[i][2]
	}

	lines := catalog.FormatCols(
		[][]int{ids, snaps},
		[][]float64{masses, rads, vols, sas,
			as, bs, cs, axs, ays, azs},
		[]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
	)
	cString := catalog.CommentString(
		[]string{"ID", "Snapshot"},
		[]string{"M_sp [M_sun/h]", "R_sp [Mpc/h]",
			"Volume [Mpc^3/h^3]", "Surface Area [Mpc^2/h^2]",
			"Major Axis [Mpc/h]",
			"Intermediate Axis [Mpc/h]",
			"Minor Axis [Mpc/h]",
			"Ax", "Ay", "Az",
		},
		[]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		[]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	)

	if logging.Mode == logging.Performance {
		log.Printf("Time: %s", time.Since(t).String())
		log.Printf("Memory:\n%s", logging.MemString())
	}

	return append([]string{cString}, lines...), nil
}

func wrapDist(x1, x2, width float32) float32 {
	dist := x1 - x2
	if dist > width/2 {
		return dist - width
	} else if dist < width/-2 {
		return dist + width
	} else {
		return dist
	}
}

func inRange(x, r, low, width, tw float32) bool {
	return wrapDist(x, low, tw) > -r && wrapDist(x, low+width, tw) < r
}

// SheetIntersect returns true if the given halo and sheet intersect one another
// and false otherwise.
func sheetIntersect(s geom.Sphere, hd *io.Header) bool {
	tw := float32(hd.TotalWidth)
	return inRange(s.C[0], s.R, hd.Origin[0], hd.Width[0], tw) &&
		inRange(s.C[1], s.R, hd.Origin[1], hd.Width[1], tw) &&
		inRange(s.C[2], s.R, hd.Origin[2], hd.Width[2], tw)
}

func binCoeffsBySnap(
	snaps, ids []int, coeffs [][]float64,
) (snapBins map[int][]int, coeffBins map[int][][]float64, idxBins map[int][]int) {
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
	coords [][]float64, hd *io.Header, c *StatsConfig, e *env.Environment,
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

func rangeSp(coeffs []float64, c *StatsConfig) (rmin, rmax float64) {
	order := findOrder(coeffs)
	shell := analyze.PennaFunc(coeffs, order, order, 2)
	return shell.RadialRange(int(c.monteCarloSamples))
}

func massContained(
	hd *io.Header, xs [][3]float32, ms []float32, coeffs []float64,
	sphere geom.Sphere, rLow, rHigh float64, threads int64,
) float64 {

	cpu := runtime.NumCPU()
	if threads > 0 {
		cpu = int(threads)
	}
	workers := int64(runtime.GOMAXPROCS(cpu))
	outChan := make(chan float64, workers)
	for i := int64(0); i < workers-1; i++ {
		go massContainedChan(
			hd, xs, ms, coeffs, sphere, rLow, rHigh, i, workers, outChan,
		)
	}

	massContainedChan(
		hd, xs, ms, coeffs, sphere, rLow, rHigh,
		workers-1, workers, outChan,
	)

	sum := 0.0
	for i := int64(0); i < workers; i++ {
		sum += <-outChan
	}

	return sum
}

// TODO: Humans cannot remember this many parameters.

func appendShellParticles(
	hd *io.Header, xs [][3]float32, pIDs []int64, coeffs []float64,
	sphere geom.Sphere, rLow, rHigh float64, shellWidth float64,
	threads int64, out []int64,
) []int64 {
	cpu := runtime.NumCPU()
	if threads > 0 {
		cpu = int(threads)
	}
	workers := int64(runtime.GOMAXPROCS(cpu))
	outChan := make(chan []int64, workers)

	for i := int64(0); i < workers-1; i++ {
		go appendShellParticlesChan(
			hd, xs, pIDs, coeffs, sphere, rLow, rHigh,
			shellWidth, i, workers, outChan,
		)
	}

	appendShellParticlesChan(
		hd, xs, pIDs, coeffs, sphere, rLow, rHigh,
		shellWidth, workers-1, workers, outChan,
	)

	for i := int64(0); i < workers; i++ {
		out = append(out, <-outChan...)
	}

	return out
}

func massContainedChan(
	hd *io.Header, xs [][3]float32, ms []float32, coeffs []float64,
	sphere geom.Sphere, rLow, rHigh float64,
	offset, workers int64, out chan float64,
) {
	tw2 := float32(hd.TotalWidth) / 2

	order := findOrder(coeffs)
	shell := analyze.PennaFunc(coeffs, order, order, 2)
	low2, high2 := float32(rLow*rLow), float32(rHigh*rHigh)

	sum := 0.0
	for i := offset; i < hd.N; i += workers {
		x, y, z := xs[i][0], xs[i][1], xs[i][2]
		x, y, z = x-sphere.C[0], y-sphere.C[1], z-sphere.C[2]
		x = wrap(x, tw2)
		y = wrap(y, tw2)
		z = wrap(z, tw2)

		r2 := x*x + y*y + z*z

		if r2 < low2 || (r2 < high2 &&
			shell.Contains(float64(x), float64(y), float64(z))) {
			sum += float64(ms[i])
		}
	}
	out <- sum
}

func appendShellParticlesChan(
	hd *io.Header, xs [][3]float32, pIDs []int64, coeffs []float64,
	sphere geom.Sphere, rLow, rHigh float64, shellWidth float64,
	offset, workers int64, outChan chan []int64,
) {
	buf := []int64{}

	tw2 := float32(hd.TotalWidth) / 2

	order := findOrder(coeffs)
	shell := analyze.PennaFunc(coeffs, order, order, 2)
	low2, high2 := float32(rLow*rLow), float32(rHigh*rHigh)
	delta := float64(sphere.R) * shellWidth

	for i := offset; i < hd.N; i += workers {
		x, y, z := xs[i][0], xs[i][1], xs[i][2]
		x, y, z = x-sphere.C[0], y-sphere.C[1], z-sphere.C[2]
		x = wrap(x, tw2)
		y = wrap(y, tw2)
		z = wrap(z, tw2)

		r2 := x*x + y*y + z*z

		if r2 < low2 || r2 < high2 {
			r := math.Sqrt(float64(r2))
			phi := math.Atan2(float64(y), float64(x))
			theta := math.Acos(float64(z) / r)

			rs := shell(phi, theta)

			if rs + delta > r && rs - delta < r {
				buf = append(buf, pIDs[i])
			}
		}
	}
	outChan <- buf
}

func binSphereIntersections(
	hds []io.Header, spheres []geom.Sphere,
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
