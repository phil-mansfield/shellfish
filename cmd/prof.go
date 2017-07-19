package cmd

import (
	"fmt"
	"log"
	"math"
	"sort"
	"time"
	"math/rand"

	msort "github.com/phil-mansfield/shellfish/math/sort"
	"github.com/phil-mansfield/shellfish/los/geom"
	"github.com/phil-mansfield/shellfish/los/analyze"
	"github.com/phil-mansfield/shellfish/cmd/catalog"
	"github.com/phil-mansfield/shellfish/cmd/env"
	"github.com/phil-mansfield/shellfish/logging"
	"github.com/phil-mansfield/shellfish/parse"
	"github.com/phil-mansfield/shellfish/io"
	"github.com/phil-mansfield/shellfish/cmd/memo"
)

type ProfConfig struct {
	bins, order, samples int64
	rMaxMult, rMinMult float64
	medianPixelLevel int64
	percentile float64

	pType profileType

}

type profileType int
const (
	densityProfile profileType = iota
	medianDensityProfile
	medianErrorProfile
	containedDensityProfile
	angularFractionProfile
)

var _ Mode = &ProfConfig{}

func (config *ProfConfig) ExampleConfig() string {
	return `[prof.config]

#####################
## Required Fields ##
#####################

# ProfileType determines what type of profile will be output.
# Known profile types are:
# density -           The traditional spherical densiy profile that we all
#                     know and love.
# median-density -    A density profile created by binning particles into
#                     equal solid-angle pixels and taking the median density
#                     across the angular bins at each radius.
# median-error -      Calculates error on the median through bootstrap sampling.
# contained-densiy -  A density profile which only uses particles.
# angular-fraction -  The angular fraction at each radius which is contained
#                     within the shell.
ProfileType = median-density

# Order is the order of the Penna-Dines shell fit that Shellfish uses. This
# variable only needs to be set if ProfileType is set to contained-density
# or angular-fraction.
# Order = 3

# Samples is the number of Monte Carlo samples used when calculating angular
# fraction profiles. It does not need to be set when other profiles are
# calculated.
# Samples = 50000

# MedianPixelLevel sets the number of angular pixel used when ProfileType is
# set to median-density. Because of Shellfish's pixelization scheme (a modified
# version of the algorithm presented in Gringorten & Yepez, 1992), the total
# number of pixels used is 2*(2*level - 1)^2. Quickly, that means 1 -> 2,
# 2 -> 18, 3 -> 50, 10 -> 722 etc.
# MedianPixelLevel = 3

#####################
## Optional Fields ##
#####################

# Percentile sets the percentile used during median-profile mode so that
# non-median percentiles can be measured.
# Percentile = 50

# Bins is the number of logarithmic radial bins used in a profile.
# Bins = 150

# RMaxMult is the maximum radius of the profile as a function of R_200m.
# RMaxMult = 3

# RMinMult is the minimum radius of the profile as a function of R_200m.
# RMinMult = 0.03
`
}


func (config *ProfConfig) ReadConfig(fname string, flags []string) error {

	vars := parse.NewConfigVars("prof.config")

	vars.Int(&config.bins, "Bins", 150)
	vars.Int(&config.order, "Order", 3)
	vars.Int(&config.samples, "Samples", 50 * 1000)
	vars.Float(&config.rMaxMult, "RMaxMult", 3.0)
	vars.Float(&config.rMinMult, "RMinMult", 0.03)
	vars.Int(&config.medianPixelLevel, "MedianPixelLevel", 3)
	vars.Float(&config.percentile, "Percentile", 50)
	var pType string
	vars.String(&pType, "ProfileType", "")

	if fname == "" {
		if len(flags) == 0 {
			return nil
		}

		err := parse.ReadFlags(flags, vars)
		if err != nil {
			return err
		}

		return config.validate()		
	}
	if err := parse.ReadConfig(fname, vars); err != nil {
		return err
	}
	if err := parse.ReadFlags(flags, vars); err != nil {
		return err
	}

	// Needs to be done here: can't be in the validate method.
	switch pType {
	case "":
		return fmt.Errorf("The variable 'ProfileType' was not set.")
	case "density":
		config.pType = densityProfile
	case "median-density":
		config.pType = medianDensityProfile
	case "median-error":
		config.pType = medianErrorProfile
	case "contained-density":
		config.pType = containedDensityProfile
	case "angular-fraction":
		config.pType = angularFractionProfile
	default:
		return fmt.Errorf("The varaiable 'ProfileType' was set to '%s'.", pType)
	}

	return config.validate()
}

func (config *ProfConfig) validate() error {
	if config.bins < 0 {
		return fmt.Errorf("The variable '%s' was set to %d.",
			"Bins", config.bins)
	} else if config.rMinMult <= 0 {
		return fmt.Errorf("The variable '%s' was set to %g.",
			"RMinMult", config.rMinMult)
	} else if config.rMaxMult <= 0 {
		return fmt.Errorf("The variable '%s' was set to %g.",
			"RMinMult", config.rMinMult)
	} else if config.medianPixelLevel < 0 {
		return fmt.Errorf("The variable '%s' was set to %g.",
			"MedianPixelLevel", config.medianPixelLevel)
	}

	return nil
}

func (config *ProfConfig) Run(
	gConfig *GlobalConfig, e *env.Environment, stdin []byte,
) ([]string, error) {
	if logging.Mode != logging.Nil {
		log.Println(`
####################
## shellfish prof ##
####################`,
		)
	}
	
	var t time.Time
	if logging.Mode == logging.Performance {
		t = time.Now()
	}

	var (
		intCols [][]int
		coords [][]float64
		shells []analyze.Shell
		err error
	)

	switch config.pType {
	case densityProfile, medianDensityProfile, medianErrorProfile:
		intColIdxs := []int{0, 1}
		floatColIdxs := []int{2, 3, 4, 5}
		
		intCols, coords, err = catalog.Parse(
			stdin, intColIdxs, floatColIdxs,
		)
		
		if err != nil {
			return nil, err
		}

		shells = make([]analyze.Shell, len(coords[0]))
	case containedDensityProfile, angularFractionProfile:
		intColIdxs := []int{0, 1}
		floatColIdxs := make([]int, 4 + config.order*config.order*2)
		for i := range floatColIdxs {
			floatColIdxs[i] += i + 2
		}

		var floatCols [][]float64
		intCols, floatCols, err = catalog.Parse(
			stdin, intColIdxs, floatColIdxs,
		)

		if err != nil {
			return nil, err
		}

		coords = floatCols[:4]
		coeffs := floatCols[4:]
		shells = make([]analyze.Shell, len(coords[0]))
		for i := range shells {
			coeffVec := make([]float64, len(coeffs))
			for j := range coeffVec {
				coeffVec[j] = coeffs[j][i]
			}
			order := int(config.order)
			shells[i] = analyze.PennaFunc(coeffVec, order, order, 2)
		}
	}

	if len(intCols) == 0 {
		return nil, fmt.Errorf("No input IDs.")
	}

	ids, snaps := intCols[0], intCols[1]
	snapBins, idxBins := binBySnap(snaps, ids)

	if config.pType == angularFractionProfile {
		return angularFractionMain(ids, snaps, shells, coords[3], config)
	}

	// Profiles for everyone
	rSets := make([][]float64, len(ids))
	rhoSets := make([][]float64, len(ids))
	for i := range rSets {
		rSets[i] = make([]float64, config.bins)
		rhoSets[i] = make([]float64, config.bins)
	}

	// Workspace buffers just for the median-density mode.
	var (
		medRhoSets [][][]float64
		medScratchBuffer []float64
	)
	if config.pType == medianDensityProfile ||
		config.pType == medianErrorProfile {

		medRhoSets = make([][][]float64, len(ids))
		n := geom.SpherePixelNum(int(config.medianPixelLevel))
		medScratchBuffer = make([]float64, n)
		for i := range medRhoSets {
			medRhoSets[i] = make([][]float64, config.bins)
			for j := range medRhoSets[i] {
				medRhoSets[i][j] = make([]float64, n)
			}
		}

	}


	sortedSnaps := []int{}
	for snap := range snapBins {
		sortedSnaps = append(sortedSnaps, snap)
	}
	sort.Ints(sortedSnaps)

	buf, err := getVectorBuffer(
		e.ParticleCatalog(snaps[0], 0),
		gConfig.SnapshotType, gConfig.Endianness,
		gConfig.GadgetNpartNum,
	)
	if err != nil {
		return nil, err
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

		hds, files, err := memo.ReadHeaders(snap, buf, e)
		if err != nil {
			return nil, err
		}
		hBounds, err := boundingSpheres(snapCoords, &hds[0], e)
		if err != nil {
			return nil, err
		}
		_, intrIdxs := binSphereIntersections(hds, hBounds)

		for i := range hds {
			if len(intrIdxs[i]) == 0 {
				continue
			}

			xs, _, ms, _, err := buf.Read(files[i])
			if err != nil {
				return nil, err
			}

			// Waarrrgggble
			for _, j := range intrIdxs[i] {
				rhos := rhoSets[idxs[j]]
				s := hBounds[j]

				if config.pType == medianDensityProfile ||
					config.pType == medianErrorProfile {
					medRhos := medRhoSets[idxs[j]]
					insertMedianPoints(medRhos, s, xs, ms, config, &hds[i])
				} else {
					insertPoints(
						rhos, s, xs, ms, shells[idxs[j]], config, &hds[i],
					)
				}
			}

			buf.Close()
		}
	}
	
	for i := range rSets {
		rMax := coords[3][i]*config.rMaxMult
		rMin := coords[3][i]*config.rMinMult
		if config.pType == medianDensityProfile {
			processMedianProfile(rSets[i], rhoSets[i],
				medRhoSets[i], medScratchBuffer, rMin, rMax,
				config.percentile,
			)
		} else if config.pType == medianErrorProfile {
			processMedianErrorProfile(rSets[i], rhoSets[i],
				medRhoSets[i], medScratchBuffer, rMin, rMax,
				config.percentile, config.samples,
			)
		} else {
			processProfile(rSets[i], rhoSets[i], rMin, rMax)
		}
	}

	rSets = transpose(rSets)
	rhoSets = transpose(rhoSets)

	order := make([]int, len(rSets) + len(rhoSets) + 2)
	for i := range order { order[i] = i }
	lines := catalog.FormatCols(
			[][]int{ids, snaps}, append(rSets, rhoSets...), order,
	)
	
	cString := catalog.CommentString(
		[]string{"ID", "Snapshot", "R [cMpc/h]", "Rho [h^2 Msun/cMpc^3]"},
		[]string{}, []int{0, 1, 2, 3},
		[]int{1, 1, int(config.bins), int(config.bins)},
	)

	if logging.Mode == logging.Performance {
		log.Printf("Time: %s", time.Since(t).String())
		log.Printf("Memory:\n%s", logging.MemString())
	}

	return append([]string{cString}, lines...), nil
}

func insertPoints(
	rhos []float64, s geom.Sphere, xs [][3]float32,
	ms []float32, shell analyze.Shell, config *ProfConfig, hd *io.Header,
) {
	lrMax := math.Log(float64(s.R) * config.rMaxMult)
	lrMin := math.Log(float64(s.R) * config.rMinMult)
	dlr := (lrMax - lrMin) / float64(config.bins)
	rMax2 := s.R * float32(config.rMaxMult)
	rMin2 := s.R * float32(config.rMinMult)
	rMax2 *= rMax2
	rMin2 *= rMin2

	x0, y0, z0 := s.C[0], s.C[1], s.C[2]
	tw2 := float32(hd.TotalWidth) / 2
	
	for i, vec := range xs {
		x, y, z := vec[0], vec[1], vec[2]
		dx, dy, dz := x - x0, y - y0, z - z0
		dx = wrap(dx, tw2)
		dy = wrap(dy, tw2)
		dz = wrap(dz, tw2)

		r2 := dx*dx + dy*dy + dz*dz
		if r2 <= rMin2 || r2 >= rMax2 {
			continue
		}

		if config.pType == containedDensityProfile &&
			!shell.Contains(float64(dx), float64(dy), float64(dz)) {
			continue
		}

		lr := math.Log(float64(r2)) / 2
		ir := int(((lr) - lrMin) / dlr)
		if ir == len(rhos) { ir-- }
		rhos[ir] += float64(ms[i])
	}
}

func insertMedianPoints(
	medRhos [][]float64, s geom.Sphere,  xs [][3]float32,
	ms []float32, config *ProfConfig, hd *io.Header,
) {
	lrMax := math.Log(float64(s.R) * config.rMaxMult)
	lrMin := math.Log(float64(s.R) * config.rMinMult)
	dlr := (lrMax - lrMin) / float64(config.bins)
	rMax2 := s.R * float32(config.rMaxMult)
	rMin2 := s.R * float32(config.rMinMult)
	rMax2 *= rMax2
	rMin2 *= rMin2

	x0, y0, z0 := s.C[0], s.C[1], s.C[2]
	tw2 := float32(hd.TotalWidth) / 2

	pixelNum := geom.SpherePixelNum(int(config.medianPixelLevel))

	for i, vec := range xs {
		x, y, z := vec[0], vec[1], vec[2]
		dx, dy, dz := x - x0, y - y0, z - z0
		dx = wrap(dx, tw2)
		dy = wrap(dy, tw2)
		dz = wrap(dz, tw2)

		r2 := dx*dx + dy*dy + dz*dz
		if r2 <= rMin2 || r2 >= rMax2 {
			continue
		}

		r := math.Sqrt(float64(r2))
		phi := math.Mod(
			math.Atan2(float64(dy), float64(dx)) + math.Pi*2, math.Pi*2,
		)
		th := math.Acos(float64(dz) / r)
		p := geom.SpherePixel(phi, th, int(config.medianPixelLevel))

		lr := math.Log(r)
		ir := int(((lr) - lrMin) / dlr)

		if ir == len(medRhos) { ir-- }
		medRhos[ir][p] += float64(ms[i])*float64(pixelNum)
	}
}

func processProfile(rs, rhos []float64, rMin, rMax float64) {
	n := len(rs)

	dlr := (math.Log(rMax) - math.Log(rMin)) / float64(n)
	lrMin := math.Log(rMin)

	for j := range rs {
		rs[j] = math.Exp(lrMin + dlr*(float64(j) + 0.5))

		rLo := math.Exp(dlr*float64(j) + lrMin)
		rHi := math.Exp(dlr*float64(j+1) + lrMin)
		dV := (rHi*rHi*rHi - rLo*rLo*rLo) * 4 * math.Pi / 3

		rhos[j] = rhos[j] / dV
	}
}

func processMedianProfile(rs, rhos []float64, medRhos [][]float64,
	medScratchBuffer []float64, rMin, rMax float64,
	percentile float64,
) {
	n := len(rs)

	dlr := (math.Log(rMax) - math.Log(rMin)) / float64(n)
	lrMin := math.Log(rMin)

	for j := range rs {
		rs[j] = math.Exp(lrMin + dlr*(float64(j) + 0.5))

		rLo := math.Exp(dlr*float64(j) + lrMin)
		rHi := math.Exp(dlr*float64(j+1) + lrMin)
		dV := (rHi*rHi*rHi - rLo*rLo*rLo) * 4 * math.Pi / 3

		rhos[j] = msort.Percentile(
			medRhos[j], percentile/100, medScratchBuffer,
		) / dV
	}
}

func processMedianErrorProfile(rs, rhos []float64, medRhos [][]float64,
	medScratchBuffer []float64, rMin, rMax float64,
	percentile float64, samples int64,
) {
	n := len(rs)

	dlr := (math.Log(rMax) - math.Log(rMin)) / float64(n)
	lrMin := math.Log(rMin)

	for j := range rs {
		rs[j] = math.Exp(lrMin + dlr*(float64(j) + 0.5))

		rLo := math.Exp(dlr*float64(j) + lrMin)
		rHi := math.Exp(dlr*float64(j+1) + lrMin)
		dV := (rHi*rHi*rHi - rLo*rLo*rLo) * 4 * math.Pi / 3

		rhos[j] = bootstrapErrorPercentile(
			medRhos[j], percentile, medScratchBuffer, samples,
		) / dV
	}
}

func bootstrapErrorPercentile(
	x []float64, percentile float64, scratchBuffer []float64, samples int64,
) float64 {
	sampleBuffer := make([]float64, len(x))

	sum := 0.0
	sqrSum := 0.0

	for i := int64(0); i < samples; i++ {
		for j := range x {
			sampleBuffer[j] = x[rand.Intn(len(x))]
		}
		p := msort.Percentile(sampleBuffer, percentile/100, scratchBuffer)
		sum += p
		sqrSum += p*p
	}

	sum /= float64(samples)
	sqrSum /= float64(samples)

	return math.Sqrt(sqrSum - sum*sum)
}

func angularFractionMain(
	ids, snaps []int, shells []analyze.Shell, rs []float64, config *ProfConfig,
) ([]string, error) {
	rCols := make([][]float64, config.bins)
	fCols := make([][]float64, config.bins)
	for i := range rCols {
		rCols[i] = make([]float64, len(ids))
		fCols[i] = make([]float64, len(ids))
	}

	for i := range shells {
		rs, fs := shells[i].AngularFractionProfile(
			int(config.samples), int(config.bins),
			rs[i] * config.rMinMult, rs[i] * config.rMaxMult,
		)

		for j := range rs {
			rCols[j][i], fCols[j][i] = rs[j], fs[j]
		}
	}

	order := make([]int, len(rCols) + len(fCols) + 2)
	for i := range order { order[i] = i }
	lines := catalog.FormatCols(
		[][]int{ids, snaps}, append(rCols, fCols...), order,
	)

	cString := catalog.CommentString(
		[]string{"ID", "Snapshot", "R [cMpc/h]", "Volume Fraction Contained"},
		[]string{}, []int{0, 1, 2, 3},
		[]int{1, 1, int(config.bins), int(config.bins)},
	)

	return append([]string{cString}, lines...), nil
}
