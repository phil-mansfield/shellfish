package env

import (
	"fmt"
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

	Rockstar HaloType = iota
	NilHalo

	ConsistentTrees TreeType = iota
	NilTree
)

type ParticleInfo struct {
	SnapshotFormat          string
	MemoDir                 string

	SnapshotFormatMeanings  []string
	ScaleFactorFile         string
	FormatMins, FormatMaxes []int64
	SnapMin, SnapMax        int64
}

type HaloInfo struct {
	HaloDir, TreeDir string
	SnapMin, SnapMax int64
}

type Environment struct {
	Catalogs
	Halos
	MemoDir string
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
				snaps= len(cols[i])
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

	out := make([][][]interface{}, snaps)

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