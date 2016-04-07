package cmd

import (
	"fmt"

	"github.com/phil-mansfield/shellfish/parse"
)

type ShellConfig struct {
	radialBins, spokes, rings int64
	rMaxMult, rMinMult float64
	rKernelMult float64

	eta float64
	order, smoothingWindow, levels, subsampleFactor int64
	losSlopeCutoff float64
}

var _ Mode = &ShellConfig{}

func (config *ShellConfig) ExampleConfig() string {
	return `[shell.config]

# Note: All fields in this file are optional. The value the field is set to
# is the default value. It is unlikely that the average user will want to
# change any of these values.
#
# If you find that the default parameters are behaving suboptimally in some
# significant way on halos with more than 50,000 particles, please submit a
# report on https://github.com/phil-mansfield/shellfish/issues.

# SubsampleFactor is the factor by which the points should be subsampled. This
# parameter is chiefly useful for convergence testing.
SubsampleFactor = 1

# RadialBins is the number of radial bins used for each line of sight.
RadialBins = 256

# Spokes is the number of lines of sight per ring.
Spokes = 256

# Rings is the number of rungs per halo.
Rings = 24

# RMaxMult is the maximum radius of a line of sight as a multiplier of R200m.
RMaxMult = 3.0

# RMaxMult is the minimum radius of a line of sight as a multiplier of R200m.
RMinMult = 0.5

# KernelRadiusMult is the radius of the spherical kernels around every
# particle as a multiplier of R200m.
RKernelMult = 0.2

# Eta is the tuning paramter of the point-filtering routines and determines the
# characteristic scale used in this step. Exactly what this entails is too
# complicated to be described here and can be found in the Shellfish paper.
Eta = 10.0

# Order indicates the order of the Penna function used to represent the
# splashback shell.
Order = 3

# Levels is the number of recursive angular splittings that should be done when
# filtering points.
Levels = 3

# SmoothingWindow is the width of the Savitzky-Golay smoothing window used
# when finding the point of steepest slope along lines of sight. Must be an odd
# number.
SmoothingWindow = 121

# Cutoff is the minimum slope allowed when finding the point of steepest slope
# for individual lines of sight.
LOSSlopeCutoff = 0.0`
}

func (config *ShellConfig) ReadConfig(fname string) error {
	vars := parse.NewConfigVars("shell.config")

	vars.Int(&config.subsampleFactor, "SubsampleFactor", 1)
	vars.Int(&config.radialBins, "RadialBins", 256)
	vars.Int(&config.spokes, "Spokes", 256)
	vars.Int(&config.rings, "Rings", 24)
	vars.Float(&config.rMaxMult, "RMaxMult", 3)
	vars.Float(&config.rMinMult, "RMinMult", 0.5)
	vars.Float(&config.rKernelMult, "RKernelMult", 0.2)
	vars.Float(&config.eta, "Eta", 10)
	vars.Int(&config.order, "Order", 3)
	vars.Int(&config.levels, "Levels", 3)
	vars.Int(&config.smoothingWindow, "SmoothingWindow", 121)
	vars.Float(&config.losSlopeCutoff, "LOSSlopeCutoff", 0.0)

	if err := parse.ReadConfig(fname, vars); err != nil { return err }
	return config.validate()
}

func (config *ShellConfig) validate() error {
	switch {
	case config.subsampleFactor <= 0:
		return fmt.Errorf("The variable '%s' was set to %d.",
			"SubsampleFactor", config.subsampleFactor)
	case config.radialBins <= 0:
		return fmt.Errorf("The variable '%s' was set to %d.",
			"RadialBins", config.radialBins)
	case config.spokes <= 0:
		return fmt.Errorf("The variable '%s' was set to %d.",
			"Spokes", config.spokes)
	case config.rings <= 0:
		return fmt.Errorf("The variable '%s' was set to %d.",
			"Rings", config.rings)
	case config.rMaxMult <= 0:
		return fmt.Errorf("The variable '%s' was set to %g.",
			"RMaxMult", config.rMaxMult)
	case config.rMinMult <= 0:
		return fmt.Errorf("The variable '%s' was set to %g.",
			"RMinMult", config.rMinMult)
	case config.rKernelMult <= 0:
		return fmt.Errorf("The variable '%s' was set to %g.",
			"RKernelMult", config.rKernelMult)
	case config.eta <= 0:
		return fmt.Errorf("The variable '%s' was set to %g.",
			"Eta", config.eta)
	case config.order <= 0:
		return fmt.Errorf("The variable '%s' was set to %d.",
			"Order", config.order)
	case config.levels <= 0:
		return fmt.Errorf("The variable '%s' was set to %d.",
			"Levels", config.levels)
	case config.smoothingWindow <= 0:
		return fmt.Errorf("The variable '%s' was set to %d.",
			"SmoothingWindow", config.smoothingWindow)
	}

	if config.rMinMult >= config.rMaxMult {
		return fmt.Errorf("The variable '%s' was set to %g, but the " +
			"variable '%s' was set to %g.", "RMinMult", config.rMinMult,
			"RMaxMult", config.rMaxMult)
	}

	return nil
}

func (config *ShellConfig) Run(
	flags []string, gConfig *GlobalConfig, stdin []string,
) ([]string, error) {
	panic("NYI")
}