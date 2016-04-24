package io

import (
	"encoding/binary"
	"io"
	"math"
	"reflect"

	"unsafe"
)

// Not threadsafe, obviously.
type VectorBuffer interface {
	Read(fname string) ([][3]float32, error)
	Close()
	IsOpen() bool
	ReadHeader(fname string, out *Header) error
}

type Header struct {
	Cosmo CosmologyHeader
	N int64
	Count int64
	TotalWidth float64
	Origin, Width [3]float32
}

func readVecAsByte(rd io.Reader, end binary.ByteOrder, buf [][3]float32) error {
	bufLen := len(buf)

	hd := *(*reflect.SliceHeader)(unsafe.Pointer(&buf))
	hd.Len *= 12
	hd.Cap *= 12
	
	byteBuf := *(*[]byte)(unsafe.Pointer(&hd))
	_, err := rd.Read(byteBuf)
	if err != nil { return err }

	if !isSysOrder(end) {
		for i := 0; i < bufLen * 3; i++ {
			for j := 0; j < 2; j++ {
				idx1, idx2 := i*4 + j, i*4 + 3 - j
				byteBuf[idx1], byteBuf[idx2] = byteBuf[idx2], byteBuf[idx1]
			}
		}
	}

	hd.Len /= 12
	hd.Cap /= 12

	return nil
}

func readInt64AsByte(rd io.Reader, end binary.ByteOrder, buf []int64) error {
	bufLen := len(buf)

	hd := *(*reflect.SliceHeader)(unsafe.Pointer(&buf))
	hd.Len *= 8
	hd.Cap *= 8
	
	byteBuf := *(*[]byte)(unsafe.Pointer(&hd))
	_, err := rd.Read(byteBuf)
	if err != nil { return err }

	if !isSysOrder(end) {
		for i := 0; i < bufLen; i++ {
			for j := 0; j < 4; j++ {
				idx1, idx2 := i*8 + j, i*8 + 7 - j
				byteBuf[idx1], byteBuf[idx2] = byteBuf[idx2], byteBuf[idx1]
			}
		}
	}

	hd.Len /= 8
	hd.Cap /= 8

	return nil
}

func isSysOrder(end binary.ByteOrder) bool {
	buf32 := []int32{1}

	hd := *(*reflect.SliceHeader)(unsafe.Pointer(&buf32))
	hd.Len *= 4
	hd.Cap *= 4

	buf8 := *(*[]int8)(unsafe.Pointer(&hd))
	if buf8[0] == 1 {
		return binary.LittleEndian == end
	}
	return binary.BigEndian == end
}

type CellBounds struct {
	Origin, Width [3]int
}

func (hd *Header) CellBounds(cells int) *CellBounds {
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