package cmd

import (
	"fmt"
	"log"
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

type PotentialConfig struct {
	ncells int64
	rGridMult float64
	rMinMult, rMaxMult float64
}

var _ Mode = &PotentialConfig{}

func (config *PotentialConfig) ExampleConfig() string {
	return `[potential.config]
NCells = 64
GridRMult = 4
# RMinMult is the radius particles muse
RMinMult = 1
# RMaxMult is the maximum radius inside which particles are used to calculate
# the potential.
RMaxMult = 50
`
}


func (config *PotentialConfig) ReadConfig(fname string, flags []string) error {
	vars := parse.NewConfigVars("prof.config")

	vars.Int(&config.ncells, "RBins", 8)
	vars.Float(&config.rGridMult, "GridRMult", 4)
	vars.Float(&config.rMaxMult, "RMaxMult", 1.0)
	vars.Float(&config.rMinMult, "RMinMult", 1.0)

	if fname == "" {
		if len(flags) == 0 { return nil }

		err := parse.ReadFlags(flags, vars)
		if err != nil { return err }
	} else {
		if err := parse.ReadConfig(fname, vars); err != nil { return err }
		if err := parse.ReadFlags(flags, vars); err != nil { return err }
	}
	
	return config.validate()
}

func (config *PotentialConfig) validate() error {
	if config.ncells < 0 {
		return fmt.Errorf("The variable '%s' was set to %d.",
			"NCells", config.ncells)
	} else if config.rGridMult < 0 {
		return fmt.Errorf("The variable '%s' was set to %d.",
			"RGridMult", config.rGridMult)
	} else if config.rMaxMult < 0 {
		return fmt.Errorf("The variable '%s' was set to %d.",
			"RMaxMult", config.rMaxMult)
	} else if config.rMinMult < 0 {
		return fmt.Errorf("The variable '%s' was set to %d.",
			"RMinMult", config.rMinMult)
	}

	return nil
}

func (config *PotentialConfig) Run(
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

	icols, fcols, err := catalog.Parse(
		stdin, []int{0, 1}, []int{2, 3, 4, 5, 6},
	)
	if err != nil { return nil, err }

	ids, snaps := icols[0], icols[1]
	hx := [3][]float64{ fcols[0], fcols[1], fcols[2] }
	hr, hm := fcols[3], fcols[4]

	if len(ids) == 0 { return nil, fmt.Errorf("No input halos.") }

	// Initialize phase profiles
	rSets := make([][]float64, len(ids))
	phiSets := [3][][]float64{
		make([][]float64, len(ids)),
		make([][]float64, len(ids)),
		make([][]float64, len(ids)),
	}

	for i := range rSets {
		rSets[i] = make([]float64, 2*config.ncells)
		phiSets[0][i] = make([]float64, config.ncells*config.ncells)
		phiSets[1][i] = make([]float64, config.ncells*config.ncells)
		phiSets[2][i] = make([]float64, config.ncells*config.ncells)
	}

	snapBins, idxBins := binBySnap(snaps, ids)

	sortedSnaps := []int{}
	for snap := range snapBins {
		sortedSnaps = append(sortedSnaps, snap)
	}
	sort.Ints(sortedSnaps)
	
	buf, err := getVectorBuffer(
		e.ParticleCatalog(snaps[0], 0), gConfig,
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
			snapCoords[0][i] = hx[0][idx]
			snapCoords[1][i] = hx[1][idx]
			snapCoords[2][i] = hx[2][idx]
			snapCoords[3][i] = hr[idx]*config.rMaxMult
		}

		hds, files, err := memo.ReadHeaders(snap, buf, e)
		if err != nil {
			return nil, err
		}
		hxBounds, err := boundingSpheres(snapCoords, &hds[0], e)
		if err != nil {
			return nil, err
		}
		_, intrIdxs := binSphereIntersections(hds, hxBounds)

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
				phisXY := phiSets[0][idxs[j]]
				phisYZ := phiSets[1][idxs[j]]
				phisZX := phiSets[2][idxs[j]]

				insertPotentialPoints(
					phisXY, phisYZ, phisZX,
					hxBounds[j],
					xs, ms,
					config, &hds[i],
				)
			}

			buf.Close()
		}
	}
	
	for i := range rSets {
		processPotential(
			rSets[i],
			phiSets[0][i], phiSets[1][i], phiSets[2][i],
			hr[i], hm[i], config,
		)
	}

	rSets = transpose(rSets)
	phiSets[0] = transpose(phiSets[0])
	phiSets[1] = transpose(phiSets[1])
	phiSets[2] = transpose(phiSets[2])

	order := make([]int, len(rSets) + len(phiSets[0])*3 + 2)
	for i := range order { order[i] = i }
	lines := catalog.FormatCols(
		[][]int{ids, snaps},
		append(append(append(
			rSets, phiSets[0]...),
			phiSets[1]...),
			phiSets[2]...),
		order,
	)
	
	cString := catalog.CommentString(
		[]string{"ID", "Snapshot", "R [cMpc/h]",
			"Phi_xy/(Rvir/(G Mvir m))",
			"Phi_yz/(Rvir/(G Mvir m))",
			"Phi_zx/(Rvir/(G Mvir m))"},
		[]string{}, []int{0, 1, 2, 3, 4, 5},
		[]int{1, 1, int(config.ncells),
			int(config.ncells*config.ncells),
			int(config.ncells*config.ncells),
			int(config.ncells*config.ncells)},
	)

	if logging.Mode == logging.Performance {
		log.Printf("Time: %s", time.Since(t).String())
		log.Printf("Memory:\n%s", logging.MemString())
	}

	return append([]string{cString}, lines...), nil
}

func insertPotentialPoints(
	phisXY, phisYZ, phisZX []float64,
	hx geom.Sphere,
	xs [][3]float32,
	ms []float32,
	config *PotentialConfig, hd *io.Header,
) {
	panic("NYI")
}

func processPotential(
	rs, phisXY, phisYZ, phisZX []float64,
	hr, hm float64, config *PotentialConfig,
) {
	panic("NYI")
}
