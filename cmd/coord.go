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
# Values are the names of the values you want to write to an output catalog.
# The default order is the one which is needed by Shellfish. Any other order
# would correspond to a catalog which is for your personal use only.
Values = X, Y, Z, R200m
`
}

func (config *CoordConfig) ReadConfig(fname string, flags []string) error {
	vars := parse.NewConfigVars("coord.config")
	vars.Strings(&config.values, "Values", []string{"X", "Y", "Z", "R200m"})

	if fname == "" {
		if len(flags) == 0 {
			return nil
		}
		return parse.ReadFlags(flags, vars)
	}
	if err := parse.ReadConfig(fname, vars); err != nil {
		return err
	}
	return parse.ReadFlags(flags, vars)
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
	gConfig *GlobalConfig, e *env.Environment, stdin []byte,
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

	intCols, _, err := catalog.Parse(stdin, []int{0, 1}, []int{})
	if err != nil {
		return nil, err
	}
	ids, snaps := intCols[0], intCols[1]

	if len(ids) == 0 {
		return nil, fmt.Errorf("In input IDs.")
	}

	vars := halo.NewVarColumns(
		gConfig.HaloValueNames, gConfig.HaloValueColumns,
		gConfig.HaloRadiusUnits,
	)
	if err := config.validate(vars); err != nil {
		return nil, err
	}

	buf, err := getVectorBuffer(
		e.ParticleCatalog(snaps[0], 0), gConfig,
	)
	if err != nil {
		return nil, err
	}

	cols, err := readHaloCoords(
		ids, snaps, config.values, vars, buf, e, gConfig,
	)
	if err != nil {
		return nil, err
	}

	icols := [][]int{ids, snaps}
	fcols := [][]float64{}
	icolOrder := []int{0, 1}
	fcolOrder := []int{}
	
	intNum := 0
	for _, valueName := range config.values {
		switch valueName {
		case "R200m", "R200c", "R500c", "R2500c":
		default:
			j := findString(valueName, gConfig.HaloValueNames)
			comment := gConfig.HaloValueComments[j]
			if isIntType(comment) { intNum++ }
		}
	}

	for i, valueName := range config.values {
		var comment string
		switch valueName {
		case "R200m", "R200c", "R500c", "R2500c":
			comment = "Mpc/h"
		default:
			j := findString(valueName, gConfig.HaloValueNames)
			comment = gConfig.HaloValueComments[j]
		}

		if isIntType(comment) {
			fcol := cols[i]
			icol := make([]int, len(fcol))
			for i := range icol { icol[i] = int(fcol[i]) }
			icols = append(icols, icol)
			icolOrder = append(icolOrder, i + 2)
		} else {
			fcols = append(fcols, cols[i])
			fcolOrder = append(fcolOrder, i + 2)
		}
	}

	colOrder := append(icolOrder, fcolOrder...)
	lines := catalog.FormatCols(icols, fcols, colOrder)

	cString := makeCommentString(gConfig, config)

	if logging.Mode == logging.Performance {
		log.Printf("Time: %s", time.Since(t).String())
		log.Printf("Memory:\n%s", logging.MemString())
	}

	return append([]string{cString}, lines...), nil
}

func isIntType(comment string) bool {
	return comment == "int" || comment == "\"int\""
}

func makeCommentString(gConfig *GlobalConfig, config *CoordConfig) string {
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
			gConfig.HaloValueComments[j] == "\"\"" ||
			isIntType(gConfig.HaloValueComments[j]) {

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
	if len(snaps) == 0 { return nil, nil }

	snapBins, idxBins := binBySnap(snaps, ids)

	cols = make([][]float64, len(valNames))
	for i := range cols {
		cols[i] = make([]float64, len(ids))
	}

	for snap, _ := range snapBins {
		if snap == -1 {
			continue
		}

		hds, _, err := memo.ReadHeaders(snaps[0], buf, e)
		if err != nil { return nil, err }
		cosmo := &hds[0].Cosmo

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
			case "X", "Y", "Z":
				ucf := halo.UnitConversionFactor(
					gConfig.HaloPositionUnits, cosmo,
				)
				for j := range scols[i] {
					scols[i][j] *= ucf
				}
			case "R200m", "R200c", "R500c", "Rs" :
				ucf := halo.UnitConversionFactor(
					gConfig.HaloRadiusUnits, cosmo,
				)
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
