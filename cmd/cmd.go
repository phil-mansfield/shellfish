/*package cmd contains code for running shellfish in its various command
line modes */
package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	
	"github.com/phil-mansfield/shellfish/cmd/env"
	"github.com/phil-mansfield/shellfish/parse"
	"github.com/phil-mansfield/shellfish/version"
)

var ModeNames map[string]Mode = map[string]Mode{
	"id":    &IDConfig{},
	"tree":  &TreeConfig{},
	"coord": &CoordConfig{},
	"prof": &ProfConfig{},
	"shell": &ShellConfig{},
	"stats": &StatsConfig{},
	"check": &CheckConfig{},
}

// Mode represents the interface used by the main binary when interacting with
// a given command line mode.
type Mode interface {
	// ReadConfig reads a mode-specific config file and stores its contents
	// within the Mode.
	ReadConfig(fname string, flags []string) error
	// ExampleConfig returns the text of an example config file of this mode.
	ExampleConfig() string
	// Run executes the mode. It takes
	// an initialized GlobalConfig struct, and a slice of lines representing the
	// contents of stdin. It will return a slice of lines that should be
	// written to stdout along with an error if one occurs.
	Run(gConfig *GlobalConfig, e *env.Environment, stdin []byte) ([]string, error)
}

// GlobalConfig is a config file used by every mode. It contains information on
// the directories that various files are stored in.
type GlobalConfig struct {
	env.ParticleInfo
	env.HaloInfo

	Version           string

	SnapshotType      string
	HaloType          string
	TreeType          string

	MemoDir           string

	HaloValueNames    []string
	HaloValueColumns  []int64
	HaloValueComments []string

	HaloPositionUnits string
	HaloRadiusUnits   string
	HaloMassUnits     string

	Endianness        string
	ValidateFormats   bool
	Threads           int64

	Logging           string

	GadgetDMTypeIndices []int64
	GadgetSingleMassIndices []int64
	GadgetPositionUnits float64
	GadgetMassUnits float64

	LGadgetNpartNum   int64

	NilSnapOmegaM float64
	NilSnapOmegaL float64
	NilSnapH100 float64
	NilSnapScaleFactors []float64
	NilSnapTotalWidth float64
}

var _ Mode = &GlobalConfig{}

// ReadConfig reads a config file and returns an error, if applicable.
func (config *GlobalConfig) ReadConfig(fname string, flags []string) error {

	vars := parse.NewConfigVars("config")
	vars.String(&config.Version, "Version", version.SourceVersion)
	vars.String(&config.SnapshotFormat, "SnapshotFormat", "")
	vars.String(&config.SnapshotType, "SnapshotType", "")
	vars.String(&config.HaloDir, "HaloDir", "")
	vars.String(&config.HaloType, "HaloType", "nil")
	vars.String(&config.TreeDir, "TreeDir", "")
	vars.String(&config.TreeType, "TreeType", "nil")
	vars.String(&config.MemoDir, "MemoDir", "")

	vars.Strings(&config.HaloValueNames, "HaloValueNames", []string{})
	vars.Ints(&config.HaloValueColumns, "HaloValueColumns", []int64{})
	vars.Strings(&config.HaloValueComments, "HaloValueComments", []string{})

	vars.String(&config.HaloPositionUnits, "HaloPositionUnits", "")
	vars.String(&config.HaloRadiusUnits, "HaloRadiusUnits", "")
	vars.String(&config.HaloMassUnits, "HaloMassUnits", "")

	vars.Strings(&config.SnapshotFormatMeanings,
		"SnapshotFormatMeanings", []string{})
	vars.String(&config.ScaleFactorFile, "ScaleFactorFile", "")
	vars.Ints(&config.BlockMins, "BlockMins", []int64{})
	vars.Ints(&config.BlockMaxes, "BlockMaxes", []int64{})
	vars.Int(&config.SnapMin, "SnapMin", -1)
	vars.Int(&config.SnapMax, "SnapMax", -1)
	vars.String(&config.Endianness, "Endianness", "SystemOrder")
	vars.Bool(&config.ValidateFormats, "ValidateFormats", false)

	vars.Int(&config.Threads, "Threads", -1)
	vars.String(&config.Logging, "Logging", "nil")

	vars.Ints(&config.GadgetDMTypeIndices,
		"GadgetDMTypeIndices", []int64{1})
	vars.Ints(&config.GadgetSingleMassIndices,
		"GadgetSingleMassIndices", []int64{1})
	vars.Float(&config.GadgetPositionUnits, "GadgetPositionUnits", 1.0)
	vars.Float(&config.GadgetMassUnits, "GadgetMassUnits", 1.0)

	vars.Int(&config.LGadgetNpartNum, "LGadgetNpartNum", 2)

	vars.Float(&config.NilSnapOmegaM, "NilSnapOmegaM", -1)
	vars.Float(&config.NilSnapOmegaL, "NilSnapOmegaL", -1)
	vars.Float(&config.NilSnapH100, "NilSnapH100", -1)
	vars.Floats(&config.NilSnapScaleFactors, "NilSnapScaleFactors", []float64{})
	vars.Float(&config.NilSnapTotalWidth, "NilSnapTotalWidth", -1)

	if err := parse.ReadConfig(fname, vars); err != nil {
		return err
	}
	config.HSnapMax = config.SnapMax
	config.HSnapMin = config.SnapMin
	
	return config.validate()
}

// validate checks that all the user-generated fields of GlobalConfig are
// properly set.
func (config *GlobalConfig) validate() error {
	major, minor, patch, err := version.Parse(config.Version)
	if err != nil {
		return fmt.Errorf("I couldn't parse the 'Version' variable: %s",
			err.Error())
	}
	smajor, sminor, spatch, _ := version.Parse(version.SourceVersion)
	if (smajor == 0 || major == 0) && (major != smajor || minor != sminor || patch != spatch) {
		return fmt.Errorf("The 'Version' variable is set to %s, but the "+
			"version of the source is %s",
			config.Version, version.SourceVersion)
	} else if (smajor == 1 && major == 1) && (minor > sminor || patch > spatch) {
		return fmt.Errorf("You are attempting to run a config file from the " + 
			"future: The 'Version' variable is set to %s, but the version of " +
			"the source code is %s", config.Version, version.SourceVersion)
	}

	switch config.SnapshotType {
	case "gotetra", "LGadget-2", "Gadget-2", "ARTIO", "Bolshoi", "nil":
	case "":
		return fmt.Errorf("The 'SnapshotType variable isn't set.'")
	default:
		return fmt.Errorf("The 'SnapshotType' variable is set to '%s', "+
			"which I don't recognize.", config.SnapshotType)
	}

	switch config.HaloType {
	case "Text", "nil":
	case "":
		return fmt.Errorf("The 'HaloType' variable isn't set.'")
	default:
		return fmt.Errorf("The 'HaloType' variable is set to '%s', "+
			"which I don't recognize.", config.HaloType)
	}

	if config.HaloType != "nil" {
		
		config.HaloPositionUnits = strings.Join(
			strings.Split(config.HaloPositionUnits, " "), "",
		)
		
		switch config.HaloPositionUnits {
		case "cMpc/h", "ckpc/h", "pMpc/h", "pkpc/h":
		case "cMpc", "ckpc", "pMpc", "pkpc":
		case "":
			return fmt.Errorf("The 'HaloPositionUnits' variable isn't set.")
		default:
			return fmt.Errorf("The 'HaloPositionUnits variable is set to '%s',"+
				" which I don't support. Only supported units are " +
				"ckpc/h and cMpc/h", config.HaloPositionUnits)
		}
		
		config.HaloRadiusUnits = strings.Join(
			strings.Split(config.HaloRadiusUnits, " "), "",
		)

		switch config.HaloRadiusUnits {
		case "cMpc/h","ckpc/h","pMpc/h","pkpc/h", "cMpc","ckpc","pMpc","pkpc":
		case "":
			return fmt.Errorf("The 'HaloRadiusUnits' variable isn't set.")
		default:
			return fmt.Errorf("The 'HaloRadiusUnits variable is set to '%s', "+
				"which I don't support. Only supported units are " +
				"ckpc/h, cMpc/h, pkpc/h, pMpc/h, ckpc, cMpc, pkpc, and pMpc",
				config.HaloPositionUnits)
		}

		config.HaloMassUnits = strings.Join(
			strings.Split(config.HaloMassUnits, " "), "",
		)

		switch config.HaloMassUnits {
		case "Msun/h":
		case "":
			return fmt.Errorf("The 'HaloMassUnits' variable isn't set.")
		default:
			return fmt.Errorf("The 'HaloMassUnits variable is set to '%s', "+
				"which I don't understand.", config.HaloPositionUnits)
		}
	} else {
		config.HaloRadiusUnits = "cMpc/h"
		config.HaloPositionUnits = "cMpc/h"
		config.HaloMassUnits = "Msun/h"
	}
	
	switch config.TreeType {
	case "consistent-trees", "nil":
	case "":
		return fmt.Errorf("The 'TreeType variable isn't set.'")
	default:
		return fmt.Errorf("The 'TreeType' variable is set to '%s', "+
			"which I don't recognize.", config.TreeType)
	}

	if config.HaloType != "nil" {
		if config.HaloDir == "" {
			return fmt.Errorf("The 'HaloDir' variable isn't set.")
		} else if err = validateDir(config.HaloDir); err != nil {
			return fmt.Errorf("The 'HaloDir' variable is set to '%s', but %s",
				config.HaloDir, err.Error())
		}
	}

	if config.HaloType != "nil" {
		if config.TreeDir == "" {
			return fmt.Errorf("The 'TreeDir' variable isn't set.")
		} else if err = validateDir(config.TreeDir); err != nil {
			return fmt.Errorf("The 'TreeDir' variable is set to '%s', but %s",
				config.TreeDir, err.Error())
		}
	}

	if config.MemoDir == "" {
		return fmt.Errorf("The 'MemoDir' variable isn't set.")
	} else if err = validateDir(config.MemoDir); err != nil {
		return fmt.Errorf("The 'MemoDir' variable is set to '%s', but %s",
			config.MemoDir, err.Error())
	}

	if config.HaloType != "nil" {
		if len(config.HaloValueNames) == 0 {
			return fmt.Errorf("The 'HaloValueNames' variable isn't set.")
		} else if len(config.HaloValueColumns) == 0 {
			return fmt.Errorf("The 'HaloValueColumns' variable isn't set.")
		} else if len(config.HaloValueNames) != len(config.HaloValueColumns) {
			return fmt.Errorf("len(HaloValueColumns) = %d, but " +
				"len(HaloValueNames) = %d.", len(config.HaloValueColumns),
				len(config.HaloValueNames))
		}
		
		switch {
		case !inStringSlice("ID", config.HaloValueNames):
			return fmt.Errorf(
				"'HaloValueNames' does not contain the 'ID' name.",
			)
		case !inStringSlice("X", config.HaloValueNames):
			return fmt.Errorf(
				"'HaloValueNames' does not contain the 'X' name.",
			)
		case !inStringSlice("Y", config.HaloValueNames):
			return fmt.Errorf(
				"'HaloValueNames' does not contain the 'Y' name.",
			)
		case !inStringSlice("Z", config.HaloValueNames):
			return fmt.Errorf(
				"'HaloValueNames' does not contain the 'Z' name.",
			)
		case !inStringSlice("M200m", config.HaloValueNames):
			return fmt.Errorf(
				"'HaloValueNames' does not contain the 'M200m' name.",
			)
		}
	}

	if config.SnapshotType == "nil" {
		switch {
		case config.NilSnapOmegaM == -1:
			return fmt.Errorf(
				"'NilSnapOmegaM' not set even though SnapshotType == 'nil'",
			)
		case config.NilSnapOmegaL == -1:
			fmt.Println()
		case config.NilSnapH100 == -1:
			return fmt.Errorf(
				"'NilSnapOmegaM' not set even though SnapshotType == 'nil'",
			)
			fmt.Println()
		case config.NilSnapTotalWidth == -1:
			return fmt.Errorf(
				"'NilSnapOmegaM' not set even though SnapshotType == 'nil'",
			)
			fmt.Println()
		case len(config.NilSnapScaleFactors) == 0:
			return fmt.Errorf(
				"'NilSnapOmegaM' not set even though SnapshotType == 'nil'",
			)
			fmt.Println()
		}
	}


	switch config.Endianness {
	case "":
		return fmt.Errorf("The variable 'Endianness' was not set.")
	case "LittleEndian", "BigEndian", "SystemOrder":
	default:
		return fmt.Errorf("The variable 'Endianness' must be sent to " +
			"either 'SystemOrder', 'LittleEndian', or 'BigEndian'.")
	}
	
	if len(config.HaloValueNames) != len(config.HaloValueColumns) {
		return fmt.Errorf(
			"len(HaloValueNames) = %d, but len(HaloValueColumns = %d)",
			len(config.HaloValueNames), len(config.HaloValueColumns),
		)
	} else if len(config.HaloValueNames) != len(config.HaloValueComments) {
		return fmt.Errorf(
			"len(HaloValueNames) = %d, but len(HaloValueComments = %d)",
			len(config.HaloValueNames), len(config.HaloValueComments),
		)
	}

	if config.LGadgetNpartNum > 2 || config.LGadgetNpartNum <= 0 {
		return fmt.Errorf(
			"GadgetNpartNum set to %d, but the only valid values are 1 and 2.",
			config.LGadgetNpartNum,
		)
	}

	
	return validateFormat(config)
}

func inStringSlice(x string, xs []string) bool {
	for _, xx := range xs {
		if x == xx {
			return true
		}
	}
	return false
}

// validateDir returns an error if there are any problems with the given
// directory.
func validateDir(name string) error {
	if info, err := os.Stat(name); err != nil {
		//return fmt.Errorf("%s does not exist.", name)
		return os.MkdirAll(name, os.ModeDir)
	} else if !info.IsDir() {
		return fmt.Errorf("%s is not a directory.", name)
	}

	return nil
}

// validateFormat returns an error if there are any problems with the
// given format variables.
func validateFormat(config *GlobalConfig) error {
	// TODO: This doesn't validate formats correctly.

	// This is wrong because of "%%" specifiers.
	specifiers := strings.Count(config.SnapshotFormat, "%")

	if len(config.BlockMins) != len(config.BlockMaxes) {
		return fmt.Errorf("The lengths of the variables 'FormatMins' and" +
			"'FormatMaxes' are not equal")
	}

	switch {
	case config.SnapMin == -1:
		return fmt.Errorf("The variable 'SnapMin' wasn't set.")
	case config.SnapMax == -1:
		return fmt.Errorf("The variable 'SnapMax' wasn't set.")
	}

	if config.SnapMin > config.SnapMax {
		return fmt.Errorf("'SnapMin' is larger than 'SnapMax'")
	}
	for i := range config.BlockMins {
		if config.BlockMins[i] > config.BlockMaxes[i] {
			return fmt.Errorf(
				"'FormatMins'[%d] is larger than 'FormatMaxes'[%d]", i, i,
			)
		}
	}

	if len(config.SnapshotFormatMeanings) == 0 && specifiers != 0 {
		return fmt.Errorf("'SnapshotFormatMeanings' was not set.")
	}

	if specifiers != len(config.SnapshotFormatMeanings) {
		return fmt.Errorf("The length of 'SnapshotFormatMeanings' is not " +
			"equal to the number of specifiers in 'SnapshotFormat'.")
	}

	for i, meaning := range config.SnapshotFormatMeanings {
		switch {
		case meaning == "ScaleFactor":
		case meaning == "Snapshot":
		case meaning == "Block":
		case len(meaning) > 5 && meaning[:5] == "Block":
			ending := meaning[5:]
			n, err := strconv.Atoi(ending)
			if err != nil {
				goto nextCase
			}
			if n < 0 || n >= len(config.BlockMaxes) {
				return fmt.Errorf("'SnapshotFormatMeaning'[%d] specifies an "+
					"invalid block range.", i)
			}
			return nil
		nextCase:
			fallthrough
		default:
			return fmt.Errorf("I don't understand '%s' from "+
				"'SnapshotFormatMeaning'[%d]", meaning, i)
		}
	}
	
	switch config.Logging {
	case "nil", "performance", "debug":
	default:
		return fmt.Errorf("I don't recognize the Logging mode '%s'.",
			config.Logging)
	}

	return nil
}

// ExampleConfig returns an example configuration file.
func (config *GlobalConfig) ExampleConfig() string {
	return fmt.Sprintf(`[config]
# Target version of shellfish. This option merely allows Shellfish to notice
# when its source and configuration files are not from the same version. It will
# not allow previous versions to be run from earlier versions.
Version = %s

# These variables describe the formats used by the files which Shellfish reads.
# If your simulation output uses a format not included here, you can submit a
# request for support on https://github.com/phil-mansfield/shellfish/issues.
#
# The "nil" flag tells shellfish that you won't try to use any of the tools
# related to this type of tile. If you set HaloType to nil, you can't use
# id, coord, or tree, and if you set TreeType to nil, you can't use tree. Do
# this if you're planning to read your own halo catlogs/merger trees and hand
# input directly to shellfish's shell mode. If HaloType is nil, you don't need
# to fill out any of the Halo* variables in this config file and if TreeType
# is nil, you don't need to fill out any of the Tree* variables.
#
# Supported SnapshotTypes: LGadget-2, gotetra, Gadget-2 (experimental),
# ARTIO (experimental), Bolshoi (experimental)
# Supported HaloTypes: Text, nil
# Supported TreeTypes: consistent-trees, nil
SnapshotType = LGadget-2
HaloType = Text
TreeType = consistent-trees

# HaloValueNames and HaloValueColumns tell Shellfish about the structure of your
# halo catalogs. It needs to be able to read three pieces of information: the
# halos IDs, the positions and the masses. HaloValueNames should be an ordered
# list of columns (the set below is the minimum required) and the 0-indexed
# columns of your halo catalog that they appear in.
HaloValueNames = ID, X, Y, Z, M200m
HaloValueColumns = 0, 2, 3, 4, 20
# HaloValueComments can be used to include notes about, e.g. units in output
# catalogs. These are not analyzed by Shellfish in any way, but will be
# propagated to output catalogs when relevant.
HaloValueComments = "int", "cMpc/h", "cMpc/h", "cMpc/h", "Msun/h"

# HaloPositionUnits are the units which your halo catalog reports positions in.
# Currently supported values are "cMpc/h" and "ckpc/h" (the "c" stands for
# "comoving") and "pMpc/h" and "pkpc/h" (the "p" stands for "physical"). The
# same units without the "/h" are also supported.
HaloPositionUnits = cMpc/h
# HaloRadiusUnits are the units which your halo catalog reports radii in.
# Currently supported values are "cMpc/h" and "ckpc/h" (the "c" stands for
# "comoving").
HaloRadiusUnits = ckpc/h
# HaloMassUnits are the units which your halo catalog reports masses in.
# Currently only "Msun/h" is supported.
HaloMassUnits = Msun/h

# These next couple of variables are neccessary evils due to the fact that there
# are a wide range of directory structures used in different simulations. They
# will be sufficient to specify the location of snapshots in the vast majority
# of cases. I give an in-depth description of how to use them in the file
# doc/directory_config.md.
SnapshotFormat = path/to/snapshots/snapdir_%%03d/snapshot_%%03d.%%d
# Valid values are "Snapshot", "ScaleFactor", "Block", "Block0", "Block1", etc.
# BlockN will reference the Nth element of the BlockMins and Block Maxes
# variables.
SnapshotFormatMeanings = Snapshot, Snapshot, Block
BlockMins = 0
BlockMaxes = 511
SnapMin = 0
SnapMax = 100

# ScaleFactorFile should only be set if one of the elements of
# SnapshotFormatMeanings is 'ScaleFactor'. This should point to a file which
# contains the scale factors of your files. A file like this can usually be
# generated in a few lines of Python: look in doc/scale_factor_ex.py
# for an example.
# ScaleFactorFile = path/to/file.txt

# Directory containing halo catalogs. It is assumed that when the catalog
# catalog files in this directory are sorted in alphabetical order (really: in
# lexicographical order), they will also be sorted temporally. It's also assumed
# that there are no missing snapshots in the middle of your simulation (e.g.
# if snapshot #83 and #85 exist, but snapshot #84 got corrupted and was deleted,
# that would be bad, but if you don't have catalogs for the firt ten snapshots
# your simulation, that would be fine.)
#
# If either of these isn't true, you'll still be able to use the shell finding
# and analyzing parts of Shellfish, but won't be able to use its catalog reading
# tools.
HaloDir = path/to/halos/dir/

# Directory containing merger tree.
TreeDir = path/to/merger/tree/dir/

# A directory you create the first time you run Shellfish for a particular
# simulation. Shellfish will cache certain partial results in this directoy.
# Every time a value is changed in this file, you must change the location of
# this directory.
# ("memo" is a reference to the term "memoization," which is just  a fancy
# word for caching.)
MemoDir = path/to/memo/dir/

# Endianness of any external binary data files read by Shellfish. It should be
# set to either SystemOrder, LittleEndian, BigEndian. This variable defaults to
# SystemOrder.
#
# (Any binaries data files _written_ by Shellfish will ignore this variable and
# will be written in little endian order. This and a few other details allows
# binary files written by Shellfish on one machine to be read by it on any other
# machine.)
Endianness = SystemOrder

# Threads is the number of threads that should be run simultaneously. If Threads
# is set to a non-positive value (as it is by default), it will automatically
# be set equal to the number of available cores on the current node. All threads
# will be balanced across available cores. Setting this to a value larger than
# the number of cores on the node might result in slightly suboptimal
# performance.
Threads = -1

# The logging mode to be used. There are three different logging modes:
# nil - no logging is performed.
# performance - runtime and memory consumption logging are written to stderr.
# debugging - debugging information is written to stderr
Logging = nil

###############################
## Format-specific variables ##
###############################
# If SnapshotType is set to Gadget-2, LGadget-2, or nil, extra information
# will need to be provided to read your files.

###############################
## Gadget-specific variables ##
###############################

# GadgetDMTypeIndices indicates which particle types correspond to dark matter
# particles. For a typical uniform mass DM-only simulation, this will be 1. For
# simulations with particles of multiple masses, more than one index may be
# used. This only needs to be set if SnapshßotType = Gadget-2.
# GadgetDMTypeIndices = 1

# GadgetSingleMassIndices indicates which particle types don't have entries in
# the MASS/Masses block and instead use the the Massarr/MassTable entry in the
# header. Include non-DM particle types. This only needs to be set if
# SnapshotType = Gadget-2.
# GadgetSingleMassIndices = 0, 1, 2, 3, 4, 5

# GadgetPositionUnits indicates how positions are stored within your Gadget
# snapshot. Set this variable so the following equation is true:
# (1 Mpc/h) * GadgetPositionUnits = (Your position units).
# (i.e. if your position units are smaller than 1 Mpc/h, this variable should
# be less than one). This variable only needs to be set if
# SnapshotType = Gadget-2 and your units are not 1 Mpc/h.
# GadgetPositionUnits = 1.0

# GadgetMassUnits indicates how positions are stored within your Gadget
# snapshot. Set this variable so the following equation is true:
# (1 Msun/h) * GadgetMassUnits = (Your mass units).
# (i.e. if your mass units are smaller than 1 Msun/h, this variable should be
# less than one). This variable only needs to be set if SnapshotType = Gadget-
# and your units are not 1 Msun/h.
# GadgetMassUnits = 1.0

################################
## LGadget-specific variables ##
################################

# LGadgetNpartNum is an optional variable which should only be set when using
# LGadget files. If your LGadget files use two elements of Npart and Nall to
# represent the number of dark matter particles in your simulation, set this
# variable to 2. If every element corresponds to a different particle species,
# set this variable to 1. LGadgetNpartNum defaults to the most common value, 2.
#
# If you don't know what your LGadget file does, leave this alone: Shellfish will
# fail and tell you to change this variable.
# LGadgetNpartNum = 2

##########################################
## nil (SnapshotType)-specifc variables ##
##########################################
# You will need to provide basic cosmological information which is normally
# contained in snapshot headers. The example values are for the Bolshoi halo
# catalogs.

# NilSnapOmegaM = 0.27
# NilSnapOmegaL = 0.73
# NilSnapH100 = 0.7
# In ascending order:
# NilSnapScaleFactors = [0.06635, 0.07835, 0.09635, 0.10235, 0.10835, 0.11435, 0.12035, 0.13235, 0.13835, 0.14435, 0.15035, 0.15635, 0.16235, 0.16835, 0.17435, 0.18035, 0.18635, 0.19235, 0.19835, 0.20235, 0.20435, 0.21035, 0.21635, 0.22235, 0.22835, 0.23435, 0.24635, 0.25235, 0.25835, 0.26435, 0.27035, 0.27635, 0.28235, 0.28835, 0.29435, 0.30635, 0.31235, 0.31835, 0.32435, 0.33035, 0.33635, 0.34235, 0.34835, 0.35435, 0.36035, 0.36635, 0.37235, 0.37835, 0.38435, 0.39035, 0.39635, 0.40235, 0.40835, 0.41435, 0.42035, 0.42635, 0.43235, 0.43835, 0.44435, 0.45035, 0.45635, 0.46235, 0.46835, 0.47435, 0.48035, 0.48635, 0.49835, 0.50435, 0.51035, 0.51635, 0.52235, 0.52835, 0.53235, 0.53835, 0.54435, 0.55035, 0.55635, 0.56235, 0.56835, 0.57435, 0.58035, 0.58635, 0.59235, 0.59835, 0.60435, 0.61035, 0.61635, 0.62235, 0.62835, 0.63435, 0.64035, 0.64635, 0.65235, 0.65835, 0.66435, 0.67035, 0.67635, 0.68235, 0.68835, 0.69435, 0.70035, 0.70635, 0.71235, 0.71835, 0.72435, 0.73035, 0.73635, 0.74235, 0.74835, 0.75435, 0.76035, 0.76635, 0.77235, 0.77835, 0.78435, 0.79035, 0.79635, 0.80235, 0.80835, 0.81135, 0.81435, 0.81735, 0.82035, 0.82335, 0.82635, 0.82935, 0.83235, 0.83535, 0.83835, 0.84135, 0.84435, 0.84735, 0.85035, 0.85335, 0.85635, 0.85935, 0.86235, 0.86535, 0.86835, 0.87135, 0.87435, 0.87735, 0.88035, 0.88335, 0.88635, 0.88935, 0.89235, 0.89535, 0.89835, 0.90135, 0.90435, 0.90735, 0.91035, 0.91335, 0.91635, 0.91935, 0.92235, 0.92535, 0.92835, 0.93135, 0.93435, 0.93735, 0.94335, 0.94635, 0.94935, 0.95235, 0.95835, 0.96135, 0.96435, 0.96735, 0.97035, 0.97335, 0.97635, 0.97935, 0.98235, 0.98535, 0.98835, 0.99135, 0.99435, 0.99735, 1.00035]
# Measured in Mpc/h:
# NilSnapTotalWidth = 250
`, version.SourceVersion)
}

// Run is a dummy method which allows GlobalConfig to conform to the Mode
// interface for testing purposes.
func (config *GlobalConfig) Run(
	gConfig *GlobalConfig, e *env.Environment, stdin []byte,
) ([]string, error) {
	panic("GlobalConfig.Run() should never be executed.")
}

// This needs to be global for debugging purposes.
var randSeed = uint64(time.Now().UnixNano())
