package memo

import (
	"encoding/binary"
	"fmt"
	"os"
	"path"
	"sort"

	"github.com/phil-mansfield/shellfish/cmd/env"

	"github.com/phil-mansfield/shellfish/cmd/halo"
	"github.com/phil-mansfield/shellfish/io"
)

// TODO: rewrite the six return values as a alice of slices.

const (
	rockstarMemoDir = "rockstar"
	rockstarMemoFile = "halo_%d.dat"
	rockstarShortMemoFile = "halo_short_%d.dat"
	rockstarShortMemoNum = 10 * 1000

	headerMemoFile = "hd_snap%d.dat"
)

// ReadSortedRockstarIDs returns a slice of IDs corresponding to the highest
// values of some quantity in a particular snapshot. maxID is the number of
// halos to return.
func ReadSortedRockstarIDs(
	snap, maxID int, vars *halo.VarColumns, e *env.Environment,
) ([]int, error) {
	dir := path.Join(e.MemoDir, rockstarMemoDir)
	if _, err := os.Stat(dir); err != nil {
		err = os.Mkdir(dir, 0777)
		if err != nil { return nil, err }
	}

	var (
		ids []int
		ms []float64
		err error
	)
	if maxID >= rockstarShortMemoNum || maxID == -1 {
		file := path.Join(dir, fmt.Sprintf(rockstarMemoFile, snap))
		ids, _, _, _, ms, _, err = readRockstar(
			file, -1, snap, nil, vars, e,
		)
		if err != nil { return nil, err }
	} else {
		file := path.Join(dir, fmt.Sprintf(rockstarShortMemoFile, snap))
		ids, _, _, _, ms, _, err = readRockstar(
			file, rockstarShortMemoNum, snap, nil, vars, e,
		)
		if err != nil { return nil, err }
	}

	if len(ids) < maxID {
		return nil, fmt.Errorf(
			"ID %d too large for snapshot %d", maxID, snap,
		)
	}

	sortRockstar(ids, ms)
	if maxID == -1 { return ids, nil }
	return ids[:maxID+1], nil
}

type massSet struct {
	ids []int
	ms []float64
}

func (set massSet) Len() int { return len(set.ids) }
// We're reverse sorting.
func (set massSet) Less(i, j int) bool { return set.ms[i] > set.ms[j] }
func (set massSet) Swap(i, j int) {
	set.ms[i], set.ms[j] = set.ms[j], set.ms[i]
	set.ids[i], set.ids[j] = set.ids[j], set.ids[i]
}

func sortRockstar(ids []int, ms []float64) {
	set := massSet{ ids, ms }
	sort.Sort(set)
}

// This function does fairly large heap allocations even when it doesn't need
// to. Consider passing it a buffer.
func ReadRockstar(
	snap int, ids []int, vars *halo.VarColumns, e *env.Environment,
) (outIDs []int, xs, ys, zs, ms, rs []float64, err error) {
	// Find binFile.
	dir := path.Join(e.MemoDir, rockstarMemoDir)
	if _, err := os.Stat(dir); err != nil {
		err = os.Mkdir(dir, 0777)
		if err != nil { return nil, nil, nil, nil, nil, nil, err }
	}

	binFile := path.Join(dir, fmt.Sprintf(rockstarMemoFile, snap))
	shortBinFile := path.Join(dir, fmt.Sprintf(rockstarShortMemoFile, snap))

	// This wastes a read the first time it's called. You need to decide if you
	// care. (Answer: probably.)
	outIDs, xs, ys, zs, ms, rs, err = readRockstar(
		shortBinFile, rockstarShortMemoNum, snap, ids, vars, e,
	)
	// TODO: Fix error handling here.
	if err == nil { return outIDs, xs, ys, zs, ms, rs, err }
	outIDs, xs, ys, zs, ms, rs, err = readRockstar(
		binFile, -1, snap, ids, vars, e,
	)
	return outIDs, xs, ys, zs, ms, rs, nil
}

func readRockstar(
	binFile string, n, snap int, ids []int,
	vars *halo.VarColumns, e *env.Environment,
) (outIDs []int, xs, ys, zs, ms, rs []float64, err error) {
	hds, _, err := ReadHeaders(snap, e)
	if err != nil { return nil, nil, nil, nil, nil, nil, err }
	hd := &hds[0]

	// If binFile doesn't exist, create it.
	if _, err := os.Stat(binFile); err != nil {
		if n == -1 {
			err = halo.RockstarConvert(
				e.HaloCatalog(snap), binFile, vars, &hd.Cosmo,
			)
			if err != nil { return nil, nil, nil, nil, nil, nil, err }
		} else {
			err = halo.RockstarConvertTopN(
				e.HaloCatalog(snap), binFile, n, vars, &hd.Cosmo,
			)
			if err != nil { return nil, nil, nil, nil, nil, nil, err }
		}
	}
	
	rids, xs, ys, zs, ms, rs, err := halo.ReadBinaryRockstar(binFile)
	if err != nil { return nil, nil, nil, nil, nil, nil, err }
	rvals := [][]float64{ xs, ys, zs, ms, rs }

	// Select out only the IDs we want.
	if ids == nil { return rids, xs, ys, zs, ms, rs, nil }
	vals := make([][]float64, len(rvals))

	for i := range vals { vals[i] = make([]float64, len(ids)) }
	f := NewIntFinder(rids)
	for i, id := range ids {
		line, ok := f.Find(id)
		err = fmt.Errorf("Could not find ID %d", id)
		if !ok { return nil, nil, nil, nil, nil, nil, err }
		for vi := range vals { vals[vi][i] = rvals[vi][line] }
	}
	xs, ys, zs, ms, rs = vals[0], vals[1], vals[2], vals[3], vals[4]

	return ids, xs, ys, zs, ms, rs, nil
}

// A quick generic wrapper for doing those one-to-one mappings I need to do so
// often. Written like this so the backend can be swapped out easily.
type IntFinder struct {
	m map[int]int
}

// NewIntFinder creates a new IntFinder struct for a given slice of Rockstar
// IDs.
func NewIntFinder(rids []int) IntFinder {
	f := IntFinder{}
	f.m = make(map[int]int)
	for i, rid := range rids { f.m[rid] = i }
	return f
}

// Find returns the index which the given ID corresponds to and true if the
// ID is in the finder. Otherwise, false is returned.
func (f IntFinder) Find(rid int) (int, bool) {
	line, ok := f.m[rid]
	return line, ok
}

func readHeadersFromSheet(
	snap int, e *env.Environment,
) ([]io.GotetraHeader, []string, error) {
	files := make([]string, e.Blocks())
	hds := make([]io.GotetraHeader, e.Blocks())
	for i := range files {
		files[i] = e.ParticleCatalog(snap, i)
		err := io.ReadSheetHeaderAt(files[i], &hds[i])
		if err != nil { return nil, nil, err }
	}
	return hds, files, nil
}

// ReadHeaders returns all the segment headers and segment file names for all
// the segments at a given snapshot.
func ReadHeaders(
	snap int, e *env.Environment,
) ([]io.GotetraHeader, []string, error) {
	if _, err := os.Stat(e.MemoDir); err != nil { return nil, nil, err }
	memoFile := path.Join(e.MemoDir, fmt.Sprintf(headerMemoFile, snap))

	if _, err := os.Stat(memoFile); err != nil {
		// File not written yet.
		hds, files, err := readHeadersFromSheet(snap, e)
		if err != nil { return nil, nil, err }
		
        f, err := os.Create(memoFile)
        if err != nil { return nil, nil, err }
        defer f.Close()

		raws := make([]io.RawGotetraHeader, len(hds))
		for i := range raws { raws[i] = hds[i].RawGotetraHeader
		}
        binary.Write(f, binary.LittleEndian, raws)

		return hds, files, nil
	} else {
		// File exists: read from it instead.

		f, err := os.Open(memoFile)
        if err != nil { return nil, nil, err }
        defer f.Close()

		hds := make([]io.GotetraHeader, e.Blocks())
		raws := make([]io.RawGotetraHeader, e.Blocks())
        binary.Read(f, binary.LittleEndian, raws)
		for i := range hds { raws[i].Postprocess(&hds[i]) }
		files := make([]string, e.Blocks())
		for i := range files { files[i] = e.ParticleCatalog(snap, i) }

		return hds, files, nil
	}
}
