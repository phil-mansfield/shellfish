package cmd

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestExampleFiles(t *testing.T) {
	tests := []Mode {
		&GlobalConfig{},
		&IDConfig{},
	}

	for i := range tests {
		mode := tests[i]
		f, err := ioutil.TempFile("", "shellfish_config_test")
		if err != nil { panic(err.Error()) }
		defer os.Remove(f.Name())

		if _, err = f.Write([]byte(mode.ExampleConfig())); err != nil {
			panic(err.Error())
		}
		if err = f.Close(); err != nil {
			panic(err.Error())
		}

		err = mode.ReadConfig(f.Name())
		if err != nil {
			t.Errorf("%d) Got error when parsing config file:\n%s",
				i, err.Error())
		}
	}
}