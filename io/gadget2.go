package io

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

// gadgetHeader is the formatting for meta-information used by Gadget 2.
type gadget2Header struct {
	NPart                      [6]uint32
	Mass                       [6]float64
	Time, Redshift             float64
	FlagSfr, FlagFeedback      int32
	NumPartTotal                 [6]uint32
	FlagCooling, NumFiles      int32
	BoxSize, Omega0            float64
	OmegaLambda, HubbleParam   float64
	FlagStellarAge, FlagMetals int32
	NumPartTotalHW             [6]uint32
	FlagEntropyICs             int32

	Padding [56]byte
}

type Gadget2Header gadget2Header

func (gh *gadget2Header) postprocess(
	xs [][3]float32, context *Context, out *Header,
) {
	// Assumes the catalog has already been checked for corruption.
	
	out.TotalWidth = gh.BoxSize

	out.N = 0
	for _, i := range context.GadgetDMTypeIndices {
		out.N += int64(gh.NPart[i])
	}


	out.Cosmo.Z = gh.Redshift
	out.Cosmo.OmegaM = gh.Omega0
	out.Cosmo.OmegaL = gh.OmegaLambda
	out.Cosmo.H100 = gh.HubbleParam

	out.Origin, out.Width = boundingBox(xs, gh.BoxSize)
}

func readGadget2Header(
	path string, order binary.ByteOrder, out *gadget2Header,
) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_ = readInt32(f, order)
	err = binary.Read(f, binary.LittleEndian, out)
	return err
}

func (buf *Gadget2Buffer) readGadget2Particles(
	path string,
	order binary.ByteOrder,
	xsBuf, vsBuf [][3]float32,
	multiMsBuf, msBuf []float32,
	idsBuf []int64,
) (xs, vs [][3]float32, multiMs, ms []float32, ids []int64, err error) {

	// Open the buffer and read the raw gadget header.

	f, err := os.Open(path)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	defer f.Close()

	gh := &gadget2Header{}

	_ = readInt32(f, order)
	binary.Read(f, order, gh)
	_ = readInt32(f, order)

	// Figure out particle counts so we can size buffers correctly.

	totalN := particleCount(gh)
	dmN := dmCount(gh, &buf.context)
	multiN := multiMassParticleCount(gh, &buf.context)

	// Resize buffers

	xsBuf = expandVectors(xsBuf[:0], totalN)
	vsBuf = expandVectors(vsBuf[:0], totalN)
	idsBuf = expandInts(idsBuf[:0], totalN)
	msBuf = expandScalars(msBuf[:0], totalN)
	multiMsBuf = expandScalars(multiMsBuf[:0], multiN)

	// Read all particles into buffers. Because Gadget uses Fortran-style
	// binary formatting, we can do consistency checks on our buffers.

	// Positions
	xSize := readInt32(f, order)
	if int(xSize) != 12 * len(xsBuf) {
		return nil, nil, nil, nil, nil, fmt.Errorf(
			"Position block size is %d, but expected %d",
			xSize, 12 * len(xsBuf),
		)
	}
	readVecAsByte(f, order, xsBuf)
	_ = readInt32(f, order)

	// Velocities
	vSize := readInt32(f, order)
	if int(vSize) != 4 * len(vsBuf) {
		return nil, nil, nil, nil, nil, fmt.Errorf(
			"Velocity block size is %d, but expected %d",
			vSize, 12 * len(xsBuf),
		)
	}
	readVecAsByte(f, order, vsBuf)
	_ = readInt32(f, order)

	// IDs
	idSize := readInt32(f, order)
	if int(idSize) != 8 * len(idsBuf) {
		return nil, nil, nil, nil, nil, fmt.Errorf(
			"ID block size is %d, but expected %d",
			idSize, 12 * len(xsBuf),
		)
	}
	readInt64AsByte(f, order, idsBuf)
	_ = readInt32(f, order)

	// Masses
	mSize := readInt32(f, order)
	if int(mSize) != 12 * len(xsBuf) {
		return nil, nil, nil, nil, nil, fmt.Errorf(
			"Mass block size is %d, but expected %d",
			mSize, 12 * len(xsBuf),
		)
	}
	readFloat32AsByte(f, order, multiMsBuf)
	_ = readInt32(f, order)

	// Expand uniform mass types

	unpackMass(gh, &buf.context, multiMsBuf, msBuf)

	// Remove non-DM particle types

	packVec(gh, &buf.context, xsBuf)
	packVec(gh, &buf.context, vsBuf)
	packInt64(gh, &buf.context, idsBuf)
	packFloat32(gh, &buf.context, msBuf)

	// Resize
	xsBuf = xsBuf[0: dmN]
	vsBuf = vsBuf[0: dmN]
	idsBuf = idsBuf[0: dmN]
	msBuf = msBuf[0: dmN]

	err = fix(gh, &buf.context, path, xsBuf, vsBuf, msBuf)

	return xsBuf, vsBuf, multiMsBuf, msBuf, idsBuf, err
}

func isMultiMass(context *Context, i int) bool {
	for _, j := range context.GadgetDMSingleMassIndices {
		if int(j) == i { return false }
	}
	return true
}

func isDM(context *Context, i int) bool {
	for _, j := range context.GadgetDMSingleMassIndices {
		if int(j) == i { return false }
	}
	return false
}

func particleCount(gh *gadget2Header) int {
	n := 0
	for i := range gh.NPart { n += int(gh.NPart[i]) }
	return n
}

func dmCount(gh *gadget2Header, context *Context) int {
	n := 0
	for i := range gh.NPart {
		if isDM(context, i) { n += int(gh.NPart[i]) }
	}
	return n
}

func multiMassParticleCount(gh *gadget2Header, context *Context) int {
	n := 0
	for i := range gh.NPart {
		if isMultiMass(context, i) { n += int(gh.NPart[i]) }
	}
	return n
}

func unpackMass(gh *gadget2Header, context *Context, multiMs, ms []float32) {
	// Broken up into two passes to make it easier to read.

	// First pass to find the indices of breaks in particle types.

	multiOffsets := make([]int, 7)
	offsets := make([]int, 7)
	for i := 0; i < 6; i++ {
		if isMultiMass(context, i) {
			multiOffsets[i + 1] = multiOffsets[i] + int(gh.NPart[i])
		} else {
			multiOffsets[i + 1] = multiOffsets[i]
		}

		offsets[i + 1] = offsets[i] + int(gh.NPart[i])
	}

	// Second pass to copy and expand values accordingly.

	for i := 0; i < 6; i++ {
		if isMultiMass(context, i) {
			copy(
				ms[offsets[i]: offsets[i + 1]],
				multiMs[multiOffsets[i]: multiOffsets[i + 1]],
			)
		} else {
			for j := offsets[i]; j < offsets[i + 1]; j++ {
				ms[j] = float32(gh.Mass[i])
			}
		}
	}
}

func packVec(gh *gadget2Header, context *Context, buf [][3]float32) {
	dmOffsets := make([]int, 7)
	offsets := make([]int, 7)
	for i := 0; i < 6; i++ {
		if isDM(context, i) {
			dmOffsets[i + 1] = dmOffsets[i] + int(gh.NPart[i])
		} else {
			dmOffsets[i + 1] = dmOffsets[i]
		}

		offsets[i + 1] = offsets[i] + int(gh.NPart[i])
	}

	for i := 0; i < 6; i++ {
		if isDM(context, i) {
			copy(
				buf[offsets[i]: offsets[i + 1]],
				buf[dmOffsets[i]: dmOffsets[i+1]],
			)
		}
	}
}

func packInt64(gh *gadget2Header, context *Context, buf []int64) {
	dmOffsets := make([]int, 7)
	offsets := make([]int, 7)
	for i := 0; i < 6; i++ {
		if isDM(context, i) {
			dmOffsets[i + 1] = dmOffsets[i] + int(gh.NPart[i])
		} else {
			dmOffsets[i + 1] = dmOffsets[i]
		}

		offsets[i + 1] = offsets[i] + int(gh.NPart[i])
	}

	for i := 0; i < 6; i++ {
		if isDM(context, i) {
			copy(
				buf[offsets[i]: offsets[i + 1]],
				buf[dmOffsets[i]: dmOffsets[i+1]],
			)
		}
	}
}

func packFloat32(gh *gadget2Header, context *Context, buf []float32) {
	dmOffsets := make([]int, 7)
	offsets := make([]int, 7)
	for i := 0; i < 6; i++ {
		if isDM(context, i) {
			dmOffsets[i + 1] = dmOffsets[i] + int(gh.NPart[i])
		} else {
			dmOffsets[i + 1] = dmOffsets[i]
		}

		offsets[i + 1] = offsets[i] + int(gh.NPart[i])
	}

	for i := 0; i < 6; i++ {
		if isDM(context, i) {
			copy(
				buf[offsets[i]: offsets[i + 1]],
				buf[dmOffsets[i]: dmOffsets[i+1]],
			)
		}
	}
}

// Fix periodicity and units.
func fix(
	gh *gadget2Header, context *Context, path string,
	xs, vs [][3]float32, ms[]float32,
) error {
	rootA := float32(math.Sqrt(float64(gh.Time)))

	tw := float32(gh.BoxSize)
	for i := range xs {
		for j := 0; j < 3; j++ {
			vs[i][j] = vs[i][j] * rootA
			
			if xs[i][j] < 0 {
				xs[i][j] += tw
			} else if xs[i][j] >= tw {
				xs[i][j] -= tw
			}

			if math.IsNaN(float64(xs[i][j])) ||
				math.IsInf(float64(xs[i][j]), 0) ||
				xs[i][j] < -tw || xs[i][j] > 2*tw {

				return fmt.Errorf(
					"Corruption detected in the file %s. I can't analyze it.",
					path,
				)
			}
		}
	}

	return nil
}

type Gadget2Buffer struct {
	open        bool
	order       binary.ByteOrder
	hd          gadget2Header
	mass        float32
	xs, vs      [][3]float32
	ms, multiMs []float32
	ids         []int64
	context     Context
}

func NewGadget2Buffer(
	path, orderFlag string, context Context,
) (VectorBuffer, error) {
	
	var order binary.ByteOrder = binary.LittleEndian
	switch orderFlag {
	case "LittleEndian":
	case "BigEndian":
		order = binary.BigEndian
	case "SystemOrder":
		if !IsSysOrder(order) {
			order = binary.BigEndian
		}
	}

	buf := &Gadget2Buffer{order: order}
	err := readGadget2Header(path, order, &buf.hd)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (buf *Gadget2Buffer) Read(fname string) (
	xs, vs [][3]float32, ms []float32, ids []int64, err error,
) {
	if buf.open {
		panic("Buffer already open.")
	}
	buf.open = true

	// I am not proud of this
	buf.xs, buf.vs, buf.multiMs, buf.ms, buf.ids, err =
	buf.readGadget2Particles(
		fname, buf.order, buf.xs, buf.vs, buf.multiMs, buf.ms, buf.ids,
	)

	return buf.xs, buf.vs, buf.ms, buf.ids, err
}

func (buf *Gadget2Buffer) Close() {
	if !buf.open {
		panic("Buffer not open.")
	}
	buf.open = false
}

func (buf *Gadget2Buffer) IsOpen() bool {
	return buf.open
}

func (buf *Gadget2Buffer) ReadHeader(fname string, out *Header) error {
	err := readGadget2Header(fname, buf.order, &buf.hd)
	if err != nil {
		return err
	}
	defer buf.Close()
	xs, _, _, _, err := buf.Read(fname)
	if err != nil {
		return err
	}

	buf.hd.postprocess(xs, &buf.context, out)

	return nil
}

func (buf *Gadget2Buffer) MinMass() float32 { return buf.mass }

func (buf *Gadget2Buffer) TotalParticles(fname string) (int, error) {
	hd := &gadget2Header{}
	err := readGadget2Header(fname, buf.order, hd)
	if err != nil { return 0, err }

	n := 0
	for _, i := range buf.context.GadgetDMTypeIndices {
		n += int(int(hd.NumPartTotal[i]) + int(hd.NumPartTotalHW[i]) << 32)
	}

	return n, nil
}