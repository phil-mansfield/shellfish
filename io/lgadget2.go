package io

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
)

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

func (gh *lGadget2Header) postprocess(
	xs [][3]float32, context Context, out *Header,
) {
	// Assumes the catalog has already been checked for corruption.
	
	out.TotalWidth = gh.BoxSize

	out.N = lgadgetParticleNum(gh.NPart, gh, context)

	out.Cosmo.Z = gh.Redshift
	out.Cosmo.OmegaM = gh.Omega0
	out.Cosmo.OmegaL = gh.OmegaLambda
	out.Cosmo.H100 = gh.HubbleParam

	out.Origin, out.Width = boundingBox(xs, gh.BoxSize)
}

func lgadgetParticleNum(
	npart [6]uint32, gh *lGadget2Header, context Context,
) int64 {
	if context.LGadgetNPartNum == 2 {
		if npart[0] > 100 * 1000 {
			panic(
				"Simulation contains too many particles. This is probably " +
				"because GadgetNpartNum is set to 2 when it " +
				"should be set to 1.",
			)
		}
		return int64(npart[1]) + int64(uint32(npart[0])) << 32
	} else {
		return int64(npart[0])
	}
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

func readLGadget2Header(
	path string, order binary.ByteOrder, out *lGadget2Header,
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

func (buf *LGadget2Buffer) readLGadget2Particles(
	path string,
	order binary.ByteOrder,
	xsBuf, vsBuf [][3]float32,
	msBuf []float32,
	idsBuf []int64,
) (xs, vs [][3]float32, ms []float32, ids []int64, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer f.Close()

	gh := &lGadget2Header{}

	_ = readInt32(f, order)
	binary.Read(f, binary.LittleEndian, gh)
	_ = readInt32(f, order)

	count := lgadgetParticleNum(gh.NPart, gh, buf.context)

	xsBuf = expandVectors(xsBuf[:0], int(count))
	vsBuf = expandVectors(vsBuf[:0], int(count))
	idsBuf = expandInts(idsBuf[:0], int(count))

	_ = readInt32(f, order)
	readVecAsByte(f, order, xsBuf)
	_ = readInt32(f, order)
	_ = readInt32(f, order)
	readVecAsByte(f, order, vsBuf)
	_ = readInt32(f, order)
	_ = readInt32(f, order)
	readInt64AsByte(f, order, idsBuf)

	// Fix periodicity of particles and convert the units of our velocities.

	rootA := float32(math.Sqrt(float64(gh.Time)))

	tw := float32(gh.BoxSize)
	for i := range xsBuf {
		for j := 0; j < 3; j++ {
			vsBuf[i][j] = vsBuf[i][j] * rootA
			
			if xsBuf[i][j] < 0 {
				xsBuf[i][j] += tw
			} else if xsBuf[i][j] >= tw {
				xsBuf[i][j] -= tw
			}

			if math.IsNaN(float64(xsBuf[i][j])) ||
				math.IsInf(float64(xsBuf[i][j]), 0) ||
				xsBuf[i][j] < -tw || xsBuf[i][j] > 2*tw {

				return nil, nil, nil, nil, fmt.Errorf(
					"Corruption detected in the file %s. I can't analyze it.",
					path,
				)
			}
		}
	}

	msBuf = expandScalars(msBuf, int(count))
	for i := range msBuf {
		msBuf[i] = buf.mass
	}

	return xsBuf, vsBuf, msBuf, idsBuf, nil
}

func expandVectors(vecs [][3]float32, n int) [][3]float32 {
	switch {
	case cap(vecs) >= n:
		return vecs[:n]
	case int(float64(cap(vecs))*1.5) > n:
		return append(vecs[:cap(vecs)],
			make([][3]float32, n-cap(vecs))...)
	default:
		return make([][3]float32, n)
	}
}

func expandScalars(scalars []float32, n int) []float32 {
	switch {
	case cap(scalars) >= n:
		return scalars[:n]
	case int(float64(cap(scalars))*1.5) > n:
		return append(scalars[:cap(scalars)],
			make([]float32, n-cap(scalars))...)
	default:
		return make([]float32, n)
	}
}

func expandInts(ints []int64, n int) []int64 {
	switch {
	case cap(ints) >= n:
		return ints[:n]
	case int(float64(cap(ints))*1.5) > n:
		return append(ints[:cap(ints)], make([]int64, n-cap(ints))...)
	default:
		return make([]int64, n)
	}
}

type LGadget2Buffer struct {
	open     bool
	order    binary.ByteOrder
	hd       lGadget2Header
	mass     float32
	xs, vs   [][3]float32
	ms       []float32
	ids      []int64
	context  Context
}

func NewLGadget2Buffer(
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

	buf := &LGadget2Buffer{order: order, context: context}
	err := readLGadget2Header(path, order, &buf.hd)
	if err != nil {
		return nil, err
	}

	c := CosmologyHeader{
		Z: buf.hd.Redshift, OmegaM: buf.hd.Omega0,
		OmegaL: buf.hd.OmegaLambda, H100: buf.hd.HubbleParam,
	}
	
	totCount, err := buf.TotalParticles(path)
	if err != nil { return nil, err }

	buf.mass = calcUniformMass(int64(totCount), buf.hd.BoxSize, c)

	return buf, nil
}

func (buf *LGadget2Buffer) Read(fname string) (
	xs, vs [][3]float32, ms []float32, ids []int64, err error,
) {
	if buf.open {
		panic("Buffer already open.")
	}
	buf.open = true

	buf.xs, buf.vs, buf.ms, buf.ids, err = buf.readLGadget2Particles(
		fname, buf.order, buf.xs, buf.vs, buf.ms, buf.ids,
	)

	return buf.xs, buf.vs, buf.ms, buf.ids, err
}

func (buf *LGadget2Buffer) Close() {
	if !buf.open {
		panic("Buffer not open.")
	}
	buf.open = false
}

func (buf *LGadget2Buffer) IsOpen() bool {
	return buf.open
}

func (buf *LGadget2Buffer) ReadHeader(fname string, out *Header) error {
	err := readLGadget2Header(fname, buf.order, &buf.hd)
	if err != nil {
		return err
	}
	defer buf.Close()
	xs, _, _, _, err := buf.Read(fname)
	if err != nil {
		return err
	}

	buf.hd.postprocess(xs, buf.context, out)

	return nil
}

func (buf *LGadget2Buffer) MinMass() float32 { return buf.mass }

func (buf *LGadget2Buffer) TotalParticles(fname string) (int, error) {
	hd := &lGadget2Header{}
	err := readLGadget2Header(fname, buf.order, hd)
	if err != nil { return 0, err }
	return int(lgadgetParticleNum(hd.NPartTotal, hd, buf.context)), nil
}
