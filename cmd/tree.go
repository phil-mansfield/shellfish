package cmd

type TreeConfig struct {

}

var _ Mode = &TreeConfig{}

func (config *TreeConfig) ExampleConfig() string {
	panic("NYI")
}

func (config *TreeConfig) ReadConfig(fname string) error {
	panic("NYI")
}

func (config *TreeConfig) validate() error {
	panic("NYI")
}

func (config *TreeConfig) Run(
	flags []string, gConfig *GlobalConfig, stdin []string,
) ([]string, error) {
	panic("NYI")
}