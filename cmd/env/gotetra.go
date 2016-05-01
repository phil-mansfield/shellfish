package env

import (
	"fmt"
)

func (cat *Catalogs) InitGotetra(info *ParticleInfo, validate bool) error {

	if len(info.BlockMins) != 3 {
		return fmt.Errorf(
			"'BlockMins' had %d elements, but 3 are required for " +
			"gotetra catalogs.", len(info.BlockMins),
		)
	}

	cat.CatalogType = Gotetra
	cat.snapMin = int(info.SnapMin)

	cols := make([][]interface{}, len(info.SnapshotFormat))
	snapAligned := make([]bool, len(info.SnapshotFormat))
	for i := range cols {
		var err error
		cols[i], snapAligned[i], err = info.GetColumn(i)
		if err != nil { return err }
	}

	formatArgs := interleave(cols, snapAligned)
	cat.names = [][]string{}
	for snap := range formatArgs {
		names := []string{}
		for block := range formatArgs[snap] {
			names = append(names,
				fmt.Sprintf(info.SnapshotFormat, formatArgs[snap][block]...),
			)
		}
		cat.names = append(cat.names, names)
	}

	if validate { panic("File validation not yet implemented.") }

	return nil
}