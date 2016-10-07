package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/phil-mansfield/shellfish/cmd/catalog"
	"github.com/phil-mansfield/shellfish/cmd/env"
	"github.com/phil-mansfield/shellfish/cmd/halo"
	"github.com/phil-mansfield/shellfish/cmd/memo"
	"github.com/phil-mansfield/shellfish/io"
	"github.com/phil-mansfield/shellfish/logging"
	"github.com/phil-mansfield/shellfish/parse"
)

type CoordConfig struct {
	valueNames []string
}

var _ Mode = &CoordConfig{}

func (config *CoordConfig) ExampleConfig() string {
	return`[coord.config]

####################
## OptionalFields ##
####################

ValueNames = X, Y, Z, R200m
`
}

func (config *CoordConfig) ReadConfig(fname string) error {
		if fname == "" {
		return nil
	}

	vars := parse.NewConfigVars("tree.config")
	vars.Strings(&config.valueNames, "ValueNames", []string{})

	return parse.ReadConfig(fname, vars)
}

func (config *CoordConfig) validate() error { return nil }

func (config *CoordConfig) Run(
	flags []string, gConfig *GlobalConfig, e *env.Environment, stdin []string,
) ([]string, error) {

	if logging.Mode != logging.Nil {
		log.Println(`
#####################
## shellfish coord ##
#####################`,
		)
	}
	var t time.Time
	if logging.Mode == logging.Performance {
		t = time.Now()
	}

	intCols, _, err := catalog.ParseCols(stdin, []int{0, 1}, []int{})
	if err != nil {
		return nil, err
	}
	ids, snaps := intCols[0], intCols[1]

	if len(ids) == 0 {
		return nil, fmt.Errorf("In input IDs.")
	}

	vars := halo.NewVarColumns(
		gConfig.HaloValueNames, gConfig.HaloValueColumns,
	)

	buf, err := getVectorBuffer(
		e.ParticleCatalog(snaps[0], 0),
		gConfig.SnapshotType, gConfig.Endianness,
	)
	if err != nil {
		return nil, err
	}

	xs, ys, zs, rs, err := readHaloCoords(ids, snaps, vars, buf, e)
	if err != nil {
		return nil, err
	}

	lines := catalog.FormatCols(
		[][]int{ids, snaps}, [][]float64{xs, ys, zs, rs},
		[]int{0, 1, 2, 3, 4, 5},
	)

	cString := catalog.CommentString(
		[]string{"ID", "Snapshot"},
		[]string{"X [cMpc/h]", "Y [cMpc/h]", "Z [cMpc/h]", "R200m [cMpc/h]"},
		[]int{0, 1, 2, 3, 4, 5},
		[]int{1, 1, 1, 1, 1, 1},
	)

	if logging.Mode == logging.Performance {
		log.Printf("Time: %s", time.Since(t).String())
		log.Printf("Memory:\n%s", logging.MemString())
	}

	return append([]string{cString}, lines...), nil
}

func readHaloCoords(
	ids, snaps []int, vars *halo.VarColumns,
	buf io.VectorBuffer, e *env.Environment,
) (xs, ys, zs, rs []float64, err error) {
	snapBins, idxBins := binBySnap(snaps, ids)

	xs = make([]float64, len(ids))
	ys = make([]float64, len(ids))
	zs = make([]float64, len(ids))
	rs = make([]float64, len(ids))

	for snap, _ := range snapBins {
		if snap == -1 {
			continue
		}
		snapIDs := snapBins[snap]
		idxs := idxBins[snap]

		_, vals, err := memo.ReadRockstar(
			snap, []string{"X", "Y", "Z", "R200m", "Scale"}, snapIDs, vars, buf, e,
		)
		
		sxs, sys, szs, srs := vals[0], vals[1], vals[2], vals[3]
		
		if err != nil {
			return nil, nil, nil, nil, err
		}

		for i, idx := range idxs {
			xs[idx] = sxs[i]
			ys[idx] = sys[i]
			zs[idx] = szs[i]
			rs[idx] = srs[i]
		}
	}

	return xs, ys, zs, rs, nil
}
