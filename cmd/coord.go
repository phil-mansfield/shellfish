package cmd

import (
	"fmt"

	"github.com/phil-mansfield/shellfish/cmd/catalog"
	"github.com/phil-mansfield/shellfish/cmd/env"
	"github.com/phil-mansfield/shellfish/cmd/halo"
	"github.com/phil-mansfield/shellfish/cmd/memo"
	"github.com/phil-mansfield/shellfish/io"
)

type CoordConfig struct {
}

var _ Mode = &CoordConfig{}

func (config *CoordConfig) ExampleConfig() string { return "" }

func (config *CoordConfig) ReadConfig(fname string) error { return nil }

func (config *CoordConfig) validate() error { return nil }

func (config *CoordConfig) Run(
	flags []string, gConfig *GlobalConfig, e *env.Environment, stdin []string,
) ([]string, error) {
	intCols, _, err := catalog.ParseCols(stdin, []int{0, 1}, []int{})
	if err != nil {
		return nil, err
	}
	ids, snaps := intCols[0], intCols[1]

	if len(ids) == 0 {
		return nil, fmt.Errorf("In input IDs.")
	}

	vars := &halo.VarColumns{
		ID:    int(gConfig.HaloIDColumn),
		X:     int(gConfig.HaloPositionColumns[0]),
		Y:     int(gConfig.HaloPositionColumns[1]),
		Z:     int(gConfig.HaloPositionColumns[2]),
		M200m: int(gConfig.HaloM200mColumn),
	}

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
		[]string{"X", "Y", "Z", "R200m"},
		[]int{0, 1, 2, 3, 4, 5},
		[]int{1, 1, 1, 1, 1, 1},
	)

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

		_, sxs, sys, szs, _, srs, err := memo.ReadRockstar(
			snap, snapIDs, vars, buf, e,
		)

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
