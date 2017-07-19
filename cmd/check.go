package cmd

import (
	"fmt"
	"math"
	"os"

	"github.com/phil-mansfield/shellfish/io"
	"github.com/phil-mansfield/shellfish/parse"
	"github.com/phil-mansfield/shellfish/cmd/env"
	"github.com/phil-mansfield/shellfish/cmd/memo"
)


type CheckConfig struct {
	h0, omegaM, omegaL float64
	boxWidth float64
	particleMasses []float64
	particleCount int64
}

var _ Mode = &CheckConfig{}

func (config *CheckConfig) ExampleConfig() string {
	return `[check.config]

# All fields are optional. Report float values to at least two decimal places.

# H0 = 70.0
# OmegaM = 0.27
# OmegaL = 0.73
# BoxWidth = 125.0
# ParticleMasses = 1.7e7, 1.8e8, 1.1e9
# ParticleCount = 1073741824
`
}

func (config *CheckConfig) ReadConfig(fname string, flags []string) error {

	vars := parse.NewConfigVars("check.config")

	vars.Float(&config.h0, "H0", -1)
	vars.Float(&config.omegaM, "OmegaM", -1)
	vars.Float(&config.omegaL, "OmegaL", -1)
	vars.Float(&config.boxWidth, "BoxWidth", -1)
	vars.Floats(&config.particleMasses, "ParticleMasses", []float64{})
	vars.Int(&config.particleCount, "ParticleCount", -1)

	if fname == "" {
		if len(flags) == 0 { return nil }
		if err := parse.ReadFlags(flags, vars); err != nil { return err }
	} else {
		if err := parse.ReadConfig(fname, vars); err != nil { return err }
		if err := parse.ReadFlags(flags, vars); err != nil { return err }
	}

	return nil
}

func (config *CheckConfig) Run(
	gConfig *GlobalConfig, e *env.Environment, stdin []byte,
) ([]string, error) {

	failedTests := []string{}

	buf, err := getVectorBuffer(
		e.ParticleCatalog(int(gConfig.HSnapMax), 0),
		gConfig.SnapshotType, gConfig.Endianness,
		gConfig.GadgetNpartNum,
	)

	hds, fnames, err := memo.ReadHeaders(int(gConfig.HSnapMax), buf, e)
	if err != nil { return nil, err }
	hd, fname := hds[0], fnames[0]

	failedTests = headerChecks(hd, config, failedTests)
	failedTests, err = particleChecks(buf, fname, config, failedTests)
	if err != nil { return nil, err }
	failedTests = haloChecks(hd, buf, config, failedTests)

	if len(failedTests) > 0 {
		if len(failedTests) == 1 {
			fmt.Println("Sanity check failed:")
		} else {
			fmt.Println("Sanity checks failed:")
		}

		for _, test := range failedTests {
			 fmt.Println(test)
		}
		os.Exit(1)
	}

	return nil, nil
}

func checkAlmostEq(x, y float64) bool {
	delta := y / 10
	return math.Abs(x - y) < delta
}

func headerChecks(
	hd io.Header, config *CheckConfig, failedTests []string,
) []string {
	if config.h0 > 0 && !checkAlmostEq(config.h0, hd.Cosmo.H100 * 100) {
		msg := fmt.Sprintf(
			"H0 value in check.config is %g, but read H0 value is %g.",
			config.h0, hd.Cosmo.H100 * 100,
		)
		failedTests = append(failedTests, msg)
	}

	if config.omegaM > 0 && !checkAlmostEq(config.omegaM, hd.Cosmo.OmegaM) {
		msg := fmt.Sprintf(
			"OmegaM value in check.config is %g, but read OmegaM value is %g.",
			config.omegaM, hd.Cosmo.OmegaM,
		)
		failedTests = append(failedTests, msg)
	}

	if config.omegaL > 0 && !checkAlmostEq(config.omegaL, hd.Cosmo.OmegaL) {
		msg := fmt.Sprintf(
			"OmegaL value in check.config is %g, but read OmegaL value is %g.",
			config.omegaL, hd.Cosmo.OmegaL,
		)
		failedTests = append(failedTests, msg)
	}

	if config.boxWidth > 0 && !checkAlmostEq(config.boxWidth, hd.TotalWidth) {
		msg := fmt.Sprintf(
			"BoxWidth value in check.config is %g, but read " +
			"BoxWidth value is %g.",  config.boxWidth, hd.TotalWidth,
		)
		failedTests = append(failedTests, msg)
	}

	return failedTests
}

func particleChecks(
	buf io.VectorBuffer, fname string,
	config *CheckConfig, failedTests []string,
) ([]string, error) {

	if len(config.particleMasses) > 0 {

		_, _, ms, _, err := buf.Read(fname)
		if err != nil {
			return failedTests, err
		}

		failureMass := -1.0

		for _, m := range ms {
			found := false
			for _, mm := range config.particleMasses {
				found = found || checkAlmostEq(float64(m), mm)
			}

			if !found {
				failureMass = float64(m)
				break
			}
		}

		if failureMass != -1 {
			msg := fmt.Sprintf(
				"Allowed masses in check.config are %g, but a particle " +
				"with mass %g was found in a paritcle snapshot.",
				config.particleMasses, failureMass,
			)

			failedTests = append(failedTests, msg)
		}
	}

	if config.particleCount > 0 {
		n, err := buf.TotalParticles(fname)
		if err != nil { return failedTests, err }
		if int64(n) != config.particleCount {
			msg := fmt.Sprintf(
				"ParticleCount value in check.config is %d, but read value " +
				"is %d.", config.particleCount, n,
			)
			failedTests = append(failedTests, msg)
		}
	}

	return failedTests, nil
}

func haloChecks(
	hd io.Header, buf io.VectorBuffer,
	config *CheckConfig, failedTests []string,
) []string {
	return failedTests
}