package io

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	artio "github.com/phil-mansfield/go-artio"
	"github.com/phil-mansfield/shellfish/cosmo"
)

const (
	// emulateHubble is used for debugging purposes. I've never had access to
	// a cosmological simulation, so this is necessary. Don't worry: even if
	// this flag is set, an error will still be returned if called on invalid
	// header contents. It will just occur late enough to allow for illustrative
	// logging.
	emulateHubble = true
)

type ARTIOBuffer struct {
	open                bool
	xsBuf               [][3]float32
	msBuf               []float32
	xsBufs              [][][3]float32
	msBufs              [][]float32
	idsBuf              []int64
	sMasses             []float32
	sFlags              []bool // True if the species is "N-BODY" type.
	fileset, exFilename string
}

func NewARTIOBuffer(filename string) (VectorBuffer, error) {
	fileset, _, err := parseARTIOFilename(filename)
	if err != nil {
		return nil, err
	}

	h, err := artio.FilesetOpen(fileset, artio.OpenHeader, artio.NullContext)
	if err != nil {
		return nil, err
	}
	defer h.Close()

	numSpecies := h.GetInt(h.Key("num_particle_species"))[0]
	sMasses := h.GetFloat(h.Key("particle_species_mass"))

	var h100 float64
	if !h.HasKey("hubble") {
		if emulateHubble {
			h100 = 0.7
		} else {
			return nil, fmt.Errorf(
				"ARTIO header does not contain 'hubble' field.",
			)
		}
	} else {
		h100 = h.GetDouble(h.Key("hubble"))[0]
	}
	massUnit := (h100 / (cosmo.MSunMks * 1000)) *
		h.GetDouble(h.Key("mass_unit"))[0]
	for i := range sMasses {
		sMasses[i] *= float32(massUnit)
	}

	sFlags, err := nBodyFlags(h, fileset)
	if err != nil {
		return nil, err
	}

	return &ARTIOBuffer{
		xsBufs:     make([][][3]float32, numSpecies),
		msBufs:     make([][]float32, numSpecies),
		sMasses:    sMasses,
		sFlags:     sFlags,
		fileset:    fileset,
		exFilename: filename,
	}, nil
}

func parseARTIOFilename(fname string) (fileset string, block int, err error) {
	split := strings.LastIndex(fname, ".")
	if split == -1 || split == len(fname)-1 {
		return "", -1, fmt.Errorf(
			"'%s' is not the name of an ARTIO block.", fname,
		)
	}

	fileset, blockString := fname[:split], fname[split+1:]
	block, err = strconv.Atoi(strings.Trim(blockString, "p"))
	if err != nil {
		return "", -1, fmt.Errorf(
			"'%s' is not the name of an ARTIO block.", fname,
		)
	}

	return fileset, block, nil
}

func (buf *ARTIOBuffer) Read(
	filename string,
) (xs, vs[][3]float32, ms []float32, ids []int64, err error) {
	// Open the file.
	if buf.open {
		panic("Buffer already open.")
	}
	buf.open = true

	h, err := artio.FilesetOpen(
		buf.fileset, artio.OpenHeader, artio.NullContext,
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer h.Close()

	// I'm not sure if this can just be replaced with putting an
	// artio.OpenParticles flag in artio.FilesetOpen(). Someone with more
	// knowledge about ARTIO than me should figure this out.
	err = h.OpenParticles()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Flag N_BODY particles.
	flags, err := nBodyFlags(h, buf.fileset)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Get SFC range.
	_, fIdx, err := parseARTIOFilename(filename)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	fileIdxs := h.GetLong(h.Key("particle_file_sfc_index"))
	sfcStart, sfcEnd := fileIdxs[fIdx], fileIdxs[fIdx+1]-1

	// Counts and buffer manipulation. Do the reading.
	sCounts, err := h.CountInRange(sfcStart, sfcEnd)
	totCount := int64(0)

	for i := range sCounts {
		if flags[i] {
			totCount += sCounts[i]
			buf.xsBufs[i] = expandVectors(buf.xsBufs[i][:0], int(sCounts[i]))
			err = h.GetPositionsAt(i, sfcStart, sfcEnd, buf.xsBufs[i])
			if err != nil {
				return nil, nil, nil, nil, err
			}

			buf.msBufs[i] = expandScalars(buf.msBufs[i][:0], int(sCounts[i]))
			for j := range buf.msBufs[i] {
				buf.msBufs[i][j] = buf.sMasses[i]
			}
		}
	}

	// Copy to output buffer.
	buf.xsBuf = expandVectors(buf.xsBuf[:0], int(totCount))
	buf.msBuf = expandScalars(buf.msBuf[:0], int(totCount))
	k := 0
	for j := range buf.xsBufs {
		for i := range buf.xsBufs[j] {
			buf.xsBuf[k] = buf.xsBufs[j][i]
			buf.msBuf[k] = buf.msBufs[j][i]
			k++
		}
	}

	var h100 float32
	if !h.HasKey("hubble") {
		if emulateHubble {
			h100 = 0.7
		} else {
			return nil, nil, nil, nil, fmt.Errorf(
				"ARTIO header does not contain 'hubble' field.",
			)
		}
	} else {
		h100 = float32(h.GetDouble(h.Key("hubble"))[0])
	}

	lengthUnit := float32(h100) / (cosmo.MpcMks * 100) *
		float32(h.GetDouble(h.Key("length_unit"))[0])
	for i := range buf.xsBuf {
		buf.xsBuf[i][0] *= lengthUnit
		buf.xsBuf[i][1] *= lengthUnit
		buf.xsBuf[i][2] *= lengthUnit
	}

	return buf.xsBuf, nil, buf.msBuf, buf.idsBuf, nil
}

func nBodyFlags(h artio.Fileset, fname string) ([]bool, error) {
	speciesLabels := h.GetString(h.Key("particle_species_labels"))
	isNBody, nBodyCount := make([]bool, len(speciesLabels)), 0
	for i := range isNBody {
		isNBody[i] = speciesLabels[i] == "N-BODY"
		nBodyCount++
	}
	if nBodyCount == 0 {
		return nil, fmt.Errorf("ARTIO fileset '%s' does not contain any "+
			"particle species of type 'N-BODY'.", fname)
	}
	return isNBody, nil
}

func (buf *ARTIOBuffer) Close() {
	if !buf.open {
		panic("Buffer not open.")
	}
	buf.open = false
}

func (buf *ARTIOBuffer) IsOpen() bool {
	return buf.open
}

func (buf *ARTIOBuffer) ReadHeader(fileNumStr string, out *Header) error {
	xs, _, _, _, err := buf.Read(fileNumStr)
	defer buf.Close()

	h, err := artio.FilesetOpen(
		buf.fileset, artio.OpenHeader, artio.NullContext,
	)
	if err != nil {
		return err
	}
	defer h.Close()

	var h100 float64
	if !h.HasKey("hubble") {
		if emulateHubble {
			h100 = 0.7
		} else {
			return fmt.Errorf(
				"ARTIO header does not contain 'hubble' field.",
			)
		}
	} else {
		h100 = h.GetDouble(h.Key("hubble"))[0]
	}

	lengthUnit := h100 / (cosmo.MpcMks * 100) *
		h.GetDouble(h.Key("length_unit"))[0]
	out.TotalWidth = h.GetDouble(h.Key("box_size"))[0] * lengthUnit

	if out.TotalWidth < 0.001 || out.TotalWidth > 1e6 {
		return fmt.Errorf("ARTIO box_size calculated to be %g Mpc/h. This "+
			"is an internal bug: please submit an issue.", out.TotalWidth)
	}

	out.Origin, out.Width = boundingBox(xs, out.TotalWidth)
	out.N = int64(len(xs))

	min, max := xs[0], xs[0]
	for i := range xs {
		for j := 0; j < 3; j++ {
			if xs[i][j] < min[j] {
				min[j] = xs[i][j]
			}
			if xs[i][j] > max[j] {
				max[j] = xs[i][j]
			}
		}
	}

	switch {
	case !h.HasKey("auni"):
		return fmt.Errorf("ARTIO header does not contain 'auni' field.")
	case !h.HasKey("OmegaM"):
		return fmt.Errorf("ARTIO header does not contain 'OmegaM' field.")
	case !h.HasKey("OmegaL"):
		return fmt.Errorf("ARTIO header does not contain 'OmegaL' field.")

	}

	out.Cosmo.Z = 1/h.GetDouble(h.Key("auni"))[0] - 1
	out.Cosmo.OmegaM = h.GetDouble(h.Key("OmegaM"))[0]
	out.Cosmo.OmegaL = h.GetDouble(h.Key("OmegaL"))[0]
	out.Cosmo.H100 = h.GetDouble(h.Key("hubble"))[0]

	if out.Cosmo.H100 > 10 {
		panic("Oops, Phil misunderstood the meaning of an ARTIO field. " +
			"Please submit an issue.")
	}

	return nil
}

func (buf *ARTIOBuffer) MinMass() float32 {
	minMass := float32(math.Inf(+1))
	for i := range buf.sMasses {
		if buf.sFlags[i] && buf.sMasses[i] < minMass {
			minMass = buf.sMasses[i]
		}
	}
	return minMass
}

func (buf *ARTIOBuffer) TotalParticles(fname string) (int, error) {
	return -1, nil
}