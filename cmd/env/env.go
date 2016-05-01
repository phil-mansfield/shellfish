package env

import (
	"fmt"
	"io/ioutil"
	"strings"
	"strconv"
)

///////////
// Types //
///////////

type (
	CatalogType int
	HaloType int
	TreeType int
)

const (
	Gotetra CatalogType = iota
	LGadget2
	ARTIO

	Rockstar HaloType = iota
	NilHalo

	ConsistentTrees TreeType = iota
	NilTree
)

type Environment struct {
	Catalogs
	Halos
	MemoDir string
}

//////////////////
// Info structs //
//////////////////

type ParticleInfo struct {
	SnapshotFormat         string
	MemoDir                string

	SnapshotFormatMeanings []string
	ScaleFactorFile        string
	BlockMins, BlockMaxes  []int64
	SnapMin, SnapMax       int64
}

type HaloInfo struct {
	HaloDir, TreeDir   string
	HSnapMin, HSnapMax int64
}

func (info *ParticleInfo) GetColumn(
	i int,
) (col []interface{}, snapAligned bool, err error) {

	fmt.Println(i)
	
	m := info.SnapshotFormatMeanings[i]
	switch {
	case m == "ScaleFactor":
		bs, err := ioutil.ReadFile(info.ScaleFactorFile)
		if err != nil { return nil, false, err }
		text := string(bs)
		lines := strings.Split(text, "\n")

		out := []string{}
		for i := range lines {
			line := strings.Trim(lines[i], " ")
			if len(line) != 0 { out = append(out, line) }
		}

		if len(out) != int(info.SnapMax - info.SnapMin) + 1 {
			return nil, false, fmt.Errorf(
				"%s has %d non-empty lines, but SnapMax = %d and SnapMin = %d.",
				info.ScaleFactorFile, len(out), info.SnapMax, info.SnapMin,
			)
		}

		return anonymize(out), true, nil

	case m == "Snapshot":
		out := make([]int, int(info.SnapMax - info.SnapMax) + 1)
		for i := range out { out[i] = i + int(info.SnapMin) }
		return anonymize(out), true, nil

	case len(m) > 5 && m[:5] == "Block":
		idx := 0
		if len(m) > 5 {
			var err error
			idx, err = strconv.Atoi(m[5:])
			if err != nil { return nil, false, err }
		}
		out := make([]int, info.BlockMaxes[idx] - info.BlockMins[idx] + 1)
		for i := range out { out[i] = i + int(info.BlockMins[idx]) + 1 }
		return anonymize(out), false, nil
	}
	panic("Impossible")
}

func anonymize(col interface{}) []interface{} {
	switch xs := col.(type) {
	case []int:
		out := make([]interface{}, len(xs))
		for i := range out { out[i] = xs[i] }
		return out
	case []string:
		out := make([]interface{}, len(xs))
		for i := range out { out[i] = xs[i] }
		return out
	}
	panic("Unknown type.")
}

//////////////
// Catalogs //
//////////////

type Catalogs struct {
	CatalogType
	snapMin int
	names [][]string
}


func (cat *Catalogs) Blocks() int {
	return len(cat.names[0])
}

func (cat *Catalogs) ParticleCatalog(snap, block int) string {
	return cat.names[snap - cat.snapMin][block]
}

///////////
// Halos //
///////////

type Halos struct {
	HaloType
	TreeType
	snapMin int
	snapOffset int
	names []string
}

func (h *Halos) HaloCatalog(snap int) string {
	return h.names[snap - h.snapMin]
}

func (h *Halos) SnapOffset() int {
	return h.snapOffset
}

func (h *Halos) InitNilHalo(dir string, snapMin, snapMax int64) error {
	h.HaloType = NilHalo
	h.TreeType = NilTree

	return nil
}

//////////////////////////////////
// Format Argument Manipulation //
//////////////////////////////////

// interleave is a bit hard to explain. It will be easier to understand
// what it's supposed to do by reading the source than it would be to read prose
// trying to explain it precisely.
func interleave(cols [][]interface{}, snapAligned []bool) [][][]interface{} {
	// Check that all snap-aligned columns are same length.
	snaps, nonSnaps := -1, 1
	for i := range cols {
		if snapAligned[i] {
			if snaps == -1 {
				snaps = len(cols[i])
			} else if snaps != len(cols[i]) {
				// This is safe to think of as an internal error.
				panic(fmt.Sprintf("Column %d is snapAligned, but has height " +
					"%d instead of %d.", i, len(cols[i]), snaps))
			}
		} else {
			nonSnaps *= len(cols[i])
		}
	}

	// This is safe to think of as an internal error.
	if snaps == 1 { panic("All values in snapAligned are false.") }

	out := [][][]interface{}{}

	for snap := 0; snap < snaps; snap++ {
		snapRows := make([][]interface{}, nonSnaps)
		for i := range snapRows { snapRows[i] = make([]interface{}, len(cols)) }

		divSize, modSize := 1, 1

		for col := range cols {
			if snapAligned[col] {
				for row := range snapRows {
					snapRows[row][col] = cols[col][snap]
				}
			} else {
				divSize *= modSize
				modSize = len(cols[col])

				for row := range snapRows {
					snapRows[row][col] = cols[col][(row / divSize) % modSize]
				}
			}
		}

		out = append(out, snapRows)
	}

	return out
}
