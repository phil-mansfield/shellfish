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
	snap, maxID int, e *env.Environment, flag halo.Val,
) ([]int, error) {
	dir := path.Join(e.MemoDir, rockstarMemoDir)
	if err, _ := os.Stat(dir); err != nil { os.Mkdir(dir, 0777) }

	var (
		vals [][]float64
		err error
	)
	if maxID >= rockstarShortMemoNum || maxID == -1 {
		file := path.Join(dir, fmt.Sprintf(rockstarMemoFile, snap))
		vals, err = readRockstar(
			file, -1, snap, nil, e, halo.ID, flag,
		)
		if err != nil { return nil, err }
	} else {
		file := path.Join(dir, fmt.Sprintf(rockstarShortMemoFile, snap))
		vals, err = readRockstar(
			file, rockstarShortMemoNum, snap, nil, e, halo.ID, flag,
		)
		if err != nil { return nil, err }
	}
	
	fids, ms := vals[0], vals[1]
	ids := make([]int, len(fids))
	for i := range ids { ids[i] = int(fids[i]) }

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
	snap int, ids []int, e *env.Environment, valFlags ...halo.Val,
) ([][]float64, error) {
	// Find binFile.
	dir := path.Join(e.MemoDir, rockstarMemoDir)
	if err, _ := os.Stat(dir); err != nil { os.Mkdir(dir, 0777) }

	binFile := path.Join(dir, fmt.Sprintf(rockstarMemoFile, snap))
	shortBinFile := path.Join(dir, fmt.Sprintf(rockstarShortMemoFile, snap))

	// This wastes a read the first time it's called. You need to decide if you
	// care. (Answer: probably.)
	vals, err := readRockstar(
		shortBinFile, rockstarShortMemoNum, snap, ids, e, valFlags...,
	)
	if err == nil { return vals, err }
	return readRockstar(binFile, -1, snap, ids, e, valFlags...)
}

func readRockstar(
	binFile string, n, snap int, ids []int,
	e *env.Environment, valFlags ...halo.Val,
) ([][]float64, error) {
	// If binFile doesn't exist, create it.
	if _, err := os.Stat(binFile); err != nil {
		if n == -1 {
			err = halo.RockstarConvert(e.HaloCatalog(snap), binFile)
			if err != nil { return nil, err }
		} else {
			err = halo.RockstarConvertTopN(e.HaloCatalog(snap), binFile, n)
			if err != nil { return nil, err}
		}
	}

	hds, _, err := ReadHeaders(snap, e)
	if err != nil { return nil, err }
	hd := &hds[0]

	
	rids, rvals, err := halo.ReadBinaryRockstarVals(
		binFile, &hd.Cosmo, valFlags...,
	)	
	if err != nil { return nil, err }
	
	// Select out only the IDs we want.
	if ids == nil { return rvals, nil }
	vals := make([][]float64, len(rvals))

	for i := range vals { vals[i] = make([]float64, len(ids)) }
	f := NewIntFinder(rids)
	for i, id := range ids {
		line, ok := f.Find(id)
		if !ok { return nil, fmt.Errorf("Could not find ID %d", id) }
		for vi := range vals { vals[vi][i] = rvals[vi][line] }
	}
	
	return vals, nil
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
