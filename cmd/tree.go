package cmd

import (
	"io/ioutil"
	"log"
	"path"
	"time"

	"github.com/phil-mansfield/shellfish/cmd/catalog"
	"github.com/phil-mansfield/shellfish/cmd/env"
	"github.com/phil-mansfield/shellfish/logging"
	"github.com/phil-mansfield/shellfish/los/tree"
	"github.com/phil-mansfield/shellfish/parse"
)

type TreeConfig struct {
	selectSnaps []int64
}

var _ Mode = &TreeConfig{}

func (config *TreeConfig) ExampleConfig() string {
	return `[tree.config]

#####################
## Optional Fields ##
#####################

# SelectSnaps is a list of all the snapshots which halo IDs should be
# output at. If not set, IDs will be output at all snapshots.
#
# SelectSnaps = 36, 47, 64, 77, 87, 100`
}

func (config *TreeConfig) ReadConfig(fname string) error {
	if fname == "" {
		return nil
	}

	vars := parse.NewConfigVars("tree.config")
	vars.Ints(&config.selectSnaps, "SelectSnaps", []int64{})

	if err := parse.ReadConfig(fname, vars); err != nil {
		return err
	}

	return nil
}

func (config *TreeConfig) validate() error { return nil }

func (config *TreeConfig) Run(
	flags []string, gConfig *GlobalConfig, e *env.Environment, stdin []string,
) ([]string, error) {
	if logging.Mode != logging.Nil {
		log.Println(`
####################
## shellfish tree ##
####################`,
		)
	}
	var t time.Time
	if logging.Mode == logging.Performance {
		t = time.Now()
	}

	intCols, _, err := catalog.ParseCols(stdin, []int{0, 1}, []int{})
	if err != nil {
		return nil, err
	}
	inputIDs := intCols[0]

	trees, err := treeFiles(gConfig)
	if err != nil {
		return nil, err
	}

	idSets, snapSets, err := tree.HaloHistories(
		trees, inputIDs, e.SnapOffset(),
	)
	if err != nil {
		return nil, err
	}

	ids, snaps := []int{}, []int{}
	for i := range idSets {
		ids = append(ids, idSets[i]...)
		snaps = append(snaps, snapSets[i]...)
		// Sentinels:
		if i != len(idSets)-1 {
			ids = append(ids, -1)
			snaps = append(snaps, -1)
		}
	}

	lines := catalog.FormatCols(
		[][]int{ids, snaps}, [][]float64{}, []int{0, 1},
	)
	fLines := []string{}
	for i := range lines {
		if snaps[i] >= int(gConfig.SnapMin) &&
			snaps[i] <= int(gConfig.SnapMax) {

			if len(config.selectSnaps) > 0 {
				for j := range config.selectSnaps {
					if int(config.selectSnaps[j]) == snaps[i] {
						fLines = append(fLines, lines[i])
					}
				}
			} else {
				fLines = append(fLines, lines[i])
			}
		}
	}

	cString := catalog.CommentString(
		[]string{"ID", "Snapshot"}, []string{}, []int{0, 1}, []int{1, 1},
	)

	if logging.Mode == logging.Performance {
		log.Printf("Time: %s", time.Since(t).String())
		log.Printf("Memory:\n%s", logging.MemString())
	}

	return append([]string{cString}, fLines...), nil
}

func treeFiles(gConfig *GlobalConfig) ([]string, error) {
	infos, err := ioutil.ReadDir(gConfig.TreeDir)
	if err != nil {
		return nil, err
	}

	names := []string{}
	for _, info := range infos {
		name := info.Name()
		n := len(name)
		// This is pretty hacky.
		if n > 4 && name[:5] == "tree_" && name[n-4:] == ".dat" {
			names = append(names, path.Join(gConfig.TreeDir, name))
		}
	}
	return names, nil
}
