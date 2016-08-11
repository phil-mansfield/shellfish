package halo

import (
	"encoding/binary"
	go_io "io"
	"os"
	"sort"
	"strings"

	"github.com/phil-mansfield/shellfish/cmd/catalog"
	"github.com/phil-mansfield/shellfish/io"
)

// halos allows for arrays of halo properties to be sorted simultaneously.
type halos struct {
	rids               []int
	xs, ys, zs, ms, rs []float64
}

type VarColumns struct {
	ID, X, Y, Z, M200m int
}

func (hs *halos) Len() int           { return len(hs.rs) }
func (hs *halos) Less(i, j int) bool { return hs.rs[i] < hs.rs[j] }
func (hs *halos) Swap(i, j int) {
	hs.rs[i], hs.rs[j] = hs.rs[j], hs.rs[i]
	hs.ms[i], hs.ms[j] = hs.ms[j], hs.ms[i]
	hs.xs[i], hs.xs[j] = hs.xs[j], hs.xs[i]
	hs.ys[i], hs.ys[j] = hs.ys[j], hs.ys[i]
	hs.zs[i], hs.zs[j] = hs.zs[j], hs.zs[i]
	hs.rids[i], hs.rids[j] = hs.rids[j], hs.rids[i]
}

func RockstarConvert(
	inFile, outFile string, vars *VarColumns, cosmo *io.CosmologyHeader,
) error {
	valIdxs := []int{vars.ID, vars.X, vars.Y, vars.Z, vars.M200m}

	cols, err := readTable(inFile, valIdxs)
	if err != nil {
		return err
	}

	cols = genRadiiMasses(cols, vars, cosmo)

	f, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer f.Close()

	err = binary.Write(f, binary.LittleEndian, int64(len(cols[0])))
	if err != nil {
		return err
	}
	for _, col := range cols {
		err := binary.Write(f, binary.LittleEndian, col)
		if err != nil {
			return err
		}
	}

	return nil
}

func genRadiiMasses(
	cols [][]float64, vars *VarColumns, cosmo *io.CosmologyHeader,
) [][]float64 {
	ms := cols[4]
	rs := make([]float64, len(ms))
	R200m.Radius(cosmo, ms, rs)

	cols = append(cols, [][]float64{rs, ms}...)

	return cols
}

type idxSet struct {
	xs   []float64
	idxs []int
}

func (set idxSet) Less(i, j int) bool { return set.xs[i] < set.xs[j] }
func (set idxSet) Len() int           { return len(set.xs) }
func (set idxSet) Swap(i, j int) {
	set.xs[i], set.xs[j] = set.xs[j], set.xs[i]
	set.idxs[i], set.idxs[j] = set.idxs[j], set.idxs[i]
}

func idxSort(xs []float64) []int {
	xsCopy := make([]float64, len(xs))
	copy(xsCopy, xs)
	idxs := make([]int, len(xs))
	for i := range idxs {
		idxs[i] = i
	}

	set := idxSet{}
	set.idxs = idxs
	set.xs = xsCopy
	sort.Sort(set)
	return idxs
}

func RockstarConvertTopN(
	inFile, outFile string, n int, vars *VarColumns, cosmo *io.CosmologyHeader,
) error {
	valIdxs := []int{vars.ID, vars.X, vars.Y, vars.Z, vars.M200m}

	cols, err := readTable(inFile, valIdxs)
	if err != nil {
		return err
	}

	cols = genRadiiMasses(cols, vars, cosmo)

	if n > len(cols[0]) {
		n = len(cols[0])
	}
	idxs := idxSort(cols[4])[len(cols[0])-n:]

	outCols := make([][]float64, len(cols))
	for i := range cols {
		outCols[i] = make([]float64, len(idxs))
	}

	for j := range cols {
		for i, idx := range idxs {
			outCols[j][i] = cols[j][idx]
		}
	}

	f, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer f.Close()

	err = binary.Write(f, binary.LittleEndian, int64(n))
	if err != nil {
		return err
	}
	for _, col := range outCols {
		err := binary.Write(f, binary.LittleEndian, col)
		if err != nil {
			return err
		}
	}

	return nil
}

func ReadBinaryRockstar(
	file string,
) (ids []int, xs, ys, zs, ms, rs []float64, err error) {
	return readRockstarVals(file, binaryColGetter)
}

func readRockstarVals(
	file string, getter colGetter,
) (ids []int, xs, ys, zs, ms, rs []float64, err error) {
	colIdxs := []int{0, 1, 2, 3, 4, 5}
	vals, err := getter(file, colIdxs)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	xs, ys, zs, ms, rs = vals[1], vals[2], vals[3], vals[4], vals[5]

	ids = make([]int, len(vals[0]))
	for i := range vals[0] {
		ids[i] = int(vals[0][i])
	}

	return ids, xs, ys, zs, ms, rs, nil
}

type colGetter func(file string, colIdxs []int) ([][]float64, error)

func binaryColGetter(file string, colIdxs []int) ([][]float64, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	n := int64(0)
	err = binary.Read(f, binary.LittleEndian, &n)
	if err != nil {
		return nil, err
	}

	jump := n * 8
	cols := make([][]float64, len(colIdxs))
	for i := range cols {
		cols[i] = make([]float64, n)
	}
	for i, colIdx := range colIdxs {
		_, err = f.Seek(8+jump*int64(colIdx), 0)
		if err != nil {
			return nil, err
		}
		err = binary.Read(f, binary.LittleEndian, cols[i])
		if err != nil {
			return nil, err
		}
	}
	return cols, nil
}

func readTable(file string, colIdxs []int) ([][]float64, error) {
	// TODO: Heavily optimize this.

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	bs := make([]byte, info.Size())
	_, err = go_io.ReadFull(f, bs)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(bs), "\n")

	_, floats, err := catalog.ParseCols(lines, nil, colIdxs)
	return floats, err
}
