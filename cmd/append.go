package cmd

type AppendConfig struct {

}

var _ Mode = &AppendConfig{}

func (config *AppendConfig) ExampleConfig() string {
	panic("NYI")
}

func (config *AppendConfig) ReadConfig(fname string) error {
	panic("NYI")
}

func (config *AppendConfig) validate() error {
	panic("NYI")
}

func (config *AppendConfig) Run(
	flags []string, gConfig *GlobalConfig, stdin []string,
) ([]string, error) {
	panic("NYI")
}