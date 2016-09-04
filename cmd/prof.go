package cmd

import (
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"github.com/phil-mansfield/shellfish/los/geom"
	"github.com/phil-mansfield/shellfish/cmd/catalog"
	"github.com/phil-mansfield/shellfish/cmd/env"
	"github.com/phil-mansfield/shellfish/logging"
	"github.com/phil-mansfield/shellfish/parse"
	"github.com/phil-mansfield/shellfish/io"
	"github.com/phil-mansfield/shellfish/cmd/memo"
)

type ProfConfig struct {
	bins, order int64
	pType profileType
	
	rMaxMult, rMinMult float64
}

type profileType int
const (
	densityProfile profileType = iota
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
# density -          the traditional spherical densiy profile that we all
                     know and love.
# contained-densiy - a density profile which only uses particles.
# angular-fraction - the angular fraction at each radius which is contained
                     within the shell.
ProfileType = density

# Order is the order of the Penna-Dines shell fit that Shellfish uses. This
# variable only needs to be set if ProfileType is set to contained-density
# or angular-fraction.
Order = 3

#####################
## Optional Fields ##
#####################

# Bins is the number of logarithmic radial bins used in a profile.
# Bins = 150

# RMaxMult is the maximum radius of the profile as a function of R_200m.
# RMaxMult = 3

# RMinMult is the minimum radius of the profile as a function of R_200m.
# RMinMult = 0.03
`
}


func (config *ProfConfig) ReadConfig(fname string) error {
	if fname == "" {
		return nil
	}

	vars := parse.NewConfigVars("shell.config")

	vars.Int(&config.bins, "Bins", 150)
	vars.Int(&config.order, "Order", 3)
	vars.Float(&config.rMaxMult, "RMaxMult", 3.0)
	vars.Float(&config.rMinMult, "RMinMult", 0.03)
	var pType string
	vars.String(&pType, "ProfileType", "")

	if err := parse.ReadConfig(fname, vars); err != nil {
		return err
	}

	switch pType {
	case "":
		return fmt.Errorf("The variable 'ProfileType' was not set".)
	case "density":
		config.pType = densityProfile
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
	}
	return nil
}

func (config *ProfConfig) Run(
	flags []string, gConfig *GlobalConfig, e *env.Environment, stdin []string,
) ([]string, error) {
	if logging.Mode != logging.Nil {
		log.Println(`
####################
## shellfish tree ##
####################`,
		)
	}
	
	var t time.Time
	if logging.Mode == logging.Performance {
		t = time.Now()
	}

	var (
		intCols []int
		coords, coeffs []float64
		err error
	)

	switch config.pType {
	case densityProfile:
		intColIdxs := []int{0, 1}
		floatColIdxs := []int{2, 3, 4, 5}
		
		intCols, coords, err = catalog.ParseCols(
			stdin, intColIdxs, floatColIdxs,
		)
		
		if err != nil {
			return nil, err
		}
	case containedDensityProfile, angularFractionProfile:
		intColIdxs := []int{0, 1}
		floatColIdxs := make([]int, 4 + pType.order*pType.order*2)
		for i := range floatColIdxs {
			floatColIdxs[i] + i + 2
		}

		var floatCols []float64
		intCols, floatCols, err = catalog.ParseCols(
			stdin, intColIdxs, floatColIdxs,
		)

		if err != nil {
			return nil, err
		}

		coords = floatCols[:4]
		coeffs = floatCols[4:]
	}
	
	if len(intCols) == 0 {
		return nil, fmt.Errorf("No input IDs.")
	}

	ids, snaps := intCols[0], intCols[1]
	snapBins, idxBins := binBySnap(snaps, ids)

	rSets := make([][]float64, len(ids))
	rhoSets := make([][]float64, len(ids))
	for i := range rSets {
		rSets[i] = make([]float64, config.bins)
		rhoSets[i] = make([]float64, config.bins)
	}

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

			xs, ms, _, err := buf.Read(files[i])
			if err != nil {
				return nil, err
			}

			// Waarrrgggble
			for _, j := range intrIdxs[i] {
				rhos := rhoSets[idxs[j]]
				s := hBounds[j]

				insertPoints(rhos, s, xs, ms, config, &hds[i])
			}

			buf.Close()
		}
	}

	for i := range rSets {
		rMax := coords[3][i]*config.rMaxMult
		rMin := coords[3][i]*config.rMinMult
		processProfile(rSets[i], rhoSets[i], rMin, rMax)
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
	
	for i, vec := range xs {
		x, y, z := vec[0], vec[1], vec[2]
		dx, dy, dz := x - x0, y - y0, z - z0
		dx = wrap(dx, tw2)
		dy = wrap(dy, tw2)
		dz = wrap(dz, tw2)

		r2 := dx*dx + dy*dy + dz*dz
		if r2 <= rMin2 || r2 >= rMax2 { continue }
		lr := math.Log(float64(r2)) / 2
		ir := int(((lr) - lrMin) / dlr)
		if ir == len(rhos) { ir-- }
		if ir < 0 || i < 0 || ir >= len(rhos) || i >= len(ms) {
			log.Println(
				"ir", ir,
				"i", i,
				"|rhos|", len(rhos),
				"|ms|", len(ms),
			)
		}
		rhos[ir] += float64(ms[i])
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
