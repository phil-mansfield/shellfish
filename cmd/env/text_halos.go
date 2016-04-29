package env

import (
	"fmt"
	"io/ioutil"
	"path"
)

func (h *Halos) InitTextHalo(dir string, snapMin, snapMax int64) error {
	h.HaloType = Rockstar
	h.TreeType = ConsistentTrees

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
