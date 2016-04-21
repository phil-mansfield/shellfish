package io

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"

	"unsafe"
)

type GotetraBuffer struct {
	open bool

	sheet [][3]float32
	out [][3]float32

	sw, gw int

	hd SheetHeader
}

func NewGotetraBuffer(fname string) (VectorBuffer, error) {
	hd := &SheetHeader{}
	err := ReadSheetHeaderAt(fname, hd)
	return nil, err

	sw, gw := hd.segmentWidth, hd.gridWidth
	buf := &GotetraBuffer{
		sheet: make([][3]float32, gw * gw * gw),
		out: make([][3]float32, sw * sw * sw),
		open: false,
		sw: int(sw), gw: int(gw),
	}

	return buf, nil
}

func (buf *GotetraBuffer) IsOpen() bool { return buf.open }

func (buf *GotetraBuffer) Read(fname string) ([][3]float32, error) {
	if buf.open { panic("Buffer already open.") }

	err := readSheetPositionsAt(fname, buf.sheet)
	if err != nil { return nil, err }

	for z := 0; z < buf.sw; z++ {
		for y := 0; y < buf.sw; y++ {
			for x := 0; x < buf.sw; x++ {
				si := x + y*buf.sw + z*buf.sw*buf.sw
				gi := x + y*buf.gw + z*buf.gw*buf.gw
				buf.out[si] = buf.sheet[gi]
			}
		}
	}

	return buf.out, nil
}

func (buf *GotetraBuffer) Close() {
	if !buf.open { panic("Buffer already closed.") }

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
type sheetHeader struct {
	Cosmo CosmologyHeader
	Count, CountWidth int64
	segmentWidth, gridWidth, gridCount int64
	Idx, Cells int64

	Mass float64
	TotalWidth float64

	Origin, Width [3]float32
}

type SheetHeader struct {
	sheetHeader
	N int64
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

func readSheetHeaderAt(
file string, hdBuf *SheetHeader,
) (*os.File, binary.ByteOrder, error) {
	f, err := os.OpenFile(file, os.O_RDONLY, os.ModePerm)
	if err != nil { return nil, binary.LittleEndian, err }

	// order doesn't matter for this read, since flags are symmetric.
	order := endianness(readInt32(f, binary.LittleEndian))

	headerSize := readInt32(f, order)
	if headerSize != int32(unsafe.Sizeof(sheetHeader{})) {
		return nil, binary.LittleEndian,
		fmt.Errorf("Expected catalog.SheetHeader size of %d, found %d.",
			unsafe.Sizeof(sheetHeader{}), headerSize,
		)
	}

	_, err = f.Seek(4 + 4, 0)
	if err != nil { return nil, binary.LittleEndian, err }

	err = binary.Read(f, order, &hdBuf.sheetHeader)
	if err != nil { return nil, binary.LittleEndian, err }

	hdBuf.Count = hdBuf.CountWidth*hdBuf.CountWidth*hdBuf.CountWidth
	hdBuf.N = hdBuf.segmentWidth*hdBuf.segmentWidth*hdBuf.segmentWidth
	return f, order, nil
}

// ReadHeaderAt reads the header in the given file into the target Header.
func ReadSheetHeaderAt(file string, hdBuf *SheetHeader) error {
	f, _, err := readSheetHeaderAt(file, hdBuf)
	if err != nil { return err }
	if err = f.Close(); err != nil { return err }
	return nil
}

// ReadPositionsAt reads the velocities in the given file into a buffer.
func readSheetPositionsAt(file string, xsBuf [][3]float32) error {
	h := &SheetHeader{}
	f, order, err := readSheetHeaderAt(file, h)
	if err != nil { return nil }

	if h.gridCount != int64(len(xsBuf)) {
		return fmt.Errorf("Position buffer has length %d, but file %s has %d " +
		"vectors.", len(xsBuf), file, h.gridCount)
	}

	// Go to block 4 in the file.
	// The file pointer should already be here, but let's just be safe, okay?
	f.Seek(int64(4 + 4 + int(unsafe.Sizeof(SheetHeader{}))), 0)
	if err := readVecAsByte(f, order, xsBuf); err != nil { return err }

	if err := f.Close(); err != nil { return err }
	return nil
}

type CellBounds struct {
	Origin, Width [3]int
}

func (hd *SheetHeader) CellBounds(cells int) *CellBounds {
	cb := &CellBounds{}
	cellWidth := hd.TotalWidth / float64(cells)

	for j := 0; j < 3; j++ {
		cb.Origin[j] = int(
			math.Floor(float64(hd.Origin[j]) / cellWidth),
		)
		cb.Width[j] = 1 + int(
			math.Floor(float64(hd.Origin[j] + hd.Width[j]) / cellWidth),
		)

		cb.Width[j] -= cb.Origin[j]
	}

	return cb
}