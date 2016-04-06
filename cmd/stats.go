package cmd

type StatsConfig struct {

}

var _ Mode = &StatsConfig{}

func (config *StatsConfig) ExampleConfig() string {
	panic("NYI")
}

func (config *StatsConfig) ReadConfig(fname string) error {
	panic("NYI")
}

func (config *StatsConfig) validate() error {
	panic("NYI")
}

func (config *StatsConfig) Run(
	flags []string, gConfig *GlobalConfig, stdin []string,
) ([]string, error) {
	panic("NYI")
}