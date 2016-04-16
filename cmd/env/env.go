package env

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

type Environment struct {
	Catalogs
	Halos
	MemoDir string

}

//////////////
// Catalogs //
//////////////

type Catalogs struct {
	snapMin int
	names [][]string
}


func (cat *Catalogs) Blocks() int {
	return len(cat.names[0])
}

func (cat *Catalogs) ParticleCatalog(snap, block int) string {
	return cat.names[snap - cat.snapMin][block]
}

func (cat *Catalogs) InitGotetra(
	format string, snapMin, snapMax int64, blockMins, blockMaxes []int64,
	validate bool,
) error {
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

func (cat *Catalogs) InitLGadget2(
	format string, snapMin, snapMax int64, blockMins, blockMaxes []int,
) error {
	panic("NYI")
}

///////////
// Halos //
///////////

type Halos struct {
	snapMin int
	snapOffset int
	names []string
}

func (h *Halos) HaloCatalog(snap int) string {
	return h.names[snap - h.snapMin]
}

func (h *Halos) SnapOffset() int {
	return h.snapOffset
}

func (h *Halos) InitRockstar(dir string, snapMin, snapMax int64) error {
	infos, err := ioutil.ReadDir(dir)
	if err != nil { return err }

	h.snapOffset = int(snapMax) - len(infos)

	h.snapMin = int(snapMin)
	h.names = []string{}
	for i := range infos {
		h.names = append(h.names, path.Join(dir, infos[i].Name()))
	}

	if len(h.names) < int(snapMax - snapMin) + 1 {
		return fmt.Errorf(
			"There are %d files in the 'HaloDir' directory, %s, but " +
			"'SnapMin' = %d and 'SnapMax' = %d.",
			len(h.names), dir, snapMin, snapMax,
		)
	}
	h.names = h.names[len(h.names) - int(snapMax - snapMin + 1):]

	return nil
}
