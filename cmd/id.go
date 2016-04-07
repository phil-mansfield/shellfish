package cmd

import (
	"fmt"
	"github.com/phil-mansfield/shellfish/parse"
)

// IDConfig contains the configuration fileds for the 'id' mode of the shellfish
// tool.
type IDConfig struct {
	idType string
	ids []int64
	idStart, idEnd, snap, mult int64

	exclusionStrategy string
	exclusionRadiusMult float64
}

var _ Mode = &IDConfig{}

// ExampleConfig creates an example id.config file.
func (config *IDConfig) ExampleConfig() string {
	return `[id.config]
#####################
## Required Fields ##
#####################

# Index of the snapshot to be analyzed.
Snap = 100

IDs = 10, 11, 12, 13, 14

#####################
## Optional Fields ##
#####################

# IDType indicates what the input IDs correspond to. It can be set to the
# following modes:
# halo-id - The numeric IDs given in the halo catalog.
# m200m   - The rank of the halos when sorted by M200m.
#
# Defaults to m200m if not set.
# IDType = m200m

# An alternative way fo specifying IDs is to select a start and end (inclusive)
# ID value. If IDs is not set, both of these values must be set.
#
# IDStart = 10
# IDEnd = 15

# ExclusionStrategy determines how to exclude IDs from the given set. This is
# useful because splashback shells are not particularly meaningful for
# subhalos. It can be set to the following modes:
# none     - No halos are removed
# subhalos - Halos flagged as subhalos in the catalog are removed
# r200m    - Halos which have an R200m shell that overlaps with a larger halo's
#            R200m shell are removed
#
# ExclusionStrategy defaults to subhalo if not set.
#
# ExclusionStrategy = subhalo

# ExclusionRadiusMult is a multiplier of R200m applied for the sake of
# determining exclusions.
#
# ExclusionRadiusMult defaults to 1 if not set.
#
# ExclustionRadiusMult = 1

# Mult is the number of times a given ID should be repeated. This is most useful
# if you want to estimate the scatter in shell measurements for halos with a
# given set of shell parameters.
#
# Mult defaults to 1 if not set.
#
# Mult = 1`
}

// ReadConfig reads in an id.config file into config.
func (config *IDConfig) ReadConfig(fname string) error {

	vars := parse.NewConfigVars("id.config")
	vars.String(&config.idType, "IDType", "m200m")
	vars.Ints(&config.ids, "IDs", []int64{})
	vars.Int(&config.idStart, "IDStart", -1)
	vars.Int(&config.idEnd, "IDEnd", -1)
	vars.Int(&config.mult, "Mult", 1)
	vars.Int(&config.snap, "Snap", -1)
	vars.String(&config.exclusionStrategy, "ExclusionStrategy", "subhalo")
	vars.Float(&config.exclusionRadiusMult, "ExclusionRadiusMult", 1)

	if err := parse.ReadConfig(fname, vars); err != nil { return err }
	return config.validate()
}

// validate checks whether all the fields of config are valid.
func (config *IDConfig) validate() error {
	switch config.idType {
	case "halo-id", "m200m":
	default:
		return fmt.Errorf("The 'IDType' variable is set to '%s', which I " +
			"don't recognize.", config.idType)
	}

	switch config.exclusionStrategy {
	case "none", "subhalo":
	case "r200m":
		if config.exclusionRadiusMult <= 0 {
			return fmt.Errorf("The 'ExclusionRadiusMult' varaible is set to " +
				"%g, but it needs to be positive.", config.exclusionRadiusMult)
		}
	default:
		return fmt.Errorf("The 'ExclusionStrategy' variable is set to '%s', " +
		"which I don't recognize.", config.exclusionStrategy)
	}

	// TODO: Check the ranges of the IDs as well as IDStart and IDEnd
	if len(config.ids) == 0 {
		switch {
		case config.idStart == -1 && config.idEnd == -1:
			return fmt.Errorf("'IDs' variable not set.")
		case config.idStart == -1:
			return fmt.Errorf("'IDStart variable not set.")
		case config.idEnd == -1:
			return fmt.Errorf("'IDEnd' variable not set.")
		case config.idEnd < config.idStart:
			return fmt.Errorf("'IDEnd' variable set to %d, but 'IDStart' " +
				"variable set to %d.", config.idEnd, config.idStart)
		}
	}

	switch {
	case config.snap == -1:
		return fmt.Errorf("'Snap' variable not set.")
	case config.snap < 0:
		return fmt.Errorf("'Snap' variable set to %d.", config.snap)
	}

	if config.mult <= 0 {
		return fmt.Errorf("'Mult' variable set to %d", config.mult)
	}

	return nil
}

// Run executes the ID mode of shellfish tool.
func (config *IDConfig) Run(
	flags []string, gConfig *GlobalConfig, stdin []string,
) ([]string, error) {
	panic("NYI")
}