package io

import (
	"encoding/binary"
	"fmt"
	"os"
	
	"github.com/phil-mansfield/shellfish/cosmo"

	"unsafe"
)

type GotetraBuffer struct {
	open   bool
	sheet  [][3]float32
	xs     [][3]float32
	ms     []float32
	ids    []int64
	sw, gw int
	mass   float32
	hd     gotetraHeader
}

func NewGotetraBuffer(fname string) (VectorBuffer, error) {
	hd := &gotetraHeader{}
	f, _, err := loadSheetHeader(fname, hd)
	if err != nil {
		return nil, err
	}
	err = f.Close()
	if err != nil {
		return nil, err
	}

	sw, gw := hd.SegmentWidth, hd.GridWidth
	buf := &GotetraBuffer{
		sheet: make([][3]float32, gw*gw*gw),
		xs:    make([][3]float32, sw*sw*sw),
		ms:    make([]float32, sw*sw*sw),
		ids:   make([]int64, sw*sw*sw),
		open:  false,
		sw:    int(sw), gw: int(gw),
		mass:  calcUniformMass(hd.Count, hd.TotalWidth, hd.Cosmo),
	}

	return buf, nil
}

// Returned units are Msun/h.
func calcUniformMass(count int64, tw float64, c CosmologyHeader) float32 {
	rhoM0 := cosmo.RhoAverage(c.H100*100, c.OmegaM, c.OmegaL, 0)
	mTot := (tw * tw * tw) * rhoM0
	return float32(mTot / float64(count))
}

func (buf *GotetraBuffer) MinMass() float32 { return buf.mass }

func (buf *GotetraBuffer) IsOpen() bool { return buf.open }

func (buf *GotetraBuffer) Read(fname string) (
	xs, vs [][3]float32, ms []float32, ids []int64, err error,
) {
	if buf.open {
		panic("Buffer already open.")
	}
	buf.open = true

	err = readSheetPositionsAt(fname, buf.sheet)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	si := 0
	for z := 0; z < buf.sw; z++ {
		for y := 0; y < buf.sw; y++ {
			for x := 0; x < buf.sw; x++ {
				// if it ever matters, this calculation can be sped up
				// significantly.
				gi := x + y*buf.gw + z*buf.gw*buf.gw
				buf.xs[si] = buf.sheet[gi]
				buf.ms[si] = buf.mass
				si++
			}
		}
	}

	return buf.xs, nil, buf.ms, buf.ids, nil
}

func (buf *GotetraBuffer) Close() {
	if !buf.open {
		panic("Buffer already closed.")
	}

	buf.open = false
}

/*
The binary format used for phase sheets is as follows:
    |-- 1 --||-- 2 --||-- ... 3 ... --||-- ... 4 ... --||-- ... 5 ... --|

    1 - (int32) Flag indicating the endianness of the file. 0 indicates a big
        endian byte ordering and -1 indicates a little endian byte order.
    2 - (int32) Size of a Header struct. Should be checked for consistency.
    3 - (sheet.Header) Header file containing meta-information about the
        sheet fragment.
    4 - ([][3]float32) Contiguous block of x, y, z coordinates. Given in Mpc.
*/
type rawGotetraHeader struct {
	Cosmo                              CosmologyHeader
	Count, CountWidth                  int64
	SegmentWidth, GridWidth, GridCount int64
	Idx, Cells                         int64

	Mass       float64
	TotalWidth float64

	Origin, Width                 [3]float32
	VelocityOrigin, VelocityWidth [3]float32
}

func (raw *rawGotetraHeader) postprocess(hd *Header) {
	hd.Cosmo = raw.Cosmo

	hd.N = raw.SegmentWidth * raw.SegmentWidth * raw.SegmentWidth
	hd.TotalWidth = raw.TotalWidth
	hd.Origin, hd.Width = raw.Origin, raw.Width
}

func (raw *rawGotetraHeader) fileCoords() (x, y, z int) {
	xx := raw.Idx % raw.Cells
	yy := (raw.Idx / raw.Cells) % raw.Cells
	zz := raw.Idx / (raw.Cells * raw.Cells)
	return int(xx), int(yy), int(zz)
}

type gotetraHeader struct {
	rawGotetraHeader
	N     int64
	guard struct{} // Prevents accidentally trying to write/read this type.
}

// endianness is a utility function converting an endianness flag to a
// byte order.
func endianness(flag int32) binary.ByteOrder {
	if flag == 0 {
		return binary.LittleEndian
	} else if flag == -1 {
		return binary.BigEndian
	} else {
		panic("Unrecognized endianness flag.")
	}
}

func readRawGotetraHeader(file string, out *rawGotetraHeader) error {
	f, err := os.OpenFile(file, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	// order doesn't matter for this read, since flags are symmetric.
	order := endianness(readInt32(f, binary.LittleEndian))

	headerSize := readInt32(f, order)
	if headerSize != int32(unsafe.Sizeof(rawGotetraHeader{})) {
		return fmt.Errorf("Expected catalog.SheetHeader size of %d, found %d.",
			unsafe.Sizeof(rawGotetraHeader{}), headerSize,
		)
	}

	f.Seek(4+4, 0)
	err = binary.Read(f, order, out)
	return err
}

func loadSheetHeader(
	file string, hdBuf *gotetraHeader,
) (*os.File, binary.ByteOrder, error) {
	f, err := os.OpenFile(file, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, binary.LittleEndian, err
	}

	// order doesn't matter for this read, since flags are symmetric.
	order := endianness(readInt32(f, binary.LittleEndian))

	headerSize := readInt32(f, order)
	if headerSize != int32(unsafe.Sizeof(rawGotetraHeader{})) {
		return nil, binary.LittleEndian,
			fmt.Errorf("Expected catalog.SheetHeader size of %d, found %d.",
				unsafe.Sizeof(rawGotetraHeader{}), headerSize,
			)
	}

	_, err = f.Seek(4+4, 0)
	if err != nil {
		return nil, binary.LittleEndian, err
	}

	err = binary.Read(f, order, &hdBuf.rawGotetraHeader)
	if err != nil {
		return nil, binary.LittleEndian, err
	}

	// Deals with a bug in the current gotetra version.
	cw := hdBuf.CountWidth
	hdBuf.Count = cw * cw * cw

	return f, order, nil
}

func (buf *GotetraBuffer) ReadHeader(fname string, out *Header) error {
	f, _, err := loadSheetHeader(fname, &buf.hd)
	if err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}

	buf.hd.postprocess(out)

	return nil
}

// ReadPositionsAt reads the velocities in the given file into a buffer.
func readSheetPositionsAt(file string, xsBuf [][3]float32) error {
	h := &gotetraHeader{}
	f, order, err := loadSheetHeader(file, h)
	if err != nil {
		return nil
	}

	if h.GridCount != int64(len(xsBuf)) {
		return fmt.Errorf("Position buffer has length %d, but file %s has %d "+
			"vectors.", len(xsBuf), file, h.GridCount)
	}

	// Go to block 4 in the file.
	// The file pointer should already be here, but let's just be safe, okay?
	f.Seek(int64(4+4+int(unsafe.Sizeof(rawGotetraHeader{}))), 0)
	if err := readVecAsByte(f, order, xsBuf); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}
	return nil
}
