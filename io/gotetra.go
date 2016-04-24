package io

import (
	"encoding/binary"
	"fmt"
	"os"
	
	"unsafe"
)

type GotetraBuffer struct {
	open   bool

	sheet  [][3]float32
	out    [][3]float32

	sw, gw int

	hd     gotetraHeader
}

func NewGotetraBuffer(fname string) (VectorBuffer, error) {
	hd := &gotetraHeader{}
	f, _, err := loadSheetHeader(fname, hd)
	if err != nil { return nil, err }
	err = f.Close()
	if err != nil { return nil, err }

	sw, gw := hd.SegmentWidth, hd.GridWidth
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
	buf.open = true
	
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
type rawGotetraHeader struct {
	Cosmo                              CosmologyHeader
	Count, CountWidth                  int64
	SegmentWidth, GridWidth, GridCount int64
	Idx, Cells                         int64

	Mass                               float64
	TotalWidth                         float64

	Origin, Width                      [3]float32
	VelocityOrigin, VelocityWidth      [3]float32
}

func (raw *rawGotetraHeader) postprocess(hd *Header) {
	hd.Cosmo = raw.Cosmo
	hd.Count = raw.CountWidth*raw.CountWidth*raw.CountWidth
	
	hd.N = raw.SegmentWidth * raw.SegmentWidth * raw.SegmentWidth
	hd.TotalWidth = raw.TotalWidth
	hd.Origin, hd.Width = raw.Origin, raw.Width
}

type gotetraHeader struct {
	rawGotetraHeader
	N int64
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

func loadSheetHeader(
	file string, hdBuf *gotetraHeader,
) (*os.File, binary.ByteOrder, error) {
	f, err := os.OpenFile(file, os.O_RDONLY, os.ModePerm)
	if err != nil { return nil, binary.LittleEndian, err }

	// order doesn't matter for this read, since flags are symmetric.
	order := endianness(readInt32(f, binary.LittleEndian))

	headerSize := readInt32(f, order)
	if headerSize != int32(unsafe.Sizeof(rawGotetraHeader{})) {
		return nil, binary.LittleEndian,
		fmt.Errorf("Expected catalog.SheetHeader size of %d, found %d.",
			unsafe.Sizeof(rawGotetraHeader{}), headerSize,
		)
	}

	_, err = f.Seek(4 + 4, 0)
	if err != nil { return nil, binary.LittleEndian, err }

	err = binary.Read(f, order, &hdBuf.rawGotetraHeader)
	if err != nil { return nil, binary.LittleEndian, err }

	return f, order, nil
}

func (buf *GotetraBuffer) ReadHeader(fname string, out *Header) error {
	f, _, err := loadSheetHeader(fname, &buf.hd)
	if err != nil { return err }
	if err = f.Close(); err != nil { return err }

	buf.hd.postprocess(out)

	return nil
}

// ReadPositionsAt reads the velocities in the given file into a buffer.
func readSheetPositionsAt(file string, xsBuf [][3]float32) error {
	h := &gotetraHeader{}
	f, order, err := loadSheetHeader(file, h)
	if err != nil { return nil }

	if h.GridCount != int64(len(xsBuf)) {
		return fmt.Errorf("Position buffer has length %d, but file %s has %d " +
		"vectors.", len(xsBuf), file, h.GridCount)
	}

	// Go to block 4 in the file.
	// The file pointer should already be here, but let's just be safe, okay?
	f.Seek(int64(4 + 4 + int(unsafe.Sizeof(rawGotetraHeader{}))), 0)
	if err := readVecAsByte(f, order, xsBuf); err != nil { return err }

	if err := f.Close(); err != nil { return err }
	return nil
}