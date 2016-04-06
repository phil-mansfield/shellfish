package cmd

type ShellConfig struct {

}

var _ Mode = &ShellConfig{}

func (config *ShellConfig) ExampleConfig() string {
	panic("NYI")
}

func (config *ShellConfig) ReadConfig(fname string) error {
	panic("NYI")
}

func (config *ShellConfig) validate() error {
	panic("NYI")
}

func (config *ShellConfig) Run(
	flags []string, gConfig *GlobalConfig, stdin []string,
) ([]string, error) {
	panic("NYI")
}