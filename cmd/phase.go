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

type PhaseConfig struct {
	rbins, vbins int64
	rMaxMult, vMaxMult float64
	pType phaseProfileType
	subHub bool
}

type phaseProfileType int

const (
	radialPhaseProfile phaseProfileType = iota
	totalPhaseProfile
)

var _ Mode = &PhaseConfig{}

func (config *PhaseConfig) ExampleConfig() string {
	return `[phase.config]

# radial | total
ProfileType = radial

###################
# Optional fields #
###################

# RBins = 100
# VBins = 100

# Multiplies R200m:
# RMaxMult = 3.0
# Mutliplies V200m:
# VMaxMult = 3.0

# SubtractHubble = false
`
}


func (config *PhaseConfig) ReadConfig(fname string, flags []string) error {
	vars := parse.NewConfigVars("prof.config")

	vars.Int(&config.rbins, "RBins", 100)
	vars.Int(&config.vbins, "VBins", 100)
	vars.Float(&config.rMaxMult, "RMaxMult", 3.0)
	vars.Float(&config.vMaxMult, "VMaxMult", 3.0)
	vars.Bool(&config.subHub, "SubtractHubble", false)
	
	var pType string
	vars.String(&pType, "ProfileType", "")

	if fname == "" {
		if len(flags) == 0 { return nil }

		err := parse.ReadFlags(flags, vars)
		if err != nil { return err }
	} else {
		if err := parse.ReadConfig(fname, vars); err != nil { return err }
		if err := parse.ReadFlags(flags, vars); err != nil { return err }
	}
	
	// Needs to be done here: can't be in the validate method.
	switch pType {
	case "":
		return fmt.Errorf("The variable 'ProfileType' was not set.")
	case "radial":
		config.pType = radialPhaseProfile
	case "total":
		config.pType = totalPhaseProfile
	default:
		return fmt.Errorf("The varaiable 'ProfileType' was set to '%s'.", pType)
	}
	
	return config.validate()
}

func (config *PhaseConfig) validate() error {
	if config.rbins < 0 {
		return fmt.Errorf("The variable '%s' was set to %d.",
			"RBins", config.rbins)
	} else if config.vbins < 0 {
		return fmt.Errorf("The variable '%s' was set to %d.",
			"VBins", config.rbins)
	} else if config.rMaxMult < 0 {
		return fmt.Errorf("The variable '%s' was set to %d.",
			"RMaxMult", config.rbins)
	} else if config.vMaxMult < 0 {
		return fmt.Errorf("The variable '%s' was set to %d.",
			"VMaxMult", config.rbins)
	}

	return nil
}

func (config *PhaseConfig) Run(
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
		stdin, []int{0, 1}, []int{2, 3, 4, 5, 6, 7, 8, 9},
	)
	if err != nil { return nil, err }

	ids, snaps := icols[0], icols[1]
	hx := [3][]float64{ fcols[0], fcols[1], fcols[2] }
	hr, hm := fcols[3], fcols[4]
	hvx := [3][]float64{ fcols[5], fcols[6], fcols[7] }

	hvr := make([]float64, len(hr))
	for i := range hvr {
		hvr[i] = 508.0 * math.Sqrt((hm[i] / 6e13) / (hr[i] / 1.0))
	}

	if len(ids) == 0 { return nil, fmt.Errorf("No input halos.") }

	// Initialize phase profiles
	rSets := make([][]float64, len(ids))
	vSets := make([][]float64, len(ids))
	rhoSets := make([][]float64, len(ids))
	for i := range rSets {
		rSets[i] = make([]float64, config.rbins)
		vSets[i] = make([]float64, config.vbins)
		rhoSets[i] = make([]float64, config.rbins*config.vbins)
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
		snapVel := [][]float64{
			make([]float64, len(idxs)), make([]float64, len(idxs)),
			make([]float64, len(idxs)), make([]float64, len(idxs)),
		}
		for i, idx := range idxs {
			snapCoords[0][i] = hx[0][idx]
			snapCoords[1][i] = hx[1][idx]
			snapCoords[2][i] = hx[2][idx]
			snapCoords[3][i] = hr[idx]*config.rMaxMult

			snapVel[0][i] = hvx[0][idx]
			snapVel[1][i] = hvx[1][idx]
			snapVel[2][i] = hvx[2][idx]
			snapVel[3][i] = hvr[idx]*config.vMaxMult
		}

		hds, files, err := memo.ReadHeaders(snap, buf, e)
		if err != nil {
			return nil, err
		}
		hxBounds, err := boundingSpheres(snapCoords, &hds[0], e)
		hvBounds, err := boundingSpheres(snapVel, &hds[0], e)
		if err != nil {
			return nil, err
		}
		_, intrIdxs := binSphereIntersections(hds, hxBounds)

		for i := range hds {
			if len(intrIdxs[i]) == 0 {
				continue
			}

			xs, vs, ms, _, err := buf.Read(files[i])
			if err != nil {
				return nil, err
			}

			// Waarrrgggble
			for _, j := range intrIdxs[i] {
				rhos := rhoSets[idxs[j]]

				insertPhasePoints(
					rhos,
					hxBounds[j], hvBounds[j],
					xs, vs, ms,
					config, &hds[i],
				)
			}

			buf.Close()
		}
	}
	
	for i := range rSets {
		rMax := hr[i]*config.rMaxMult
		vMax := hvr[i]*config.vMaxMult
		processPhaseProfile(
			rSets[i], vSets[i], rhoSets[i], rMax, vMax, config.pType,
		)
	}

	rSets = transpose(rSets)
	vSets = transpose(vSets)
	rhoSets = transpose(rhoSets)

	order := make([]int, len(rSets) + len(vSets) + len(rhoSets) + 2)
	for i := range order { order[i] = i }
	lines := catalog.FormatCols(
		[][]int{ids, snaps},
		append(append(rSets, vSets...), rhoSets...),
		order,
	)
	
	cString := catalog.CommentString(
		[]string{"ID", "Snapshot", "R [cMpc/h]", "V [pkm/s]",
			"Rho [h^2 Msun/cMpc^3/(pkm/s), V major]"},
		[]string{}, []int{0, 1, 2, 3, 4},
		[]int{1, 1, int(config.vbins), int(config.rbins),
			int(config.rbins)*int(config.vbins)},
	)

	if logging.Mode == logging.Performance {
		log.Printf("Time: %s", time.Since(t).String())
		log.Printf("Memory:\n%s", logging.MemString())
	}

	return append([]string{cString}, lines...), nil
}

func insertPhasePoints(
	rhos []float64,
	hx, hv geom.Sphere,
	xs, vs [][3]float32,
	ms []float32,
	config *PhaseConfig, hd *io.Header,
) {
	vMin, vMax, rMax := 0.0, float64(hv.R), float64(hx.R)
	if config.pType == radialPhaseProfile { vMin = -vMax }
	dv := (vMax - vMin) / float64(config.vbins)
	dr := rMax / float64(config.rbins)

	rMax2 := rMax*rMax

	x0, y0, z0 := hx.C[0], hx.C[1], hx.C[2]
	vx0, vy0, vz0 := hv.C[0], hv.C[1], hv.C[2]
	tw2 := float32(hd.TotalWidth) / 2

	for i, vec := range xs {
		x, y, z := vec[0], vec[1], vec[2]
		dx, dy, dz := x - x0, y - y0, z - z0
		dx = wrap(dx, tw2)
		dy = wrap(dy, tw2)
		dz = wrap(dz, tw2)

		r2 := float64(dx*dx + dy*dy + dz*dz)
		if r2 >= rMax2 { continue }

		r := math.Sqrt(r2)

		var v float64
		vx, vy, vz := vs[i][0] - vx0, vs[i][1] - vy0, vs[i][2] - vz0
		if config.pType == radialPhaseProfile {
			v = float64(vx*dx + vy*dy + vz*dz) / r
			if config.subHub {
				h0 := 70.0 // TODO: Fix this.
				v -= r * h0
			}
		} else {
			if config.subHub {
				h0 := float32(70.0) // TODO: Fix this.
				vxHub, vyHub, vzHub := h0*dx, h0*dy, h0*dz
				vx, vy, vz = vx - vxHub, vy - vyHub, vz - vzHub
			}
			v = math.Sqrt(float64(vx*vx + vy*vy + vz*vz))
		}

		ir := int(r / dr)
		if ir == int(config.rbins) { ir-- }
		iv := int((v - vMin) / dv)
		if iv >= int(config.vbins) || iv < 0 { continue }

		rhos[ir*int(config.vbins) + iv] += float64(ms[i])
	}
}
func processPhaseProfile(
	rs, vs, rhos []float64, rMax, vMax float64, pType phaseProfileType,
) {	
	vMin := 0.0
	if pType == radialPhaseProfile { vMin = -vMax }

	dv := (vMax - vMin) / float64(len(vs))
	dr := rMax / float64(len(rs))
	
	for i := range vs {
		vs[i] = vMin + dv*(float64(i) + 0.5)
	}
	
	for j := range rs {
		rs[j] = dr*(float64(j) + 0.5)

		rLo := dr*float64(j)
		rHi := dr*float64(j+1)
		dV := (rHi*rHi*rHi - rLo*rLo*rLo) * 4 * math.Pi / 3
		
		for i := range vs {
			rhos[j*len(vs) + i] = rhos[j*len(vs) + i] / (dV * dv)
		}
	}
}
