/*package library implements a Library struct, which contains the names of
all the files specified by some config file input.*/
package library

import (
	"fmt"
	"os"
)

type Library struct {
	snapMin int
	names [][]string
}

func (lib *Library) Blocks() int {
	return len(lib.names[0])
}

func (lib *Library) Filename(snap, block int) string {
	return lib.names[snap + lib.snapMin][block]
}

func (lib *Library) InitGotetra(
	format string, snapMin, snapMax int, blockMins, blockMaxes []int,
) error {
	if len(blockMins) != 3 {
		return fmt.Errorf(
			"'BlockMins' had %d elements, but 3 are required for " +
			"gotetra catalogs.", len(blockMins),
		)
	}

	lib.names = make([][]string, snapMax - snapMin)
	lib.snapMin = snapMin

	for snap := snapMin; snap < snapMax; snap++ {
		for x := blockMins[0]; x < blockMaxes[0]; x++ {
			for y := blockMins[1]; y < blockMaxes[1]; y++ {
				for z := blockMins[2]; z < blockMaxes[2]; z++ {

					fname := fmt.Sprintf(format, snap, x, y, z)
					_, err := os.Stat(fname)
					if err != nil { return err }

					lib.names[snap - snapMin] = append(
						lib.names[snap - snapMin], fname,
					)
				}
			}
		}
	}

	return nil
}

func (lib *Library) InitLGadget2(
	format string, snapMin, snapMax int, blockMins, blockMaxes []int,
) error {
	panic("NYI")
}