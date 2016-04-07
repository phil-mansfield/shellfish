package cmd

import (
	"fmt"

	"github.com/phil-mansfield/shellfish/cmd/util"
)

// TODO: Generalize to other file formats

func sheetNum(config *GlobalConfig) int {
	prod := 1
	for i := range config.formatMins {
		prod *= 1 + int(config.formatMaxes[i] - config.formatMaxes[i])
	}
	return prod
}

func snapNum(config *GlobalConfig) int {
	return int(config.snapMax - config.snapMin) + 1
}

func snapOffset(config *GlobalConfig) (int, error) {
	snapNum := snapNum(config)
	hlists, err := util.DirContents(config.haloDir)
	if err != nil { return -1, err }
	if snapNum - len(hlists) < 0 {
		return -1, fmt.Errorf("Fewer particle snapshots than halo lists.")
	}
	return snapNum - len(hlists), nil
}
