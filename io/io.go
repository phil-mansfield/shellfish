/*The io package contains code for reading particles from disk. It provides
an abstract interface (VectorBuffer) that allows the main package to read
particles from different catalog formats in the same way.

Chances are that if you're reading this, you want to add a new type of catalog
type to this project. (If you're trying to do something more complicated than
that, chances are that you know what you're doing and aren't bothering with this
package comment). It's not too painful to do that in most cases. I'll lay out
exactly what this is going to look like. For the sake of the discussion I'm
going pretend that you're adding the a file-type called "my_file" to Shellfish.

0. Read up on the usual instructions for editing Shellfish code. You can find
these in the main documentation directory.

1. Make a file in this directory called my_file.go.

2. Make a struct in that file named "MyFileBuffer". Write five methods for that
struct that have the same names and type signatures as those found in the
VectorBuffer interface (the first declaration in this file).

3. Write a function "NewMyFileBuffer(...) *MyFileBuffer" that returns a new
instance of your buffer that is initialized with all the information that it
needs (or any work that you don't want to be repeated). Usually this will just
mean taking in an argument representing the byte ordering that the user
requested and, storing that in the struct and maybe allocating a couple internal
buffers. Sometimes you might also want the name of the directory containing
the files.

4. Implement each of those methods so that they do what the VectorBuffer
comments indicate that they should do. Only two are non-trivial: Read() and
(to a lesser extent) ReadHeader(). Look at the example in lgadget2.go and copy
code as needed. If your file format is just  couple arrays of particles with
some type of header (which it probably is), you can copy almost all of it and
won't have to do much.

5. Update getVectorBuffer() in cmd/util.go. It'll just be adding a case to a
switch statement.

6. Update the config file so that it knows about your file type. This means
going to cmd/cmd.go and doing two things: First, change the validate() method
there so that it doesn't crash when you pass it the name of your file type (i.e.
adding it to the first line of the config.SnapshotLine switch statement in
cmd.go). Second, go into the example config file (the big string in the same
file) and in the SnapshotType comment, explain that your file type is also
supported now.
*/
package io

import (
	"encoding/binary"
	"io"
	"reflect"
	
	"unsafe"
)

// Not threadsafe, obviously.
type VectorBuffer interface {
	// Positions in Mpc/h and masses in Msun/h.
	Read(fname string) (xs, vs [][3]float32, ms []float32, ids []int64, err error)
	Close()
	IsOpen() bool
	ReadHeader(fname string, out *Header) error
	// The minimum mass of all the particles in the simulation.
	MinMass() float32
	TotalParticles(fname string) (int, error)
}

// CosmologyHeader contains information describing the cosmological
// context in which the simulation was run.
type CosmologyHeader struct {
	Z      float64
	OmegaM float64
	OmegaL float64
	H100   float64
}

type Header struct {
	Cosmo            CosmologyHeader
	N                int64
	TotalWidth       float64
	Origin, Width    [3]float32
}

type Context struct {
	LGadgetNPartNum int64
	GadgetDMTypeIndices []int64
	GadgetDMSingleMassIndices []int64
	GadgetMassUnits float64
	GadgetPositionUnits float64
}

func reorder(buf []byte, size, words int) {
	for word := 0; word < words; word++ {
		for i := 0; i < size/2; i++ {
			i1, i2 := word*size + i, (word + 1)*size - (i + 1)
			buf[i1], buf[i2] = buf[i2], buf[i1]
		}
	}
}

func readBolshoiParticleAsByte(
	rd io.Reader, end binary.ByteOrder, buf []bolshoiParticle,
) error {
	bufLen := len(buf)
	
	hd := *(*reflect.SliceHeader)(unsafe.Pointer(&buf))
	hd.Len *= int(unsafe.Sizeof(bolshoiParticle{}))
	hd.Cap *= int(unsafe.Sizeof(bolshoiParticle{}))

	byteBuf := *(*[]byte)(unsafe.Pointer(&hd))
	_, err := rd.Read(byteBuf)
	if err != nil {
		return err
	}

	if !IsSysOrder(end) {
		for i := 0; i < bufLen; i++ {
			offset := i * 32
			reorder(byteBuf[offset: offset+24], 4, 6)
			reorder(byteBuf[offset+24: offset+24+8], 8, 1)
		}
	}

	hd.Len /= int(unsafe.Sizeof(bolshoiParticle{}))
	hd.Cap /= int(unsafe.Sizeof(bolshoiParticle{}))

	return nil
}

func readVecAsByte(rd io.Reader, end binary.ByteOrder, buf [][3]float32) error {
	bufLen := len(buf)

	hd := *(*reflect.SliceHeader)(unsafe.Pointer(&buf))
	hd.Len *= 12
	hd.Cap *= 12

	byteBuf := *(*[]byte)(unsafe.Pointer(&hd))
	_, err := rd.Read(byteBuf)
	if err != nil {
		return err
	}

	if !IsSysOrder(end) {
		for i := 0; i < bufLen*3; i++ {
			for j := 0; j < 2; j++ {
				idx1, idx2 := i*4+j, i*4+3-j
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
	if err != nil {
		return err
	}

	if !IsSysOrder(end) {
		for i := 0; i < bufLen; i++ {
			for j := 0; j < 4; j++ {
				idx1, idx2 := i*8+j, i*8+7-j
				byteBuf[idx1], byteBuf[idx2] = byteBuf[idx2], byteBuf[idx1]
			}
		}
	}

	hd.Len /= 8
	hd.Cap /= 8

	return nil
}

func readInt32AsByte(rd io.Reader, end binary.ByteOrder, buf []int32) error {
	bufLen := len(buf)

	hd := *(*reflect.SliceHeader)(unsafe.Pointer(&buf))
	hd.Len *= 4
	hd.Cap *= 4

	byteBuf := *(*[]byte)(unsafe.Pointer(&hd))
	_, err := rd.Read(byteBuf)
	if err != nil {
		return err
	}

	if !IsSysOrder(end) {
		for i := 0; i < bufLen; i++ {
			for j := 0; j < 2; j++ {
				idx1, idx2 := i*4+j, i*4+3-j
				byteBuf[idx1], byteBuf[idx2] = byteBuf[idx2], byteBuf[idx1]
			}
		}
	}

	hd.Len /= 4
	hd.Cap /= 4

	return nil
}

func readFloat32AsByte(rd io.Reader, end binary.ByteOrder, buf []float32) error {
	bufLen := len(buf)
	hd := *(*reflect.SliceHeader)(unsafe.Pointer(&buf))
	hd.Len *= 4
	hd.Cap *= 4

	byteBuf := *(*[]byte)(unsafe.Pointer(&hd))
	_, err := rd.Read(byteBuf)
	if err != nil {
		return err
	}

	if !IsSysOrder(end) {
		for i := 0; i < bufLen; i++ {
			for j := 0; j < 2; j++ {
				idx1, idx2 := i*4+j, i*4+3-j
				byteBuf[idx1], byteBuf[idx2] = byteBuf[idx2], byteBuf[idx1]
			}
		}
	}

	hd.Len /= 4
	hd.Cap /= 4

	return nil
}

func IsSysOrder(end binary.ByteOrder) bool {
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

func boundingBox(
	xs [][3]float32, totalWidth float64,
) (origin, width [3]float32) {
	// Assumes that the slice has already been checked for corruption.
	origin = xs[0]
	width = [3]float32{0, 0, 0}
	tw, tw2 := float32(totalWidth), float32(totalWidth)/2
	
	max, min := origin, origin

	for i := range xs {
		for j := 0; j < 3; j++ {
			x, x0, w := xs[i][j], origin[j], width[j]

			if x > x0 && x < x0 + w { continue }
						
			if x-x0 > tw2 {
				x -= tw
			} else if x-x0 < -tw2 {
				x += tw
			}

			if x < x0 {
				width[j] += x0 - x
				origin[j] = x
				min[j] = x
			} else if x-x0 > w {
				width[j] = x - x0
				max[j] = x
			}
		}
	}

	for j := 0; j < 3; j++ {
		if width[j] > tw2 {
			width[j] = tw
			origin[j] = 0
		}
	}
	
	return origin, width
}
