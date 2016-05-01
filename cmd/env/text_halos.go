package env

import (
	"fmt"
	"io/ioutil"
	"path"
)

func (h *Halos) InitTextHalo(info *HaloInfo) error {
	h.HaloType = Rockstar
	h.TreeType = ConsistentTrees

	infos, err := ioutil.ReadDir(info.HaloDir)
	if err != nil { return err }

	h.snapOffset = int(info.HSnapMax) - len(infos)

	h.snapMin = int(info.HSnapMin)
	h.names = []string{}
	for i := range infos {
		h.names = append(h.names, path.Join(info.HaloDir, infos[i].Name()))
	}

	if len(h.names) < int(info.HSnapMax - info.HSnapMin) + 1 {
		return fmt.Errorf(
			"There are %d files in the 'HaloDir' directory, %s, but " +
			"'SnapMin' = %d and 'SnapMax' = %d.",
			len(h.names), info.HaloDir, info.HSnapMin, info.HSnapMax,
		)
	}
	h.names = h.names[len(h.names) - int(info.HSnapMax - info.HSnapMin + 1):]

	return nil
}
