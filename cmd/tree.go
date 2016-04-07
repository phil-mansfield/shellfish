package cmd

type TreeConfig struct {

}

var _ Mode = &TreeConfig{}

func (config *TreeConfig) ExampleConfig() string { return "" }

func (config *TreeConfig) ReadConfig(fname string) error { return nil }

func (config *TreeConfig) validate() error { return nil }

func (config *TreeConfig) Run(
	flags []string, gConfig *GlobalConfig, stdin []string,
) ([]string, error) {
	panic("NYI")
}