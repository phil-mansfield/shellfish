/*package cmd contains code for running shellfish in its various command
line modes */
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/phil-mansfield/shellfish/parse"
	"github.com/phil-mansfield/shellfish/version"
)

var ModeNames map[string]Mode = map[string]Mode{
	"id": &IDConfig{},
	"tree": &TreeConfig{},
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
	Run(flags []string, gConfig *GlobalConfig, stdin []string) ([]string, error)
}

// GlobalConfig is a config file used by every mode. It contains information on
// the directories that various files are stored in.
type GlobalConfig struct {
	version string

	snapshotFormat, snapshotType string
	haloDir, haloType string
	treeDir, treeType string
	memoDir string

	formatMins, formatMaxes []int64
	snapMin, snapMax int64

	validateFormats bool
}

var _ Mode = &GlobalConfig{}

// ReadConfig reads a config file and returns an error, if applicable.
func (config *GlobalConfig) ReadConfig(fname string) error {

	vars := parse.NewConfigVars("config")
	vars.String(&config.version, "Version", version.SourceVersion)
	vars.String(&config.snapshotFormat, "SnapshotFormat", "")
	vars.String(&config.snapshotType, "SnapshotType", "")
	vars.String(&config.haloDir, "HaloDir", "")
	vars.String(&config.haloType, "HaloType", "")
	vars.String(&config.treeDir, "TreeDir", "")
	vars.String(&config.treeType, "TreeType", "")
	vars.String(&config.memoDir, "MemoDir", "")
	vars.Ints(&config.formatMins, "FormatMins", []int64{})
	vars.Ints(&config.formatMaxes, "FormatMaxes", []int64{})
	vars.Int(&config.snapMin, "SnapMin", -1)
	vars.Int(&config.snapMax, "SnapMax", -1)
	vars.Bool(&config.validateFormats, "ValidateFormats", false)

	if err := parse.ReadConfig(fname, vars); err != nil { return err }
	return config.validate()
}

// validate checks that all the user-generated fields of GlobalConfig are
// properly set.
func (config *GlobalConfig) validate() error {
	major, minor, patch, err := version.Parse(config.version)
	if err != nil {
		return fmt.Errorf("I couldn't parse the 'Version' variable: %s",
			err.Error())
	}
	smajor, sminor, spatch, _ := version.Parse(version.SourceVersion)
	if major != smajor || minor != sminor || patch != spatch {
		return fmt.Errorf("The 'Version' variable is set to %s, but the " +
			"version of the source is %s",
			config.version, version.SourceVersion)
	}

	switch config.snapshotType {
	case "LGadget-2":
	case "":
		return fmt.Errorf("The 'SnapshotType variable isn't set.'")
	default:
		return fmt.Errorf("The 'SnapshotType' variable is set to '%s', " +
			"which I don't recognize.", config.snapshotType)
	}

	switch config.haloType {
	case "Rockstar":
	case "":
		return fmt.Errorf("The 'HaloType variable isn't set.'")
	default:
		return fmt.Errorf("The 'HaloType' variable is set to '%s', " +
		"which I don't recognize.", config.haloType)
	}

	switch config.treeType {
	case "consistent-trees":
	case "":
		return fmt.Errorf("The 'TreeType variable isn't set.'")
	default:
		return fmt.Errorf("The 'TreeType' variable is set to '%s', " +
		"which I don't recognize.", config.treeType)
	}

	if config.haloDir == "" {
		return fmt.Errorf("The 'HaloDir' variable isn't set.")
	} else if err = validateDir(config.haloDir); err != nil {
		return fmt.Errorf("The 'HaloDir' variable is set to '%s', but %s",
			config.haloDir, err.Error())
	}

	if config.treeDir == "" {
		return fmt.Errorf("The 'TreeDir' variable isn't set.")
	} else if err = validateDir(config.treeDir); err != nil {
		return fmt.Errorf("The 'TreeDir' variable is set to '%s', but %s",
			config.treeDir, err.Error())
	}

	if config.memoDir == "" {
		return fmt.Errorf("The 'MemoDir' variable isn't set.")
	} else if err = validateDir(config.memoDir); err != nil {
		return fmt.Errorf("The 'MemoDir' variable is set to '%s', but %s",
			config.memoDir, err.Error())
	}

	if err = validateFormat(config); err != nil {
		return err
	}

	return nil
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
	// TODO: This doesn't validate correctly.

	// This is wrong because of "%%" specifiers.
	specifiers := strings.Count(config.snapshotFormat, "%")

	if len(config.formatMins) != len(config.formatMaxes) {
		return fmt.Errorf("The lengths of the variables 'FormatMins' and" +
			"'FormatMaxes' are not equal")
	}

	switch {
	case config.snapMin == -1:
		return fmt.Errorf("The variable 'SnapMin' wasn't set.")
	case config.snapMax == -1:
		return fmt.Errorf("The variable 'SnapMax' wasn't set.")
	}

	if config.snapMin > config.snapMax {
		return fmt.Errorf("'SnapMin' is larger than 'SnapMax'")
	}

	if len(config.formatMins) + 1 != specifiers {
		return fmt.Errorf("The length of 'FormatMins' is %d, but there " +
		"are %d '%%' specifiers in the format string '%s'.",
			len(config.formatMins), specifiers, config.snapshotFormat,
		)
	}
	// This is wrong because it doesn't check that the properly formatted

	return nil
}

// ExampleConfig returns an example configuration file.
func (config *GlobalConfig) ExampleConfig() string {
	return fmt.Sprintf(`[config]
# Target version of shellfish. This option merely allows Shellfish to notice
# when its source an configuration files are not from the same version. It will
# not allow previous versions to be run from earlier versions.
#
# This variable defaults to the source version if not included.
Version = %s

# These variables describe the formats used by the files which Shellfish reads.
# If your simulation output uses a format not included here, you can submit a
# request for support on https://github.com/phil-mansfield/shellfish/issues,
# (or you can implement it yourself: Go is extremely similar to C and can be
# learned in about an hour: http://tour.golang.org/).
# Supported SnapshotTypes: LGadget-2
# Supported HaloTypes: Rockstar
# Supported TreeTypes: consistent-trees
SnapshotType = LGadget-2
HaloType = Rockstar
TreeType = consistent-trees

# These next five variables are neccessary evil due to the fact that there are
# a wide range of directory structures used in different simulations. They will
# be sufficient to specify the location of snapshots in the vast majority of
# cases.
# SnapshotFormat is a format string (a la printf()) which can be passed a
# snapshot index and an arbitrary number of block IDs. For example, if your
# directory structure was
#
# path/to/snapshots/
#     snap000/
#         file0_1.dat
#         file0_2.dat
#         file0_3.dat
#         file1_1.dat
#         file1_2.dat
#         file1_3.dat
#     snap001/
#         ...
#     snap250/
#         ...
#
# It could be specified with the following values:
SnapshotFormat = path/to/snapshots/snap%%03d/file%%d_%%d.dat
FormatMins = 0, 1
FormtMaxes = 1, 3
SnapMin = 0
SnapMax = 250

# Directory containing halo catalogs.
HaloDir = path/to/halos/dir/

# Directory containing merger tree.
TreeDir = path/to/merger/tree/dir/

# A directory you create the first time you run Shellfish for a particular
# simulation. Shellfish will memoize certain partial results in this directoy
# (most importantly: the first couple of halos in )
MemoDir = path/to/memo/dir/

# ValidateFormats checks the the specified halo files and snapshot catalogs all
# exist at startup before running any other code. Otherwise, these will be
# checked only immediately before a particular file is opened. In general,
# it's best to check this to false for short jobs because checking every file
# is a lot of system calls and can take minutes, although it's generally a good
# idea to check at least once after making the config file that you aren't
# accidentally specifying nonexistent files.
ValidateFormats = false`, version.SourceVersion)
}

// Run is a dummy method which allows GlobalConfig to conform to the Mode
// interface for testing purposes.
func (config *GlobalConfig) Run(
	flags []string, gConfig *GlobalConfig, stdin []string,
) ([]string, error) {
	panic("GlobalConfig.Run() should never be executed.")
}