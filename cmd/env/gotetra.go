package env

import (
	"fmt"
	"os"
)

func (cat *Catalogs) InitGotetra(
format string, snapMin, snapMax int64, blockMins, blockMaxes []int64,
validate bool,
) error {
	cat.CatalogType = Gotetra

	if len(blockMins) != 3 {
		return fmt.Errorf(
			"'BlockMins' had %d elements, but 3 are required for " +
			"gotetra catalogs.", len(blockMins),
		)
	}

	cat.names = make([][]string, snapMax - snapMin + 1)
	cat.snapMin = int(snapMin)

	for snap := snapMin; snap <= snapMax; snap++ {
		for x := blockMins[0]; x <= blockMaxes[0]; x++ {
			for y := blockMins[1]; y <= blockMaxes[1]; y++ {
				for z := blockMins[2]; z <= blockMaxes[2]; z++ {

					fname := fmt.Sprintf(format, snap, x, y, z)
					if validate {
						_, err := os.Stat(fname)
						if err != nil { return err }
					}

					cat.names[snap - snapMin] = append(
						cat.names[snap - snapMin], fname,
					)
				}
			}
		}
	}
	return nil
}