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

	LGadgetNpartNum   int64
}

var _ Mode = &GlobalConfig{}

// ReadConfig reads a config file and returns an error, if applicable.
func (config *GlobalConfig) ReadConfig(fname string, flags []string) error {

	vars := parse.NewConfigVars("config")
	vars.String(&config.Version, "Version", version.SourceVersion)
	vars.String(&config.SnapshotFormat, "SnapshotFormat", "")
	vars.String(&config.SnapshotType, "SnapshotType", "")
	vars.String(&config.HaloDir, "HaloDir", "")
	vars.String(&config.HaloType, "HaloType", "")
	vars.String(&config.TreeDir, "TreeDir", "")
	vars.String(&config.TreeType, "TreeType", "")
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
	vars.String(&config.Endianness, "Endianness", "")
	vars.Bool(&config.ValidateFormats, "ValidateFormats", false)

	vars.Int(&config.Threads, "Threads", -1)
	vars.String(&config.Logging, "Logging", "nil")

	vars.Ints(&config.GadgetDMTypeIndices,
		"GadgetDMTypeIndices", []int64{1})
	vars.Ints(&config.GadgetSingleMassIndices,
		"GadgetSingleMassIndices", []int64{1})
	vars.Int(&config.LGadgetNpartNum, "LGadgetNpartNum", 2)

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
	case "gotetra", "LGadget-2", "Gadget-2", "ARTIO":
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
				"'HaloValueNames' does not contain the 'X' name.",
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
# Supported SnapshotTypes: Gadget-2, LGadget-2, ARTIO (experimental), gotetra
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

# ValidateFormats checks the the specified halo files and snapshot catalogs all
# exist at startup before running any other code. Otherwise, these will be
# checked only immediately before a particular file is opened. In general,
# it's best to set this to false for short jobs because checking every file
# is a lot of system calls and can take minutes. That said, it's generally a
# good idea to check at least once after making the config file that you aren't
# accidentally specifying nonexistent files.
ValidateFormats = false

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

# GadgetDMTypeIndices indicates which particle types correspond to dark matter
# particles. For a typical uniform mass DM-only simulation, this will be 1. For
# simulations with particles of multiple masses, more than one index may be
# used.
# GadgetDMTypeIndices = 1

# GadgetSingleMassIndices indicates which particle types don't have entries in
# the MASS/Masses block and instead use the the Massarr/MassTable entry in the
# header. Include particle non-DM particle types.
# GadgetSingleMassIndices = 0, 1, 2, 3, 4, 5

# LGadgetNpartNum is an optional variable which should only be set when using
# LGadget files. If your LGadget files use two elements of Npart and Nall to
# represent the number of dark matter particles in your simulation, set this
# variable to 2. If every element corresponds to a different particle species,
# set this variable to 1. GadgetNpartNum defaults to the most common value, 2.
#
# If you don't know what your Gadget file does, leave this alone: Shellfish will
# fail and tell you to change this variable.
# LGadgetNpartNum = 2
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
