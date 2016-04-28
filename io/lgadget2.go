package io

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
)

// CatalogHeader describes meta-information about the current catalog.
type CatalogHeader struct {
	Cosmo CosmologyHeader

	Mass       float64 // Mass of one particle
	Count      int64   // Number of particles in catalog
	TotalCount int64   // Number of particles in all catalogs
	CountWidth int64   // Number of particles "on one side": TotalCount^(1/3)

	Idx        int64   // Index of catalog: x-major ordering is used
	GridWidth  int64   // Number of gird cells "on one side"
	Width      float64 // Width of the catalog's bounding box
	TotalWidth float64 // Width of the sim's bounding box
}


// gadgetHeader is the formatting for meta-information used by Gadget 2.
type lGadget2Header struct {
	NPart                                     [6]uint32
	Mass                                      [6]float64
	Time, Redshift                            float64
	FlagSfr, FlagFeedback                     int32
	NPartTotal                                [6]uint32
	FlagCooling, NumFiles                     int32
	BoxSize, Omega0, OmegaLambda, HubbleParam float64
	FlagStellarAge, HashTabSize               int32

	Padding [88]byte
}

// Standardize returns a Header that corresponds to the source
// Gadget 2 header.
func (gh *lGadget2Header) Standardize() *CatalogHeader {
	h := &CatalogHeader{}

	h.Count = int64(gh.NPart[1] + gh.NPart[0]<<32)
	h.TotalCount = int64(gh.NPartTotal[1] + gh.NPartTotal[0]<<32)
	h.Mass = float64(gh.Mass[1])
	h.TotalWidth = float64(gh.BoxSize)
	h.Width = -1.0

	h.Cosmo.Z = gh.Redshift
	h.Cosmo.OmegaM = gh.Omega0
	h.Cosmo.OmegaL = gh.OmegaLambda
	h.Cosmo.H100 = gh.HubbleParam

	return h
}

func (gh *lGadget2Header) postprocess(xs [][3]float32, out *Header) {
	// Assumes the catalog has already been checked for corruption.

	out.N = int64(gh.NPart[1] + gh.NPart[0]<<32)
	out.Count = int64(gh.NPartTotal[1] + gh.NPartTotal[0]<<32)
	out.TotalWidth = gh.BoxSize

	out.Cosmo.Z = gh.Redshift
	out.Cosmo.OmegaM = gh.Omega0
	out.Cosmo.OmegaL = gh.OmegaLambda
	out.Cosmo.H100 = gh.HubbleParam

	out.Origin, out.Width = boundingBox(xs, gh.BoxSize)
}

// readInt32 returns single 32-bit interger from the given file using the
// given endianness.
func readInt32(r io.Reader, order binary.ByteOrder) int32 {
	var n int32
	if err := binary.Read(r, order, &n); err != nil {
		panic(err)
	}
	return n
}

// ReadGadgetHeader reads a Gadget catalog and returns a standardized
// gotetra containing its information.
func readLGadget2Header(
	path string, order binary.ByteOrder, out *lGadget2Header,
) error {
	f, err := os.Open(path)
	if err != nil { return err }
	defer f.Close()

	_ = readInt32(f, order)
	err = binary.Read(f, binary.LittleEndian, out)
	return err
}

// ReadGadgetParticlesAt reads a Gadget file and writes all the particles within
// it to the given particle buffer, ps. floatBuf and intBuf are used internally.
// The length of all three buffers must be equal to  the number of particles in
// the catalog.
func readLGadget2Positions(
	path string,
	order binary.ByteOrder,
	buf [][3]float32,
) (out [][3]float32, err error) {
	f, err := os.Open(path)
	if err != nil { return nil, err }
	defer f.Close()

	gh := &lGadget2Header{}

	_ = readInt32(f, order)
	binary.Read(f, binary.LittleEndian, gh)
	_ = readInt32(f, order)
	count := int(gh.NPartTotal[1] + gh.NPartTotal[0]<<32)
	out = expandVectors(out[:0], count)

	_ = readInt32(f, order)
	readVecAsByte(f, order, out)
	//_ = readInt32(f, order)

	tw := float32(gh.BoxSize)
	for i := range out {
		for j := 0; j < 3; j++ {
			if out[i][j] < 0 {
				out[i][j] += tw
			} else if out[i][j] >= tw {
				out[i][j] -= tw
			}

			if math.IsNaN(float64(out[i][j])) ||
				math.IsInf(float64(out[i][j]), 0) ||
				out[i][j] < -tw || out[i][j] > 2*tw {

				return nil, fmt.Errorf(
					"Corruption detected in the file %s. I can't analyze it.",
					path,
				)
			}
		}
	}

	return out, nil
}

func expandVectors(vecs [][3]float32, n int) [][3]float32 {
	switch {
	case cap(vecs) <= n:
		return vecs[:n]
	case int(float64(cap(vecs)) * 1.5) > n:
		return append(vecs[:cap(vecs)],
			make([][3]float32, n - cap(vecs))...)
	default:
		return make([][3]float32, n)
	}
}

type LGadget2Buffer struct {
	open bool
	order binary.ByteOrder
	hd lGadget2Header
	xs [][3]float32
}

func NewLGadget2Buffer(orderFlag string) VectorBuffer {
	var order binary.ByteOrder = binary.LittleEndian
	switch orderFlag {
	case "LittleEndian":
	case "BigEndian":
		order = binary.BigEndian
	case "SystemOrder":
		if !IsSysOrder(order) { order = binary.BigEndian }
	}
	return &LGadget2Buffer{ order: order }
}

func (buf *LGadget2Buffer) Read(fname string) ([][3]float32, error) {
	if buf.open { panic("Buffer already open.") }
	buf.open = true

	var err error
	buf.xs, err = readLGadget2Positions(fname, buf.order, buf.xs)
	return buf.xs, err
}

func (buf *LGadget2Buffer) Close() {
	if !buf.open { panic("Buffer not open.") }
	buf.open = false
}

func (buf *LGadget2Buffer) IsOpen() bool {
	return buf.open
}

func (buf *LGadget2Buffer) ReadHeader(fname string, out *Header) error {

	err := readLGadget2Header(fname, buf.order, &buf.hd)
	if err != nil { return err }
	_, err = buf.Read(fname)
	if err != nil { return err }

	buf.hd.postprocess(buf.xs, out)

	return nil
}