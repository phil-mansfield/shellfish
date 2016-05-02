package env

import (
	"fmt"
)

func (cat *Catalogs) InitARTIO(info *ParticleInfo, validate bool) error {
	cat.CatalogType = ARTIO
	cat.snapMin = int(info.SnapMin)

	cols := make([][]interface{}, len(info.SnapshotFormatMeanings))
	snapAligned := make([]bool, len(info.SnapshotFormatMeanings))
	for i := range cols {
		var err error
		cols[i], snapAligned[i], err = info.GetColumn(i)
		if err != nil {
			return err
		}
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

	if validate {
		panic("File validation not yet implemented.")
	}

	return nil
}
