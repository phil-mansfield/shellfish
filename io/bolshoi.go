package io

import (
	"encoding/binary"
	"os"
	"fmt"
)

// Unfortunately, the Bolshoi boundary region is too small for us to get away
// with only needing to read a single segment file. Thus, we must go through
// the whole song-and-dance

type BolshoiBuffer struct { }

type bolshoiHeader1 struct {
	aExpn, aStep float32
	iStep int32
	nRowCells, nGridCells int32
	nSpecies, nSeed int32
	omegaM, omegaL, hubble, boxWidth float32
}

type bolshoiHeader2 struct {
	k int32
	nx, ny, nz int32
	dR float32
}


func NewBolshoiBuffer(
	path string, context Context,
) (VectorBuffer, error) {
	return nil, nil
}

func (bol *BolshoiBuffer) Read(fname string) (
	xs, vs [][3]float32, ms []float32, ids []int64, err error,
) {
	return nil, nil, nil, nil, nil
}

func (bol *BolshoiBuffer)  Close() { }

func (bol *BolshoiBuffer) IsOpen() bool { return false }

func (bol *BolshoiBuffer) ReadHeader(fname string, out *Header) error {
	return nil

	f, err := os.Open(fname)
	if err != nil { return nil }
	defer f.Close()

	order := binary.LittleEndian
	bh1 := &bolshoiHeader1{}
	bh2 := &bolshoiHeader2{}

	_ = readInt32(f, order)
	binary.Read(f, order, bh1)
	_ = readInt32(f, order)

	_ = readInt32(f, order)
	binary.Read(f, order, bh2)
	_ = readInt32(f, order)

	fmt.Println(bh1)
	fmt.Println(bh2)

	return nil
}

func (bol *BolshoiBuffer) MinMass() float32 { return -1 }

func (bol *BolshoiBuffer) TotalParticles(fname string) (int, error) {
	return -1, nil
}
