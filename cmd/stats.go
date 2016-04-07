package cmd

import (
	"fmt"

	"github.com/phil-mansfield/shellfish/parse"
)

type StatsConfig struct {
	values []string
	histogramBins int64
	monteCarloSamples int64
	exclusionStrategy string
}

var _ Mode = &StatsConfig{}

func (config *StatsConfig) ExampleConfig() string {
	return `[stats.config]

#####################
## Required Fields ##
#####################

# Values determines what columns will be written to stdout. If one of the
# elements of the list corresponds to a histogram, then HistogramBins x values
# will be written starting at that column, HistogramBins y values will be
# written after that, and the specified columns will continue from there.
#
# The supported columns are:
# id       - The ID of the halo, as initially supplied.
# snap     - The snapshot index of the halo, as initially supplied.
# r-sp     - The volume-weighted splashback radius of the halo.
# m-sp     - The total mass contained within the splashback shell of the halo.
# r-sp-max - The maximum radius of the splashback shell.
# r-sp-min - The minimum radius of the splashback shell.
Values = snap, id, r-sp

#####################
## Optional Fields ##
#####################

# HistogramBins is the number of bins to use for histogramed quantities.
HistogramBins = 50

# MonteCarloSamplings The number of Monte Carlo samplings done when calculating
# properties of shells.
MonteCarloSamples = 10000

# Strategy for removing halos contained within a larger halo's splashback
# shell.
#
# The supported strategies are:
# none    - Don't try to do this.
# contain - Only halos which have a center inside a larger halo's splashback are
#           excluded.
# overlap - Halos which have a splashback shell that overlaps the splashback
#           shell of a larger halo are excluded.
#
# The default value is none.
ExclusionStrategy = none
`
}

func (config *StatsConfig) ReadConfig(fname string) error {
	vars := parse.NewConfigVars("stats.config")

	vars.Strings(&config.values, "Values", []string{})
	vars.Int(&config.histogramBins, "HistogramBins", 50)
	vars.Int(&config.monteCarloSamples, "MonteCarloSamples", 10 * 1000)
	vars.String(&config.exclusionStrategy, "ExclusionStrategy", "none")

	if err := parse.ReadConfig(fname, vars); err != nil { return err }
	return config.validate()
}

func (config *StatsConfig) validate() error {
	for i, val := range config.values {
		switch val {
		case "snap", "id", "m-sp", "r-sp", "r-sp-min", "r-sp-max":
		default:
			return fmt.Errorf("Item %d of variable 'Values' is set to '%s', " +
				"which I don't recognize.", i, val)
		}
	}

	switch config.exclusionStrategy {
	case "none", "contain", "overlap":
	default:
		return fmt.Errorf("variable 'ExclusionStrategy' set to '%s', which " +
			"I don't recognize.", config.exclusionStrategy)
	}

	switch {
	case config.histogramBins <= 0:
		return fmt.Errorf("The variable '%s' was set to %g",
			"HistogramBins", config.histogramBins)
	case config.monteCarloSamples <= 0:
		return fmt.Errorf("The variable '%s' was set to %g",
			"MonteCarloSamples", config.monteCarloSamples)
	}

	return nil
}

func (config *StatsConfig) Run(
	flags []string, gConfig *GlobalConfig, stdin []string,
) ([]string, error) {
	panic("NYI")
}