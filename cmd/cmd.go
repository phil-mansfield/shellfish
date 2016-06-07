/*package cmd contains code for running shellfish in its various command
line modes */
package cmd

import (
	"fmt"
	//"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/phil-mansfield/shellfish/cmd/env"
	"github.com/phil-mansfield/shellfish/parse"
	"github.com/phil-mansfield/shellfish/version"
	"github.com/phil-mansfield/shellfish/logging"
)

var ModeNames map[string]Mode = map[string]Mode{
	"id":    &IDConfig{},
	"tree":  &TreeConfig{},
	"coord": &CoordConfig{},
	"shell": &ShellConfig{},
	"stats": &StatsConfig{},
}

// Mode represents the interface used by the main binary when interacting with
// a given command line mode.
type Mode interface {
	// ReadConfig reads a mode-specific config file and stores its contents
	// within the Mode.
	ReadConfig(fname string) error
	// ExampleConfig returns the text of an example config file of this mode.
	ExampleConfig() string
	// Run executes the mode. It takes a list of tokenized command line flags,
	// an initialized GlobalConfig struct, and a slice of lines representing the
	// contents of stdin. It will return a slice of lines that should be
	// written to stdout along with an error if one occurs.
	Run(flags []string, gConfig *GlobalConfig,
		e *env.Environment, stdin []string) ([]string, error)
}

// GlobalConfig is a config file used by every mode. It contains information on
// the directories that various files are stored in.
type GlobalConfig struct {
	env.ParticleInfo
	env.HaloInfo

	Version string

	SnapshotType string
	HaloType     string
	TreeType     string

	MemoDir string

	HaloIDColumn        int64
	HaloM200mColumn     int64
	HaloPositionColumns []int64

	HaloPositionUnits string
	HaloMassUnits     string

	Endianness      string
	ValidateFormats bool
	Threads         int64

	Logging string
}

var _ Mode = &GlobalConfig{}

// ReadConfig reads a config file and returns an error, if applicable.
func (config *GlobalConfig) ReadConfig(fname string) error {

	vars := parse.NewConfigVars("config")
	vars.String(&config.Version, "Version", version.SourceVersion)
	vars.String(&config.SnapshotFormat, "SnapshotFormat", "")
	vars.String(&config.SnapshotType, "SnapshotType", "")
	vars.String(&config.HaloDir, "HaloDir", "")
	vars.String(&config.HaloType, "HaloType", "")
	vars.String(&config.TreeDir, "TreeDir", "")
	vars.String(&config.TreeType, "TreeType", "")
	vars.String(&config.MemoDir, "MemoDir", "")

	vars.Int(&config.HaloIDColumn, "HaloIDColumn", -1)
	vars.Int(&config.HaloM200mColumn, "HaloM200mColumn", -1)
	vars.Ints(&config.HaloPositionColumns, "HaloPositionColumns",
		[]int64{-1, -1, -1})

	vars.String(&config.HaloPositionUnits, "HaloPositionUnits", "")
	vars.String(&config.HaloMassUnits, "HaloMassUnits", "Msun/h")

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

	if err := parse.ReadConfig(fname, vars); err != nil {
		return err
	}
	config.HSnapMax = config.SnapMax
	config.HSnapMin = config.SnapMin

	switch config.Logging {
	case "nil": logging.Mode = logging.Nil
	case "performance": logging.Mode = logging.Performance
	case "debug": logging.Mode = logging.Debug
	}

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
	if major != smajor || minor != sminor || patch != spatch {
		return fmt.Errorf("The 'Version' variable is set to %s, but the "+
			"version of the source is %s",
			config.Version, version.SourceVersion)
	}

	switch config.SnapshotType {
	case "gotetra", "LGadget-2", "ARTIO":
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

	config.HaloPositionUnits = strings.Join(
		strings.Split(config.HaloPositionUnits, " "), "",
	)

	switch config.HaloPositionUnits {
	case "Mpc/h":
	case "":
		return fmt.Errorf("The 'HaloPositionUnits' variable isn't set.")
	default:
		return fmt.Errorf("The 'HaloPositionUnits variable is set to '%s', "+
			"which I don't understand.", config.HaloPositionUnits)
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

	switch config.TreeType {
	case "consistent-trees", "nil":
	case "":
		return fmt.Errorf("The 'TreeType variable isn't set.'")
	default:
		return fmt.Errorf("The 'TreeType' variable is set to '%s', "+
			"which I don't recognize.", config.TreeType)
	}

	if config.HaloDir == "" {
		return fmt.Errorf("The 'HaloDir' variable isn't set.")
	} else if err = validateDir(config.HaloDir); err != nil {
		return fmt.Errorf("The 'HaloDir' variable is set to '%s', but %s",
			config.HaloDir, err.Error())
	}

	if config.TreeDir == "" {
		return fmt.Errorf("The 'TreeDir' variable isn't set.")
	} else if err = validateDir(config.TreeDir); err != nil {
		return fmt.Errorf("The 'TreeDir' variable is set to '%s', but %s",
			config.TreeDir, err.Error())
	}

	if config.MemoDir == "" {
		return fmt.Errorf("The 'MemoDir' variable isn't set.")
	} else if err = validateDir(config.MemoDir); err != nil {
		return fmt.Errorf("The 'MemoDir' variable is set to '%s', but %s",
			config.MemoDir, err.Error())
	}

	if config.HaloIDColumn == -1 {
		return fmt.Errorf("The 'HaloIDColumn' variable isn't set.")
	} else if config.HaloM200mColumn == -1 {
		return fmt.Errorf("The 'HaloM200mColumn' variable isn't set.")
	} else if len(config.HaloPositionColumns) != 3 {
		return fmt.Errorf("The 'HaloPositionColumns' variable must have " +
			"three elements.")
	} else if config.HaloPositionColumns[0] == -1 ||
		config.HaloPositionColumns[1] == -1 ||
		config.HaloPositionColumns[2] == -1 {
		return fmt.Errorf("The 'HaloPositionColumns' variable wasn't set.")
	}

	switch config.Endianness {
	case "":
		return fmt.Errorf("The variable 'Endianness' was not set.")
	case "LittleEndian", "BigEndian", "SystemOrder":
	default:
		return fmt.Errorf("The variable 'Endianness' must be sent to " +
			"either 'SystemOrder', 'LittleEndian', or 'BigEndian'.")
	}

	return validateFormat(config)
}

// validateDir returns an error if there are any problems with the given
// directory.
func validateDir(name string) error {
	if info, err := os.Stat(name); err != nil {
		return fmt.Errorf("%s does not exist.", name)
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

	if len(config.SnapshotFormatMeanings) == 0 {
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
# request for support on https://github.com/phil-mansfield/shellfish/issues,
# (or you can implement it yourself: Go is extremely similar to C and can be
# learned in about an hour: http://tour.golang.org/).
#
# Supported SnapshotTypes: LGadget-2, gotetra
# Supported HaloTypes: Text, nil
# Supported TreeTypes: consistent-trees, nil
#
# Note the 'nil' type. This allows you to use unsupported halo types by piping
# coordinates directly into 'shellfish shell'.
SnapshotType = LGadget-2
HaloType = Text
TreeType = consistent-trees

# HaloPositionUnits = Mpc/h
# HaloMassUnits = Msun/h

# These variables specify which columns of your halo catalogs correspond to
# the variables that Shellfish needs to read.
HaloIDColumn = -1
HaloM200mColumn = -1
# HaloPositionColumns should correspond to the X, Y, and Z columns,
# respectively.
HaloPositionColumns = -1, -1, -1

# These next couple of variables are neccessarys evil due to the fact that there
# are a wide range of directory structures used in different simulations. They
# will be sufficient to specify the location of snapshots in the vast majority
# of cases. I give an in-dpeth description of how to use them in the file
# doc/directory_config.md.
SnapshotFormat = path/to/snapshots/snapdir_%%03d/snapshot_%%03d.%%d
# Valid values are "Snapshot", "ScaleFactor", "Block", "Block0", "Block1", ...
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

# Directory containing halo catalogs.
HaloDir = path/to/halos/dir/

# Directory containing merger tree.
TreeDir = path/to/merger/tree/dir/

# A directory you create the first time you run Shellfish for a particular
# simulation. Shellfish will memoize certain partial results in this directoy.
# Every time a value is changed in this file, you must change the location of
# this directory.
MemoDir = path/to/memo/dir/

# Endianness of any external binary files read by Shellfish. It should be set
# to either SystemOrder, LittleEndian, BigEndian. This variable defaults to
# SystemOrder.
#
# (Any _internal binaries_ written by Shellfish will ignore this variable and
# will be written in little endian order.)
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
# is set to a non-positive value (as it iis by default), it will automatically
# be set equal to the number of available cores on the current node. All threads
# will be balanced across available cores. Setting this to a value larger than
# the number of cores on the node will result in slightly suboptimal
# performance.
Threads = -1

# The loggign mode to be used. There are three different logging modes:
# nil - no logging is performed.
# performance - runtime and memory consumption logging are written to stderr.
# debugging - debugging loggig is written to stderr
Logging = nil`, version.SourceVersion)
}

// Run is a dummy method which allows GlobalConfig to conform to the Mode
// interface for testing purposes.
func (config *GlobalConfig) Run(
	flags []string, gConfig *GlobalConfig, e *env.Environment, stdin []string,
) ([]string, error) {
	panic("GlobalConfig.Run() should never be executed.")
}
