package io

import (
	"strconv"
)

type NilBuffer struct {
	context Context
}

func NewNilBuffer(context Context) (VectorBuffer, error) {
	return &NilBuffer{ context }, nil
}

func (buf *NilBuffer) Read(fname string) (
	xs, vs [][3]float32, ms []float32, ids []int64, err error,
) {
	panic("Cannot call Read() on a NilBuffer. " +
		"Submit a bug report about this message.")
}

func (buf *NilBuffer) Close() {
	panic("Cannot call Close() on a NilBuffer. " +
		"Submit a bug report about this message.")
}

func (buf *NilBuffer) IsOpen() bool { return false }

func (buf *NilBuffer) ReadHeader(fname string, out *Header) error {
	idx, err := strconv.Atoi(fname)
	if err != nil { err.Error() }

	cosmo := CosmologyHeader{
		OmegaM: buf.context.NilOmegaM,
		OmegaL: buf.context.NilOmegaL,
		H100: buf.context.NilH100,
		Z: 1/buf.context.NilScaleFactors[idx] - 1,
	}
	
	*out = Header{
		Cosmo: cosmo,
		N: -1,
		TotalWidth: buf.context.NilTotalWidth,
		Origin: [3]float32{0, 0, 0},
		Width: [3]float32{
			float32(buf.context.NilTotalWidth),
			float32(buf.context.NilTotalWidth),
			float32(buf.context.NilTotalWidth),
		},
	}
	
	return nil
}

func (buf *NilBuffer) MinMass() float32 {
	panic("Cannot call MinMass() on a NilBuffer. " +
		"Submit a bug report about this message.")
}

func (buf *NilBuffer) TotalParticles(fname string) (int, error) {
	panic("Cannot call TotalParticles() on a NilBuffer. " +
		"Submit a bug report about this message.")
}
