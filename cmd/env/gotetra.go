package env

import (
	"fmt"
	"os"
)

func (cat *Catalogs) InitGotetra(info *ParticleInfo, validate bool) error {
	cat.CatalogType = Gotetra

	if len(info.FormatMins) != 3 {
		return fmt.Errorf(
			"'BlockMins' had %d elements, but 3 are required for " +
			"gotetra catalogs.", len(info.FormatMins),
		)
	}

	cat.names = make([][]string, info.SnapMax - info.SnapMin + 1)
	cat.snapMin = int(info.SnapMin)

	for snap := info.SnapMin; snap <= info.SnapMax; snap++ {
		for x := info.FormatMins[0]; x <= info.FormatMaxes[0]; x++ {
			for y := info.FormatMins[1]; y <= info.FormatMaxes[1]; y++ {
				for z := info.FormatMins[2]; z <= info.FormatMaxes[2]; z++ {

					fname := fmt.Sprintf(info.SnapshotFormat, snap, x, y, z)
					if validate {
						_, err := os.Stat(fname)
						if err != nil { return err }
					}

					cat.names[snap - info.SnapMin] = append(
						cat.names[snap - info.SnapMin], fname,
					)
				}
			}
		}
	}
	return nil
}