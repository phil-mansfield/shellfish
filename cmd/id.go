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

const finderCells = 150

// IDConfig contains the configuration fileds for the 'id' mode of the shellfish
// tool.
type IDConfig struct {
	idType                     string
	ids                        []int64
	idStart, idEnd, snap, mult int64
	m200mMax, m200mMin         float64

	exclusionStrategy          string
	exclusionRadiusMult        float64
}

var _ Mode = &IDConfig{}

// ExampleConfig creates an example id.config file.
func (config *IDConfig) ExampleConfig() string {
	return `[id.config]
#####################
## Required Fields ##
#####################

# Index of the snapshot to be analyzed.
Snap = 100

# List of IDs to analyze.
IDs = 10, 11, 12, 13, 14

#####################
## Optional Fields ##
#####################

# IDType indicates what the input IDs correspond to. It can be set to the
# following modes:
# halo-id - The numeric IDs given in the halo catalog.
# m200m   - The rank of the halos when sorted by M200m.
#
# Defaults to m200m if not set.
# IDType = m200m

# An alternative way of specifying IDs is to select start and end (inclusive)
# ID values. If the IDs variable is not set, both of these values must be set.
#
# IDStart = 10
# IDEnd = 15

# Yet another alternative way to select IDs is to specify the starting and
# ending mass range (units are M_sun/h). IDType, IDs, IDStart, and IDEnd will
# be ignored if these variables are set.
#
# M200mMin = 1e12
# M200mMax = 1e13

# ExclusionStrategy determines how to exclude IDs from the given set. This is
# useful because splashback shells are not particularly meaningful for
# subhalos. It can be set to the following modes:
# none      - No halos are removed
# subhalo   - Halos flagged as subhalos in the catalog are removed (not yet
#             implemented)
# overlap   - Halos which have an R200m shell that overlaps with a larger halo's
#             R200m shell are removed
# neighbor  - Instead of removing halos, all neighboring halos within
#             ExclusionRadiusMult*R200m are added to the list.
#
# ExclusionStrategy defaults to overlap if not set.
#
# ExclusionStrategy = overlap

# ExclusionRadiusMult is a multiplier of R200m applied for the sake of
# determining exclusions. ExclusionRadiusMult defaults to 0.8 if not set.
#
# ExclusionRadiusMult = 0.8

# Mult is the number of times a given ID should be repeated. This is most useful
# if you want to estimate the scatter in shell measurements for halos with a
# given set of shell parameters.
#
# Mult defaults to 1 if not set.
#
# Mult = 1`
}

// ReadConfig reads in an id.config file into config.
func (config *IDConfig) ReadConfig(fname string, flags []string) error {

	vars := parse.NewConfigVars("id.config")
	vars.String(&config.idType, "IDType", "m200m")
	vars.Ints(&config.ids, "IDs", []int64{})
	vars.Int(&config.idStart, "IDStart", -1)
	vars.Int(&config.idEnd, "IDEnd", -1)
	vars.Int(&config.mult, "Mult", 1)
	vars.Int(&config.snap, "Snap", -1)
	vars.String(&config.exclusionStrategy, "ExclusionStrategy", "overlap")
	vars.Float(&config.exclusionRadiusMult, "ExclusionRadiusMult", 1)
	vars.Float(&config.m200mMax, "M200mMax", 0)
	vars.Float(&config.m200mMin, "M200mMin", 0)

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
	
	return config.validate()
}

// validate checks whether all the fields of config are valid.
func (config *IDConfig) validate() error {
	switch config.idType {
	case "halo-id", "m200m":
	default:
		return fmt.Errorf("The 'IDType' variable is set to '%s', which I "+
			"don't recognize.", config.idType)
	}

	switch config.exclusionStrategy {
	case "none", "subhalo", "neighbor":
	case "overlap":
		if config.exclusionRadiusMult <= 0 {
			return fmt.Errorf("The 'ExclusionRadiusMult' varaible is set to "+
				"%g, but it needs to be positive.", config.exclusionRadiusMult)
		}
	default:
		return fmt.Errorf("The 'ExclusionStrategy' variable is set to '%s', "+
			"which I don't recognize.", config.exclusionStrategy)
	}

	switch {
	case config.snap == -1:
		return fmt.Errorf("'Snap' variable not set.")
	case config.snap < 0:
		return fmt.Errorf("'Snap' variable set to %d.", config.snap)
	}

	if config.mult <= 0 {
		return fmt.Errorf("'Mult' variable set to %d", config.mult)
	}

	switch {
	case config.m200mMax < 0:
		return fmt.Errorf("M200mStart set to %g", config.m200mMax)
	case config.m200mMax > config.m200mMin:
		fmt.Errorf("M200mEnd smaller than M200mStart")
	}

	return nil
}

// Run executes the ID mode of shellfish tool.
func (config *IDConfig) Run(
	gConfig *GlobalConfig, e *env.Environment, stdin []byte,
) ([]string, error) {
	
	if logging.Mode != logging.Nil {
		log.Println(`
##################
## shellfish id ##
##################`,
		)
	}
	var t time.Time
	if logging.Mode == logging.Performance {
		t = time.Now()
	}
	if config.snap == -1 {
		return nil, fmt.Errorf("Either no id.config file was provided or " +
			"the 'Snap' variable wasn't set.")
	}
	if config.snap < gConfig.SnapMin || config.snap > gConfig.SnapMax {
		return nil, fmt.Errorf("'Snap' = %d, but 'SnapMin' = %d and "+
			"'SnapMax = %d'", config.snap, gConfig.SnapMin, gConfig.SnapMax)
	}
	// Get IDs and snapshots

	vars := halo.NewVarColumns(
		gConfig.HaloValueNames, gConfig.HaloValueColumns,
		gConfig.HaloRadiusUnits,
	)

	if config.m200mMax > 0 {
		getMassIDRange(gConfig, e, config, vars)
	}

	// This is kind of a hack to deal with the case where there are no IDs in
	// the specified mass range.
	var (
		rawIds []int
		err error
	)
	if config.idStart <= config.idEnd {
		rawIds, err = getIDs(config.idStart, config.idEnd, config.ids, stdin)
		if err != nil {
			return nil, err
		} else if len(rawIds) == 0 {
			return nil, nil
		}
	}

	var (
		ids, snaps []int
		buf        io.VectorBuffer
	)
	switch config.idType {
	case "halo-id":
		snaps = make([]int, len(rawIds))
		for i := range snaps {
			snaps[i] = int(config.snap)
		}
		ids = rawIds

		var err error
		buf, err = getVectorBuffer(
			e.ParticleCatalog(snaps[0], 0), gConfig,
		)
		if err != nil {
			return nil, err
		}
	case "m200m":
		snaps = make([]int, len(rawIds))
		for i := range snaps {
			snaps[i] = int(config.snap)
		}

		var err error
		buf, err = getVectorBuffer(
			e.ParticleCatalog(snaps[0], 0), gConfig,
		)
		if err != nil {
			return nil, err
		}

		ids, err = convertSortedIDs(rawIds, int(config.snap), vars, buf, e)
		if err != nil {
			return nil, err
		}
	default:
		panic("Impossible")
	}

	// Tag subhalos, if neccessary.
	exclude := make([]bool, len(ids))
	switch config.exclusionStrategy {
	case "none":
	case "subhalo":
		panic("subhalo is not implemented")
	case "neighbor":
		ids, snaps, err = readSubIDs(
			ids, snaps, vars, buf, e, config, gConfig,
		)
		if err != nil {
			return nil, err
		}

		exclude = make([]bool, len(ids))
	case "overlap":
		var err error
		exclude, err = findOverlapSubs(
			ids, snaps, vars, buf, e, config, gConfig,
		)
		if err != nil {
			return nil, err
		}
	}
	
	// Generate lines
	intCols := [][]int{ids, snaps}
	floatCols := [][]float64{}
	colOrder := []int{0, 1}
	lines := catalog.FormatCols(intCols, floatCols, colOrder)

	// Filter
	fLines := []string{}
	for i := range lines {
		if !exclude[i] {
			fLines = append(fLines, lines[i])
		}
	}

	// Multiply
	mLines := []string{}
	for i := range fLines {
		for j := 0; j < int(config.mult); j++ {
			mLines = append(mLines, fLines[i])
		}
	}

	cString := catalog.CommentString(
		[]string{"ID", "Snapshot"}, []string{}, []int{0, 1}, []int{1, 1},
	)
	mLines = append([]string{cString}, mLines...)

	if logging.Mode == logging.Performance {
		log.Printf("Time: %s", time.Since(t).String())
		log.Printf("Memory:\n%s", logging.MemString())
	}

	return mLines, nil
}

// getMassRange updates config so that it points to an ID range that
// corresponds to [M200mMin, M200mMax]
func getMassIDRange(
	gConfig *GlobalConfig, e *env.Environment, config *IDConfig,
	vars *halo.VarColumns,
) error {
	config.idType = "m200m"
	buf, err := getVectorBuffer(
		e.ParticleCatalog(int(config.snap), 0), gConfig,
	)

	if err != nil { return err }

	rids, err := memo.ReadSortedRockstarIDs(
		int(config.snap), -1, "M200m", vars, buf, e,
	)
	if err != nil { return err }
	_, vals, err := memo.ReadRockstar(
		int(config.snap), []string{"M200m"}, rids, vars, buf, e,
	)
	if err != nil { return err }

	ms := vals[0]
	config.idStart, config.idEnd = -1, -1
	for i := range ms {
		if config.idStart == -1 && config.m200mMax >= ms[i] {
			config.idStart = int64(i)
		}
		if config.m200mMin > ms[i] {
			config.idEnd = int64(i-1)
			return nil
		}
	}

	config.idEnd = int64(len(ms) - 1)
	return nil
}


func getIDs(idStart, idEnd int64, ids []int64, stdin []byte) ([]int, error) {
	if idStart != -1 {
		out := make([]int, idEnd-idStart)
		for i := range out {
			out[i] = int(idStart) + i
		}
		return out, nil
	} else if len(ids) > 0 {
		out := make([]int, len(ids))
		for i := range out {
			out[i] = int(ids[i])
		}
		return out, nil
	} else {
		intCols, _, err := catalog.Parse(stdin, []int{0}, []int{})
		if err != nil {
			return nil, err
		}
		return intCols[0], nil
	}
}

func convertSortedIDs(
	rawIDs []int, snap int, vars *halo.VarColumns,
	buf io.VectorBuffer, e *env.Environment,
) ([]int, error) {
	maxID := 0
	for _, id := range rawIDs {
		if id > maxID {
			maxID = id
		}
	}

	rids, err := memo.ReadSortedRockstarIDs(snap, maxID, "M200m", vars, buf, e)
	if err != nil {
		return nil, err
	}

	ids := make([]int, len(rawIDs))
	for i := range ids {
		ids[i] = rids[rawIDs[i]]
	}
	return ids, nil
}

func findOverlapSubs(
	rawIDs, snaps []int, vars *halo.VarColumns,
	buf io.VectorBuffer, e *env.Environment, config *IDConfig,
	gConfig *GlobalConfig,
) ([]bool, error) {
	isSub := make([]bool, len(rawIDs))

	// Group by snapshot.
	snapGroups := make(map[int][]int)
	groupIdxs := make(map[int][]int)
	for i, id := range rawIDs {
		snap := snaps[i]
		snapGroups[snap] = append(snapGroups[snap], id)
		groupIdxs[snap] = append(groupIdxs[snap], i)
	}

	// Load each snapshot.
	hds, _, err := memo.ReadHeaders(snaps[0], buf, e)
	if err != nil {
		return nil, err
	}
	hd := hds[0]
	cosmo := &hd.Cosmo

	for snap, group := range snapGroups {
		rids, err := memo.ReadSortedRockstarIDs(
			snap, -1, "M200m", vars, buf, e,
		)
		if err != nil {
			return nil, err
		}
		_, vals, err := memo.ReadRockstar(
			snap, []string{"X", "Y", "Z", "R200m"}, rids, vars, buf, e,
		)
		if err != nil {
			return nil, err
		}
		xs, ys, zs, rs := vals[0], vals[1], vals[2], vals[3]
		rucf := halo.UnitConversionFactor(gConfig.HaloRadiusUnits, cosmo)
		pucf := halo.UnitConversionFactor(gConfig.HaloPositionUnits, cosmo)
		for i := range rs {
			rs[i] *= rucf // Rusev, crush!
			xs[i] *= pucf
			ys[i] *= pucf
			zs[i] *= pucf
		}
		
		g := halo.NewGrid(finderCells, hd.TotalWidth, len(xs))
		g.Insert(xs, ys, zs)
		sf := halo.NewSubhaloFinder(g)
		sf.FindSubhalos(xs, ys, zs, rs, config.exclusionRadiusMult)
		
		for i, id := range group {
			origIdx := groupIdxs[snap][i]
			// TODO: Holy linear search, batman! Fix this.
			for j, checkID := range rids {
				if checkID == id {
					isSub[origIdx] = sf.HostCount(j) > 0
					break
				} else if j == len(rids)-1 {
					return nil, fmt.Errorf("ID %d not in halo list.", id)
				}
			}
		}
	}
	return isSub, nil
}

func readSubIDs(
	ids, snaps []int, vars *halo.VarColumns,
	buf io.VectorBuffer, e *env.Environment,
	config *IDConfig, gConfig *GlobalConfig,
) (
	sIDs, sSnaps []int, err error,
) {
	subIDs := make([][]int, len(ids))

	snapGroups := make(map[int][]int)
	groupIdxs := make(map[int][]int)
	for i, id := range ids {
		snap := snaps[i]
		snapGroups[snap] = append(snapGroups[snap], id)
		groupIdxs[snap] = append(groupIdxs[snap], i)
	}
	
	// Load each snapshot.
	hds, _, err := memo.ReadHeaders(snaps[0], buf, e)
	if err != nil {
		return nil, nil, err
	}
	hd := hds[0]
	cosmo := &hd.Cosmo

	for snap, group := range snapGroups {
		rids, err := memo.ReadSortedRockstarIDs(
			snap, -1, "M200m", vars, buf, e,
		)
		if err != nil {
			return nil, nil, err
		}
		_, vals, err := memo.ReadRockstar(
			snap, []string{"X", "Y", "Z", "R200m"}, rids, vars, buf, e,
		)
		xs, ys, zs, rs := vals[0], vals[1], vals[2], vals[3]
		rucf := halo.UnitConversionFactor(gConfig.HaloRadiusUnits, cosmo)
		pucf := halo.UnitConversionFactor(gConfig.HaloPositionUnits, cosmo)
		for i := range rs {
			rs[i] *= rucf
			xs[i] *= pucf
			ys[i] *= pucf
			zs[i] *= pucf
		}
		
		g := halo.NewGrid(finderCells, hd.TotalWidth, len(xs))
		g.Insert(xs, ys, zs)
		sf := halo.NewSubhaloFinder(g)
		sf.FindSubhalos(xs, ys, zs, rs, config.exclusionRadiusMult)
		
		f := newIntFinder(rids)
		idxs := groupIdxs[snap]
		for i, id := range group {
			idx := idxs[i]
			sIdx, ok := f.find(id)
			if !ok {
				return nil, nil, fmt.Errorf("Could not find ID %d.", id)
			}
			
			subMassIDs := sf.Subhalos(sIdx)

			subIDs[idx] = make([]int, len(subMassIDs))
			for i := range subMassIDs {
				subIDs[idx][i] = rids[subMassIDs[i]]
			}
		}
	}
	sIDs, sSnaps = []int{}, []int{}
	for i := range subIDs {
		sIDs = append(sIDs, ids[i])
		sSnaps = append(sSnaps, snaps[i])
		for _, id := range subIDs[i] {
			sIDs = append(sIDs, id)
			sSnaps = append(sSnaps, snaps[i])
		}
	}

	return sIDs, sSnaps, nil
}

// A quick generic wrapper for doing those one-to-one mappings I need to do so
// often. Written like this so the backend can be swapped out easily.
type intFinder struct {
	m map[int]int
}

// NewIntFinder creates a new IntFinder struct for a given slice of Rockstar
// IDs.
func newIntFinder(rids []int) intFinder {
	f := intFinder{}
	f.m = make(map[int]int)
	for i, rid := range rids { f.m[rid] = i }
	return f
}

// Find returns the index which the given ID corresponds to and true if the
// ID is in the finder. Otherwise, false is returned.
func (f intFinder) find(rid int) (int, bool) {
	line, ok := f.m[rid]
	return line, ok
}
