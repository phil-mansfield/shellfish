/*package memo is the worst code I have ever written in my life.*/
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
	rockstarMemoDir       = "rockstar"
	rockstarMemoFile      = "halo_%d.dat"
	rockstarShortMemoFile = "halo_short_%d.dat"
	rockstarShortMemoNum  = 10 * 1000

	headerMemoFile = "hd_snap%d.dat"
)

// ReadSortedRockstarIDs returns a slice of IDs corresponding to the highest
// values of some quantity in a particular snapshot. maxID is the number of
// halos to return.
func ReadSortedRockstarIDs(
	snap, maxID int, valName string, vars *halo.VarColumns,
	buf io.VectorBuffer, e *env.Environment,
) ([]int, error) {
	hds, _, err := ReadHeaders(snap, buf, e)
	if err != nil { return nil, err }
	cosmo := &hds[0].Cosmo

	dir := path.Join(e.MemoDir, rockstarMemoDir)
	if _, err := os.Stat(dir); err != nil {
		err = os.Mkdir(dir, 0777)
		if err != nil {
			return nil, err
		}
	}

	var (
		ids []int
		vals [][]float64
		ms  []float64
	)

	if maxID >= rockstarShortMemoNum || maxID == -1 {
		file := path.Join(dir, fmt.Sprintf(rockstarMemoFile, snap))
		ids, vals, err = readRockstar(
			file, []string{valName}, -1, snap, nil, vars, buf, e, cosmo,
		)
		if err != nil {
			return nil, err
		}
		ms = vals[0]
	} else {
		file := path.Join(dir, fmt.Sprintf(rockstarShortMemoFile, snap))
		ids, vals, err = readRockstar(
			file, []string{valName}, rockstarShortMemoNum,
			snap, nil, vars, buf, e, cosmo,
		)
		if err != nil {
			return nil, err
		}
		ms = vals[0]
	}

	if len(ids) < maxID {
		return nil, fmt.Errorf(
			"ID %d too large for snapshot %d", maxID, snap,
		)
	}

	sortRockstar(ids, ms)
	if maxID == -1 {
		return ids, nil
	}
	return ids[:maxID+1], nil
}

type massSet struct {
	ids []int
	ms  []float64
}

func (set massSet) Len() int { return len(set.ids) }

// We're reverse sorting.
func (set massSet) Less(i, j int) bool { return set.ms[i] > set.ms[j] }
func (set massSet) Swap(i, j int) {
	set.ms[i], set.ms[j] = set.ms[j], set.ms[i]
	set.ids[i], set.ids[j] = set.ids[j], set.ids[i]
}

func sortRockstar(ids []int, ms []float64) {
	set := massSet{ids, ms}
	sort.Sort(set)
}

// This function does fairly large heap allocations even when it doesn't need
// to. Consider passing it a buffer.
func ReadRockstar(
	snap int, valNames []string, ids []int, vars *halo.VarColumns,
	buf io.VectorBuffer, e *env.Environment,
) (outIDs []int, vals [][]float64, err error) {
	hds, _, err := ReadHeaders(snap, buf, e)
	if err != nil { return nil, nil, err }
	cosmo := &hds[0].Cosmo

	// Find binFile.
	dir := path.Join(e.MemoDir, rockstarMemoDir)
	if _, err := os.Stat(dir); err != nil {
		err = os.Mkdir(dir, 0777)
		if err != nil {
			return nil, nil, err
		}
	}

	binFile := path.Join(dir, fmt.Sprintf(rockstarMemoFile, snap))
	shortBinFile := path.Join(dir, fmt.Sprintf(rockstarShortMemoFile, snap))

	// This wastes a read the first time it's called. You need to decide if you
	// care. (Answer: probably not.)
	outIDs, vals, err = readRockstar(
		shortBinFile, valNames, rockstarShortMemoNum, snap,
		ids, vars, buf, e, cosmo,
	)
	// TODO: Fix error handling here.
	// (Confession: I have no idea what this comment means: I wrote it half a
	// year ago and didn't make an Issue about it.)
	if err == nil {
		return outIDs, vals, err
	}
	outIDs, vals, err = readRockstar(
		binFile, valNames, -1, snap, ids, vars, buf, e, cosmo,
	)

	if err != nil { return nil, nil, err }
	return outIDs, vals, nil
}

func readRockstar(
	binFile string, valNames []string, n, snap int, ids []int,
	vars *halo.VarColumns, buf io.VectorBuffer, e *env.Environment,
	cosmo *io.CosmologyHeader,
) (outIDs []int, vals [][]float64, err error) {
	hds, _, err := ReadHeaders(snap, buf, e)
	if err != nil {
		return nil, nil, err
	}
	hd := &hds[0]

	// If binFile doesn't exist, create it.
	if _, err := os.Stat(binFile); err != nil {
		if n == -1 {
			err = halo.RockstarConvert(
				e.HaloCatalog(snap), binFile, vars, &hd.Cosmo,
			)
			if err != nil {
				return nil, nil, err
			}
		} else {
			err = halo.RockstarConvertTopN(
				e.HaloCatalog(snap), binFile, n, vars, &hd.Cosmo,
			)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	rids, rawCols, err := halo.ReadBinaryRockstar(binFile, vars)
	if err != nil { return nil, nil, err }

	rvals := make([][]float64, len(valNames))
	for i := range valNames {
		rvals[i] = vars.GetColumn(rawCols, valNames[i], cosmo)
	}

	// Select out only the IDs we want.
	if ids == nil {
		return rids, rvals, nil
	}
	vals = make([][]float64, len(rvals))

	for i := range vals {
		vals[i] = make([]float64, len(ids))
	}

	f := NewIntFinder(rids)
	for i, id := range ids {
		line, ok := f.Find(id)
		err = fmt.Errorf("Could not find ID %d", id)
		if !ok {
			return nil, nil, err
		}
		for vi := range vals {
			vals[vi][i] = rvals[vi][line]
		}
	}

	return ids, vals, nil
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
	for i, rid := range rids {
		f.m[rid] = i
	}
	return f
}

// Find returns the index which the given ID corresponds to and true if the
// ID is in the finder. Otherwise, false is returned.
func (f IntFinder) Find(rid int) (int, bool) {
	line, ok := f.m[rid]
	return line, ok
}

func readUnmemoizedHeaders(
	snap int, buf io.VectorBuffer, e *env.Environment,
) ([]io.Header, []string, error) {
	files := make([]string, e.Blocks())
	hds := make([]io.Header, e.Blocks())

	for i := range files {
		files[i] = e.ParticleCatalog(snap, i)
		err := buf.ReadHeader(files[i], &hds[i])
		if err != nil {
			return nil, nil, err
		}
	}
	return hds, files, nil
}

// ReadHeaders returns all the segment headers and segment file names for all
// the segments at a given snapshot.
func ReadHeaders(
	snap int, buf io.VectorBuffer, e *env.Environment,
) ([]io.Header, []string, error) {
	if _, err := os.Stat(e.MemoDir); err != nil {
		return nil, nil, err
	}
	memoFile := path.Join(e.MemoDir, fmt.Sprintf(headerMemoFile, snap))

	if _, err := os.Stat(memoFile); err != nil {
		// File not written yet.
		hds, files, err := readUnmemoizedHeaders(snap, buf, e)
		if err != nil {
			return nil, nil, err
		}

		f, err := os.Create(memoFile)
		if err != nil {
			return nil, nil, err
		}
		defer f.Close()

		binary.Write(f, binary.LittleEndian, hds)

		return hds, files, nil
	} else {
		// File exists: read from it instead.

		f, err := os.Open(memoFile)
		if err != nil {
			return nil, nil, err
		}
		defer f.Close()

		hds := make([]io.Header, e.Blocks())
		binary.Read(f, binary.LittleEndian, hds)
		files := make([]string, e.Blocks())
		for i := range files {
			files[i] = e.ParticleCatalog(snap, i)
		}

		return hds, files, nil
	}
}
