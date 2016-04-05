/*package cmd contains code for running shellfish in its various command
line modes */
package cmd

import (
	"fmt"

	"github.com/phil-mansfield/shellfish/parse"
	"github.com/phil-mansfield/shellfish/version"
)

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
	formatRange []int64
	snapshotFormatIndex int64
}

var _ Mode = &GlobalConfig{}

// ReadConfig reads a config file and returns an error, if applicable.
func (config *GlobalConfig) ReadConfig(fname string) error {
	config.version = version.SourceVersion

	vars := parse.NewConfigVars("config")
	vars.String(&config.version, "Version")
	vars.String(&config.snapshotFormat, "SnapshotFormat")
	vars.String(&config.snapshotType, "SnapshotType")
	vars.String(&config.haloDir, "HaloDir")
	vars.String(&config.haloDir, "HaloType")
	vars.String(&config.treeDir, "TreeDir")
	vars.String(&config.snapshotType, "TreeType")
	vars.String(&config.memoDir, "MemoDir")
	vars.Ints(&config.formatRange, "FormatRange")
	vars.Int(&config.snapshotFormatIndex, "SnapshotFormatIndex")

	err := parse.ReadConfig(fname, vars)
	if err != nil { return err }

	return nil
}

// ExampleConfig returns an example configuration file.
func (config *GlobalConfig) ExampleConfig() string {
	return fmt.Sprintf(`[config]
# Target version of shellfish. This option merely allows Shellfish to notice
# when its source an configuration files are not from the same version. It will
# not allow previous versions to be run from earlier versions.
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

# These next three variables are neccessary evil due to the fact that there are
# a wide range of directory structures used in different simulations. They will
# be sufficient to specify the location of snapshots in the vast majority of
# cases.
# SnapshotFormat is a format string (a la printf()) which can be passed a
# snapshot index and an arbitrary number of block IDs. For example, if your
# directory structure was
# path/to/snapshots/
#     snap000/
#         file0_0.dat
#         file0_1.dat
#         file0_2.dat
#         file1_0.dat
#         file1_1.dat
#         file1_2.dat
#     snap001/
#         ...
# It could be specified with the following values:
SnapshotFormat = path/to/snapshots/snap%%03d/file%%d_%%d.dat
FormatRange = 2, 3
SnapshotFormatIndex = 0

# Directory containing halo catalogs.
HaloDir = path/to/halos/dir/

# Directory containing merger tree.
TreeDir = path/to/merger/tree/dir/

# A directory you create the first time you run Shellfish for a particular
# simulation. Shellfish will memoize certain partial results in this directoy
# (most importantly: the first couple of halos in )
MemoDir = path/to/memo/dir/`, version.SourceVersion)
}

// Run is a dummy method which allows GlobalConfig to conform to the Mode
// interface for testing purposes.
func (config *GlobalConfig) Run(
	flags []string, gConfig *GlobalConfig, stdin []string,
) ([]string, error) {
	panic("GlobalConfig.Run() should never be executed.")
}