package io

// ParticleBuffer is a wrapper around catalog files which allows floats to be
// appended to files on the fly without much overhead from I/O or without
// excessive memory usage/reallocating.
type ParticleBuffer struct {
	xBuf, vBuf [][3]float32
	idBuf []int64
	idx int
	xs, vs [][3]float32
}

// NewParticleBuffer creates a ParticleBuffer associated with the given file.
func NewParticleBuffer(xs, vs [][3]float32, bufSize int) *ParticleBuffer {
	pb := &ParticleBuffer{
		make([][3]float32, bufSize),
		make([][3]float32, bufSize),
		make([]int64, bufSize),
		0, xs, vs,
	}
	return pb
}

// Append adds a value to the float buffer, which will eventually be
// written to the target file.
func (pb *ParticleBuffer) Append(xBuf, vBuf [][3]float32, idBuf []int64) {
	for pi := range xBuf {
		pb.xBuf[pb.idx] = xBuf[pi]
		pb.vBuf[pb.idx] = vBuf[pi]
		pb.idBuf[pb.idx] = idBuf[pi]
		
		pb.idx++
		if pb.idx == len(pb.xBuf) {
			pb.Flush()
		}
	}
}

// Flush writes the contents of the buffer to its target file. This will
// be called automatically whenever the buffer fills.
func (pb *ParticleBuffer) Flush() {
	for i := 0; i < pb.idx; i++ {
		pb.xs[pb.idBuf[i] - 1] = pb.xBuf[i]
		pb.vs[pb.idBuf[i] - 1] = pb.vBuf[i]
	}
	pb.idx = 0
}
