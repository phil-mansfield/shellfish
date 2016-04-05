/*package cmd contains code for running shellfish in its various command
line modes */
package cmd

// Mode represents the interface used by the main binary when interacting with
// a given command line mode.
type Mode interface {
	// ReadConfig reads a mode-specific config file and stores its contents
	// within the Mode.
	ReadConfig(name string) error
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
	snapshotFmt, snapshotType string
	rockstarDir, treeDir, memorDir string
	snapshotRange []int
}