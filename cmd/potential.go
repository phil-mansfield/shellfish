package cmd

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"time"
	"runtime"

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
	frac float64
}

var _ Mode = &PotentialConfig{}

func (config *PotentialConfig) ExampleConfig() string {
	return `[potential.config]
NCells = 64
GridRMult = 8
# RMinMult is the minimum radius inside which particles are used to calculate
# the potential.
RMinMult = 1
# RMaxMult is the maximum radius inside which particles are used to calculate
# the potential.
RMaxMult = 50
# Percentage of particles that will be used to compute the potential.
ParticleFraction = 0.01
`
}


func (config *PotentialConfig) ReadConfig(fname string, flags []string) error {
	vars := parse.NewConfigVars("prof.config")

	vars.Int(&config.ncells, "RBins", 64)
	vars.Float(&config.rGridMult, "GridRMult", 8)
	vars.Float(&config.rMaxMult, "RMaxMult", 1.0)
	vars.Float(&config.rMinMult, "RMinMult", 1.0)
	vars.Float(&config.frac, "ParticleFraction", 1.0)

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
		rSets[i] = make([]float64, config.ncells)
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

	// Count number of workers

	workers := runtime.NumCPU()
	if gConfig.Threads > 0 { workers = int(gConfig.Threads) }
	runtime.GOMAXPROCS(workers)

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
				phisXZ := phiSets[2][idxs[j]]

				lg := NewLockGroup(workers)

				for k := 0; k < workers; k++ {
					insertPotentialPoints(
						phisXY, phisYZ, phisXZ,
						hxBounds[j],
						xs, ms,
						config, &hds[i],
						lg.Lock(k),
					)
				}

				lg.Synchronize()
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
			"Phi_xy/(G Mvir / Rvir)",
			"Phi_yz/(G Mvir / Rvir)",
			"Phi_xz/(G Mvir / Rvir)"},
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
	phisXY, phisYZ, phisXZ []float64,
	hx geom.Sphere,
	xs [][3]float32,
	ms []float32,
	config *PotentialConfig, hd *io.Header,
	lock *Lock,
) {
	rmax2 := (config.rMaxMult*config.rMaxMult) * float64(hx.R*hx.R)
	rmin2 := (config.rMinMult*config.rMinMult) * float64(hx.R*hx.R)

	gridR := float64(hx.R) * config.rGridMult
	delta := gridR * 2 / float64(config.ncells)
	nc := int(config.ncells)

	for i := range xs {
		// Think carefully about this bit. (although it actually makes
		// convergence tests easier to evaluate.)
		if rand.Float64() > config.frac { continue }

		dx0 := float64(xs[i][0] - hx.C[0])
		dy0 := float64(xs[i][1] - hx.C[1])
		dz0 := float64(xs[i][2] - hx.C[2])

		dr02 := dx0*dx0 + dy0*dy0 + dz0*dz0
		pm := float64(ms[i])

		if dr02 > rmax2 || dr02 < rmin2 { continue }

		for i := lock.Idx; i < len(phisXY); i += lock.Workers {
			var (
				ix, iy, iz int
				dx, dy, dz, dr float64
			)

			// phiXY
			ix, iy = i % nc, i / nc
			dx = dx0 - gridR + delta*float64(ix)
			dy = dy0 - gridR + delta*float64(iy)
			dr = math.Sqrt(dx*dx + dy*dy + dz0*dz0)
			phisXY[i] -= pm/dr

			// phiXZ
			ix, iz = i % nc, i / nc
			dx = dx0 - gridR + delta*float64(ix)
			dz = dz0 - gridR + delta*float64(iz)
			dr = math.Sqrt(dx*dx + dy0*dy0 + dz*dz)
			phisXZ[i] -= pm/dr

			// phiYZ
			iy, iz = i % nc, i / nc
			dy = dy0 - gridR + delta*float64(iy)
			dz = dz0 - gridR + delta*float64(iz)
			dr = math.Sqrt(dx0*dx0 + dy*dy + dz*dz)
			phisYZ[i] -= pm/dr
		}
	}

	lock.Unlock()
}

func processPotential(
	rs, phisXY, phisYZ, phisXZ []float64,
	hr, hm float64, config *PotentialConfig,
) {
	for i := range phisXY {
		phisXY[i] /= config.frac * hm/hr
		phisXZ[i] /= config.frac * hm/hr
		phisYZ[i] /= config.frac * hm/hr
	}

	delta := hr * config.rGridMult * 2 / float64(config.ncells)
	for i := range rs {
		rs[i] = delta*float64(i) - hr*config.rGridMult
	}
}
