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
	values []string
}

var _ Mode = &CoordConfig{}

func (config *CoordConfig) ExampleConfig() string {
	return`[config.coord]
Values = X, Y, Z, R200m
`
}

func (config *CoordConfig) ReadConfig(fname string) error {
	vars := parse.NewConfigVars("coord.config")
	vars.Strings(&config.values, "Values", []string{"X", "Y", "Z", "R200m"})

	if fname == "" {
		return nil
	}

	return parse.ReadConfig(fname, vars)
}

func (config *CoordConfig) validate(vars *halo.VarColumns) error {
	for _, val := range config.values {
		if _, ok := vars.ColumnLookup[val]; !ok {
			return fmt.Errorf(
				"Value '%s' requested by coord mode, but isn't " +
				"in HaloVaueNames.", val,
			)
		}
	}
	return nil
}

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
	if err := config.validate(vars); err != nil {
		return nil, err
	}

	buf, err := getVectorBuffer(
		e.ParticleCatalog(snaps[0], 0),
		gConfig.SnapshotType, gConfig.Endianness,
	)
	if err != nil {
		return nil, err
	}

	columns, err := readHaloCoords(
		ids, snaps, config.values, vars, buf, e, gConfig,
	)
	if err != nil {
		return nil, err
	}

	colOrder := make([]int, 2 + len(columns))
	for i := range colOrder {
		colOrder[i] = i
	}
	lines := catalog.FormatCols([][]int{ids, snaps}, columns, colOrder)

	cString := makeCommentString(gConfig, config)

	if logging.Mode == logging.Performance {
		log.Printf("Time: %s", time.Since(t).String())
		log.Printf("Memory:\n%s", logging.MemString())
	}

	return append([]string{cString}, lines...), nil
}

func makeCommentString(gConfig *GlobalConfig, config *CoordConfig) string {
	//colNames := make([]string, 2 + len(config.values))
	//colNames[0], colNames[1] = "ID", "Snap"
	colNames := make([]string, len(config.values))
	for i := 0; i < len(config.values); i++ {
		switch config.values[i] {
		case "R200m", "R200c", "R500c", "Rs":
			colNames[i] = fmt.Sprintf(
				"%s [%s]", config.values[i], gConfig.HaloPositionUnits,
			)
			continue
		}
		
		j := findString(config.values[i], gConfig.HaloValueNames)
		if gConfig.HaloValueComments[j] == "" ||
			gConfig.HaloValueComments[j] == "\"\"" {
			colNames[i] = config.values[i]
		} else {
			colNames[i] = fmt.Sprintf(
				"%s [%s]", config.values[i], gConfig.HaloValueComments[j],
			)
		}
	}

	colOrder := make([]int, 2 + len(config.values))
	colSizes := make([]int, 2 + len(config.values))
	for i := range colOrder {
		colOrder[i], colSizes[i] = i, 1
	}

	return catalog.CommentString(
		[]string{"ID", "Snapshot"}, colNames, colOrder, colSizes,
	)
}

func findString(x string, xs []string) int {
	for i := range xs {
		if xs[i] == x { return i }
	}
	panic("Impossible")
}



func readHaloCoords(
	ids, snaps []int, valNames []string, vars *halo.VarColumns,
	buf io.VectorBuffer, e *env.Environment, gConfig *GlobalConfig,
) (cols [][]float64, err error) {
	snapBins, idxBins := binBySnap(snaps, ids)

	cols = make([][]float64, len(valNames))
	for i := range cols {
		cols[i] = make([]float64, len(ids))
	}

	for snap, _ := range snapBins {
		if snap == -1 {
			continue
		}
		snapIDs := snapBins[snap]
		idxs := idxBins[snap]

		_, scols, err := memo.ReadRockstar(
			snap, valNames, snapIDs, vars, buf, e,
		)

		if err != nil {
			return nil, err
		}

		for i := range valNames {
			switch valNames[i] {
			// All positional variables
			case "R200m", "R200c", "R500c", "Rs", "X", "Y", "Z":
				ucf := UnitConversionFactor(gConfig.HaloPositionUnits)
				for j := range scols[i] {
					scols[i][j] *= ucf
				}
			}
		}

		for i, idx := range idxs {
			for j := range cols {
				cols[j][idx] = scols[j][i]
			}
		}
	}

	return cols, nil
}

// UnitConversionFactor returns the multiplicative factor needed to convert
// the given units into cMpc/h.
func UnitConversionFactor(unitStr string) float64 {
	switch unitStr {
	case "cMpc/h": return 1.0
	case "ckpc/h": return 1e-3
	default: panic(fmt.Sprintf("Unrecognized unit string '%s'", unitStr))
	}
}