package io

import (
	"fmt"
	"strconv"

	artio "github.com/phil-mansfield/go-artio"
)

type ARTIOBuffer struct {
	open bool
	buf [][3]float32
	sBufs [][][3]float32
	fileset string
}

func NewARTIOBuffer(fileset string) (VectorBuffer, error) {
	h, err := artio.FilesetOpen(fileset, artio.OpenHeader, artio.NullContext)
	if err != nil { return nil, err  }
	defer h.Close()

	numSpecies := h.GetInt(h.Key("num_particle_species"))[0]

	return &ARTIOBuffer{ sBufs: make([][][3]float32, numSpecies) }, nil
}

func (buf *ARTIOBuffer) Read(fileNumStr string) ([][3]float32, error) {
	// Open the file.
	if buf.open { panic("Buffer already open.") }
	buf.open = true

	h, err := artio.FilesetOpen(
		buf.fileset, artio.OpenHeader, artio.NullContext,
	)
	if err != nil { return nil, err  }
	defer h.Close()

	// I'm not sure if this can just be replaced with putting an
	// artio.OpenParticles flag in artio.FilesetOpen(). Someone with more
	// knowledge about ARTIO than me should figure this out.
	err = h.OpenParticles()
	if err != nil { return nil, err}

	// Flag N_BODY particles.
	flags, err := nBodyFlags(h, buf.fileset)
	if err != nil { return nil, err }

	// Get SFC range.
	fIdx, err := strconv.Atoi(fileNumStr)
	fileIdxs := h.GetLong(h.Key("particle_file_sfc_index"))
	sfcStart, sfcEnd := fileIdxs[fIdx], fileIdxs[fIdx + 1] - 1

	// Counts and buffer manipulation. Do the reading.
	sCounts, err := h.CountInRange(sfcStart, sfcEnd)
	totCount := int64(0)
	for i := range sCounts {
		if flags[i] {
			totCount += sCounts[i]
			expandVectors(buf.sBufs[i], int(sCounts[i]))
			err = h.GetPositionsAt(i, sfcStart, sfcEnd, buf.sBufs[i])
			if err != nil { return nil, err }
		}
	}

	// Copy to output buffer.
	expandVectors(buf.buf, int(totCount))
	k := 0
	for j := range buf.sBufs {
		for i := range buf.sBufs[j] {
			buf.buf[i] = buf.sBufs[j][k]
			k++
		}
	}

	return buf.buf, nil
}

func nBodyFlags(h artio.Fileset, fname string) ([]bool, error) {
	speciesLabels := h.GetString(h.Key("particle_species_labels"))
	isNBody, nBodyCount := make([]bool, len(speciesLabels)), 0
	for i := range isNBody {
		isNBody[i] = speciesLabels[i] == "N-BODY"
		nBodyCount++
	}
	if nBodyCount == 0 {
		return nil, fmt.Errorf("ARTIO fileset '%s' does not contain any " +
		"particle species of type 'N-BODY'.", fname)
	}
	return isNBody, nil

}

func (buf *ARTIOBuffer) Close() {
	if !buf.open { panic("Buffer not open.") }
	buf.open = false
}

func (buf *ARTIOBuffer) IsOpen() bool {
	return buf.open
}

func (buf *ARTIOBuffer) ReadHeader(fileNumStr string, out *Header) error {
	xs, err := buf.Read(fileNumStr)

	h, err := artio.FilesetOpen(
		buf.fileset, artio.OpenHeader, artio.NullContext,
	)
	if err != nil { return err }
	defer h.Close()

	out.TotalWidth = h.GetDouble(h.Key("box_size"))[0]
	out.Origin, out.Width = boundingBox(xs, out.TotalWidth)
	out.N = int64(len(xs))
	out.Count = -1

	// I get the cosmology afterwards to aid in debugging.
	switch {
	case !h.HasKey("auni"):
		return fmt.Errorf("ARTIO header does not contain 'auni' field.")
	case !h.HasKey("OmegaM"):
		return fmt.Errorf("ARTIO header does not contain 'OmegaM' field.")
	case !h.HasKey("OmegaL"):
		return fmt.Errorf("ARTIO header does not contain 'OmegaL' field.")
	case !h.HasKey("hubble"):
		return fmt.Errorf("ARTIO header does not contain 'hubble' field.")
	}

	out.Cosmo.Z = 1/h.GetDouble(h.Key("auni"))[0] - 1
	out.Cosmo.OmegaM = h.GetDouble(h.Key("OmegaM"))[0]
	out.Cosmo.OmegaL = h.GetDouble(h.Key("OmegaL"))[0]
	out.Cosmo.H100 = h.GetDouble(h.Key("hubble"))[0]

	return nil
}