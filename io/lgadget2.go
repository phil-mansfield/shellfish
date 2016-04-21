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

// CosmologyHeader contains information describing the cosmological
// context in which the simulation was run.
type CosmologyHeader struct {
	Z      float64
	OmegaM float64
	OmegaL float64
	H100   float64
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

// readInt32 returns single 32-bit interger from the given file using the
// given endianness.
func readInt32(r io.Reader, order binary.ByteOrder) int32 {
	var n int32
	if err := binary.Read(r, order, &n); err != nil {
		panic(err)
	}
	return n
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

// WrapDistance takes a value and interprets it as a position defined within
// a periodic domain of width h.BoxSize.
func (h *lGadget2Header) WrapDistance(x float64) float64 {
	if x < 0 {
		return x + h.BoxSize
	} else if x >= h.BoxSize {
		return x - h.BoxSize
	}
	return x
}

// ReadGadgetHeader reads a Gadget catalog and returns a standardized
// gotetra containing its information.
func ReadGadgetHeader(path string, order binary.ByteOrder) *CatalogHeader {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	gh := &lGadget2Header{}

	_ = readInt32(f, order)
	binary.Read(f, binary.LittleEndian, gh)
	h := gh.Standardize()

	return h
}

// ReadGadgetParticlesAt reads a Gadget file and writes all the particles within
// it to the given particle buffer, ps. floatBuf and intBuf are used internally.
// The length of all three buffers must be equal to  the number of particles in
// the catalog.
//
// This call signature, and espeically the Particle type are all a consequence
// of soem shockingly poor early design decisions.
func ReadGadgetParticlesAt(
path string,
order binary.ByteOrder,
xs, vs [][3]float32,
ids []int64,
) {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	gh := &lGadget2Header{}

	_ = readInt32(f, order)
	binary.Read(f, binary.LittleEndian, gh)
	_ = readInt32(f, order)

	h := gh.Standardize()

	if int64(len(xs)) != h.Count {
		panic(fmt.Sprintf(
			"Incorrect length for xs buffer. Found %d, expected %d",
			len(xs), h.Count,
		))
	} else if int64(len(vs)) != h.Count {
		panic(fmt.Sprintf(
			"Incorrect length for vs buffer. Found %d, expected %d",
			len(vs), h.Count,
		))
	} else if int64(len(ids)) != h.Count {
		panic(fmt.Sprintf(
			"Incorrect length for int buffer. Found %d, expected %d",
			len(ids), h.Count,
		))
	}

	_ = readInt32(f, order)
	readVecAsByte(f, order, xs)
	_ = readInt32(f, order)
	_ = readInt32(f, order)
	readVecAsByte(f, order, vs)
	_ = readInt32(f, order)
	_ = readInt32(f, order)
	readInt64AsByte(f, order, ids)
	//_ = readInt32(f, order)

	fmt.Printf("%.3g\n", xs[:20])
	fmt.Printf("%.3g\n", vs[:20])
	fmt.Printf("%d\n", ids[:20])
	fmt.Printf("%x\n", ids[:20])

	rootA := float32(math.Sqrt(float64(gh.Time)))
	for i := range xs {
		for j := 0; j < 3; j++ {
			xs[i][j] = float32(gh.WrapDistance(float64(xs[i][j])))
			vs[i][j] = vs[i][j] * rootA
		}
	}
}

// ReadGadget reads the gadget particle catalog located at the given location
// and written with the given endianness. Its header and particle sequence
// are returned in a standardized format.
func ReadGadget(
path string, order binary.ByteOrder,
) (hd *CatalogHeader, xs, vs [][3]float32, ids []int64) {

	hd = ReadGadgetHeader(path, order)
	xs = make([][3]float32,  hd.Count)
	vs = make([][3]float32,  hd.Count)
	ids = make([]int64, hd.Count)

	ReadGadgetParticlesAt(path, order, xs, vs, ids)
	return hd, xs, vs, ids
}
