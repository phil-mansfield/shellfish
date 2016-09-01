package path

import (
	"log"
	"time"

	"github.com/phil-mansfield/shellfish/cmd/env"
	"github.com/phil-mansfield/shellfish/logging"
)

type ProfConfig struct {
	selectSnaps []int64
}

var _ Mode = &ProfConfig{}

func (config *ProfConfig) ExampleConfig() string {
	return `[tree.config]

#####################
## Optional Fields ##
#####################

# SelectSnaps is a list of all the snapshots which halo IDs should be
# output at. If not set, IDs will be output at all snapshots.
#
# SelectSnaps = 36, 47, 64, 77, 87, 100`
}


func (config *ProfConfig) ReadConfig(fname string) error {
	panic("NYI")
}

func (config *ProfConfig) validate() error { return nil }

func (config *ProfConfig) Run(
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

	if logging.Mode == logging.Performance {
		log.Printf("Time: %s", time.Since(t).String())
		log.Printf("Memory:\n%s", logging.MemString())
	}
}