package cmd

import (
	"fmt"
	"math"
	"os"
	"log"

	"github.com/phil-mansfield/shellfish/io"
	"github.com/phil-mansfield/shellfish/parse"
	"github.com/phil-mansfield/shellfish/cmd/env"
	"github.com/phil-mansfield/shellfish/cmd/memo"
	"github.com/phil-mansfield/shellfish/los/geom"
)


type CheckConfig struct {
	h0, omegaM, omegaL float64
	boxWidth float64
	particleMasses []float64
	particleCount int64
	exampleHalo []float64
}

var _ Mode = &CheckConfig{}

func (config *CheckConfig) ExampleConfig() string {
	return `[check.config]

# All fields are optional. Unless otherwise stated, checks are very rough and
# only done to 10%, so  report float values to at least two decimal places.

# H0 = 70
# OmegaM = 0.30
# OmegaL = 0.70

# BoxWidth is in units of Mpc/h (not your code units).
# BoxWidth = 125

# The total number of particles in the simulation box. This is checked exactly.
# ParticleCount = 1073741824

# ParticleMasses A list of all the valid DM particle masses in Msun/h (not your
# code units).
# ParticleMasses = 1.7e7, 1.8e8, 1.1e9

# ExampleHalo is a the location and size of one halo from your catalog. It
# is a list of six values:
#    ExampleHalo[0] = snapshot number
#    ExampleHalo[1 - 3] = (X, Y, Z) [Mpc/h]
#    ExampleHalo[4] = Radius [Mpc/h]
#    ExampleHalo[5] = Unbound mass contained within radius [Msun/h]
# The radius used for elements 4 and 5 can correspond to any definition or
# no definition at all, but be careful: most halo catalogs report *bound*,
# not *unbound* masses.
# Report position and radius to the highest accuracy that you know.
# ExampleHalo = 100, 4.68299, 100.552, 80.9536, 2.68893, 1.446e+15
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
	vars.Floats(&config.exampleHalo, "ExampleHalo", []float64{})

	if fname == "" {
		if len(flags) == 0 { return nil }
		if err := parse.ReadFlags(flags, vars); err != nil { return err }
	} else {
		if err := parse.ReadConfig(fname, vars); err != nil { return err }
		if err := parse.ReadFlags(flags, vars); err != nil { return err }
	}

	if len(config.exampleHalo) != 0 && len(config.exampleHalo) != 6 {
		return fmt.Errorf(
			"ExampleHalo field must have 6 entries, not %d",
			len(config.exampleHalo),
		)
	}

	return nil
}

func (config *CheckConfig) Run(
	gConfig *GlobalConfig, e *env.Environment, stdin []byte,
) ([]string, error) {

	failedTests := []string{}

	buf, err := getVectorBuffer(
		e.ParticleCatalog(int(gConfig.HSnapMax), 0), gConfig,
	)

	hds, fnames, err := memo.ReadHeaders(int(gConfig.HSnapMax), buf, e)
	if err != nil { return nil, err }
	hd, fname := hds[0], fnames[0]

	failedTests = headerChecks(hd, config, failedTests)
	failedTests, err = particleChecks(buf, fname, config, failedTests)
	if err != nil { return nil, err }
	failedTests, err = haloChecks(hd, buf, config, failedTests, e)
	if err != nil { return nil, err }

	log.Printf("Mass in sphere: %.6g", massContainedMass)
	log.Printf("Particle count in sphere: %d", massContainedCount)
	log.Printf("Average particle mass in sphere: %.6g",
		massContainedMass/float64(massContainedCount))

	log.Printf("Total mass considered by kernel: %.6g", totalMass)
	log.Printf("Particle count considered by kernel: %d", totalCount)
	log.Printf("Average particle mass in kernel: %.6g",
		totalMass/float64(totalCount))

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

		buf.Close()
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
	e *env.Environment,
) ([]string, error) {
	if len(config.exampleHalo) == 0 { return failedTests, nil }

	snap := int(config.exampleHalo[0])
	hx,hy,hz:=config.exampleHalo[1],config.exampleHalo[2],config.exampleHalo[3]
	hr, hm := config.exampleHalo[4], config.exampleHalo[5]

	snapCoords := [][]float64{
		[]float64{hx}, []float64{hy}, []float64{hz}, []float64{hr},
	}
	s := geom.Sphere{
		C: [3]float32{float32(hx), float32(hy), float32(hz)},
		R: float32(hr),
	}

	/* The following code is identical to the prof mainloop. */

	hds, files, err := memo.ReadHeaders(snap, buf, e)
	if err != nil {
		return nil, err
	}
	hBounds, err := boundingSpheres(snapCoords, &hds[0], e)
	if err != nil {
		return nil, err
	}
	_, intrIdxs := binSphereIntersections(hds, hBounds)

	haloMass := 0.0

	for i := range hds {
		if len(intrIdxs[i]) == 0 {
			continue
		}

		xs, _, ms, _, err := buf.Read(files[i])
		if err != nil {
			return nil, err
		}

		haloMass += addMass(s, &hds[i], xs, ms)

		buf.Close()
	}

	if !checkAlmostEq(hm, haloMass) {
		msg := fmt.Sprintf(
			"ExampleHalo mass in check.config is %g, but measured " +
			"halo mass is %g.", hm, haloMass,
		)
		failedTests = append(failedTests, msg)
	}


	return failedTests, nil
}

func addMass(
	s geom.Sphere, hd *io.Header, xs [][3]float32, ms []float32,
) float64 {
	rMax2 := s.R*s.R

	x0, y0, z0 := s.C[0], s.C[1], s.C[2]
	tw2 := float32(hd.TotalWidth) / 2

	m := 0.0
	
	for i, vec := range xs {
		x, y, z := vec[0], vec[1], vec[2]
		dx, dy, dz := x - x0, y - y0, z - z0
		dx = wrap(dx, tw2)
		dy = wrap(dy, tw2)
		dz = wrap(dz, tw2)

		r2 := dx*dx + dy*dy + dz*dz
		if  r2 <= rMax2 {
			m += float64(ms[i])
			massContainedCount++
			massContainedMass += float64(ms[i])
		}

		totalMass += float64(ms[i])
		totalCount++

	}
	
	return m
}
