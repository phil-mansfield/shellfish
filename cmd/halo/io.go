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

// This type is a look into the heart of darkness.
type VarColumns struct {
	ColumnLookup map[string]int
	Names []string
	Columns []int
	Generator []string
	NBinary int
	RadiusUnits string
}

func NewVarColumns(
	names []string, columns []int64, radiusUnits string,
) *VarColumns {
	vc :=&VarColumns{}
	vc.ColumnLookup = make(map[string]int)

	for i := range names {
		vc.ColumnLookup[names[i]] = i
		vc.Names = append(vc.Names, names[i])
		vc.Columns = append(vc.Columns, int(columns[i]))
	}
	vc.Generator = make([]string, len(vc.Names))
	vc.NBinary = len(vc.Names)

	mNames := []string{"M200m", "M200c", "M500c", "M2500c"}
	rNames := []string{"R200m", "R200c", "R500c", "R2500c"}
	for i, _ := range mNames {
		_, rOk := vc.ColumnLookup[rNames[i]]
		_, mOk := vc.ColumnLookup[mNames[i]]
		if mOk && !rOk {
			vc.ColumnLookup[rNames[i]] = len(vc.Names)
			vc.Names = append(vc.Names, rNames[i])
			vc.Columns = append(vc.Columns, -1)
			vc.Generator = append(vc.Generator, mNames[i])
		}
	}
	vc.RadiusUnits = radiusUnits

	return vc
}

func (vc *VarColumns) GetColumn(
	cols [][]float64, name string, cosmo *io.CosmologyHeader,
) []float64 {
	col, _ := vc.ColumnLookup[name]
	if vc.Generator[col] == "" { return cols[col] }
	
	gCol, _ := vc.ColumnLookup[vc.Generator[col]]
	ms := cols[gCol]
	rs := make([]float64, len(ms))

	rad, _ := RadiusFromString(name)
	
	rad.Radius(cosmo, ms, rs)
	// Convert from Mpc/h to the units of the halo catalog.
	ucf := UnitConversionFactor(vc.RadiusUnits, cosmo)
	for i := range rs {
		rs[i] /= ucf
	}
	
	return rs
}

func (vc *VarColumns) GetIDs(cols [][]float64) []int {
	col, _ := vc.ColumnLookup["ID"]

	out := make([]int, len(cols[col]))
	for i := range out { out[i] = int(cols[col][i]) }
	return out
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
	valIdxs := vars.Columns
	for i := range valIdxs {
		if valIdxs[i] == -1 {
			valIdxs = valIdxs[:i]
			break
		}
	}

	cols, err := readTable(inFile, valIdxs)
	if err != nil {
		return err
	}

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
	valIdxs := vars.Columns
	for i := range valIdxs {
		if valIdxs[i] == -1 {
			valIdxs = valIdxs[:i]
			break
		}
	}

	cols, err := readTable(inFile, valIdxs)
	if err != nil {
		return err
	}

	if n > len(cols[0]) {
		n = len(cols[0])
	}

	idxs := idxSort(vars.GetColumn(cols, "M200m", cosmo))[len(cols[0])-n:]

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
	file string, vc *VarColumns,
) (ids []int, rawCols [][]float64, err error) {
	return readRockstarVals(file, binaryColGetter, vc)
}

func readRockstarVals(
	file string, getter colGetter, vc *VarColumns,
) (ids []int, rawCols [][]float64, err error) {
	
	colIdxs := make([]int, vc.NBinary)
	for i := range colIdxs { colIdxs[i] = i }

	rawCols, err = getter(file, colIdxs)
	if err != nil { return nil, nil, err }

	ids = vc.GetIDs(rawCols)
	return ids, rawCols, err
}

type colGetter func(file string, colIdxs []int) ([][]float64, error)

func binaryColGetter(file string, colIdxs []int) ([][]float64, error) {
	// TODO: make this not slow as molasses. Just read everything in one go.
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
