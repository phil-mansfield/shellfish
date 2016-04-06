package cmd

type IDConfig struct {

}

var _ Mode = &IDConfig{}

func (config *IDConfig) ExampleConfig() string {
	panic("NYI")
}

func (config *IDConfig) ReadConfig(fname string) error {
	panic("NYI")
}

func (config *IDConfig) validate() error {
	panic("NYI")
}

func (config *IDConfig) Run(
	flags []string, gConfig *GlobalConfig, stdin []string,
) ([]string, error) {
	panic("NYI")
}