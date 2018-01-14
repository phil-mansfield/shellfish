package io

import (
	"encoding/binary"
	"os"
)

// Unfortunately, the BolshoiP boundary region is too small for us to get away
// with only needing to read a single segment file. Thus, we must go through
// the whole song-and-dance

type BolshoiPBuffer struct {
	order binary.ByteOrder
	bh1 bolshoiPHeader1
	bh2 bolshoiPHeader2

	xReadBuf, vReadBuf [3][]float32
	
	xsBuf, vsBuf [][3]float32
	msBuf, wpBuf []float32
	idsBuf []int64
}

type bolshoiPHeader1 struct {
	AExpn, AStep float32
	IStep int32
	NRowCells, NGridCells int32
	NSpecies, NSeed int32
	OmegaM, OmegaL, H100, BoxWidth, Mass float32
}

type bolshoiPHeader2 struct {
	K int32
	Nx, Ny, Nz int32
	DR float32
	NBuffer int32
}

func NewBolshoiPBuffer(
	path string, orderFlag string, context Context,
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
	
	f, err := os.Open(path)
	if err != nil { return nil, err }
	defer f.Close()

	buf := &BolshoiPBuffer{ order: order }
	comment := [45]byte{}

	fortranRead(f, order, &comment)
	fortranRead(f, order, &buf.bh1)
	fortranRead(f, order, &buf.bh2)
	
	return buf, nil
}

func (bol *BolshoiPBuffer) Read(fname string) (
	xs, vs [][3]float32, ms []float32, ids []int64, err error,
) {
	f, err := os.Open(fname)
	if err != nil { return nil, nil, nil, nil, err }
	defer f.Close()
	
	comment := [45]byte{}
	bh1 := bolshoiPHeader1{}
	bh2 := bolshoiPHeader2{}
	boundary := [6]float32{}
	var nPart, m int32
	
	fortranRead(f, bol.order, &comment)
	fortranRead(f, bol.order, &bh1)
	fortranRead(f, bol.order, &bh2)
	fortranRead(f, bol.order, &boundary)
	fortranRead(f, bol.order, &nPart)

	for i := 0; i < 3; i++ {
		bol.xReadBuf[i] = expandScalars(bol.xReadBuf[i], int(nPart))
		bol.vReadBuf[i] = expandScalars(bol.vReadBuf[i], int(nPart))
	}
	
	bol.xsBuf = expandVectors(bol.xsBuf, int(nPart))
	bol.vsBuf = expandVectors(bol.vsBuf, int(nPart))
	bol.idsBuf = expandInts(bol.idsBuf, int(nPart))
	bol.msBuf = expandScalars(bol.msBuf, int(nPart))
	bol.wpBuf = expandScalars(bol.wpBuf, int(nPart))

	for nRead := int32(0); nRead < nPart; {
		fortranRead(f, bol.order, &m)

		_ = readInt32(f, bol.order)
		for i := 0; i < 3; i++ {
			readFloat32AsByte(f, bol.order, bol.xReadBuf[i][nRead: nRead+m])
		}
		_ = readInt32(f, bol.order)
		
		_ = readInt32(f, bol.order)
		for i := 0; i < 3; i++ {
			readFloat32AsByte(f, bol.order, bol.vReadBuf[i][nRead: nRead+m])
		}
		_ = readInt32(f, bol.order)

		_ = readInt32(f, bol.order)
		readFloat32AsByte(f, bol.order, bol.wpBuf[nRead: nRead+m])
		readInt64AsByte(f, bol.order, bol.idsBuf[nRead: nRead+m])
		_ = readInt32(f, bol.order)
		
		nRead += m
	}

	hd := &Header{}
	err = bol.ReadHeader(fname, hd)
	if err != nil { return nil, nil, nil, nil, err }

	mp := bh1.Mass
	scale := float32(1/(1 + hd.Cosmo.Z))
	
	j := 0	
	for i := range bol.xsBuf {
		bol.xsBuf[i] = [3]float32{
			bol.xReadBuf[0][i], bol.xReadBuf[1][i], bol.xReadBuf[2][i],
		}
		bol.vsBuf[i] = [3]float32{
			bol.vReadBuf[0][i], bol.vReadBuf[1][i], bol.vReadBuf[2][i],
		}
		
		if !vecInRange(bol.xsBuf[i], hd) { continue }
		
		bol.xsBuf[j] = bol.xsBuf[i]
		
		bol.vsBuf[j][0] = scale*bol.vsBuf[i][0]
		bol.vsBuf[j][1] = scale*bol.vsBuf[i][1]
		bol.vsBuf[j][2] = scale*bol.vsBuf[i][2]

		bol.idsBuf[j] = bol.idsBuf[i]
		
		bol.msBuf[j] = mp
		
		j++
	}
	
	bol.xsBuf, bol.vsBuf = bol.xsBuf[:j], bol.vsBuf[:j]
	bol.msBuf, bol.idsBuf = bol.msBuf[:j], bol.idsBuf[:j]
	
	return bol.xsBuf, bol.vsBuf, bol.msBuf, bol.idsBuf, nil
}


func vecInRange(v [3]float32, hd *Header) bool {
	return v[0] > hd.Origin[0] &&
		v[1] > hd.Origin[1] &&
		v[2] > hd.Origin[2] &&
		v[0] <= hd.Origin[0] + hd.Width[0] &&
		v[1] <= hd.Origin[1] + hd.Width[1] &&
		v[2] <= hd.Origin[2] + hd.Width[2]
}

func (bol *BolshoiPBuffer)  Close() { }

func (bol *BolshoiPBuffer) IsOpen() bool { return false }

func (bol *BolshoiPBuffer) ReadHeader(fname string, out *Header) error {
	f, err := os.Open(fname)
	if err != nil { return err }
	defer f.Close()
	
	comment := [45]byte{}
	bh1 := bolshoiPHeader1{}
	bh2 := bolshoiPHeader2{}
	boundary := [6]float32{}
	var nPart int32
	
	fortranRead(f, bol.order, &comment)
	fortranRead(f, bol.order, &bh1)
	fortranRead(f, bol.order, &bh2)
	fortranRead(f, bol.order, &boundary)
	fortranRead(f, bol.order, &nPart)

	cosmo := CosmologyHeader{
		OmegaM: float64(bh1.OmegaM),
		OmegaL: float64(bh1.OmegaL),
		H100: float64(bh1.H100),
		Z: float64(1/bh1.AExpn - 1),
	}

	ix := float32((bh2.K - 1) % bh2.Nx)
	iy := float32(((bh2.K - 1) / bh2.Nx) % bh2.Ny)
	iz := float32((bh2.K - 1) / (bh2.Nx*bh2.Ny))
	width := [3]float32{
		bh1.BoxWidth / float32(bh2.Nx),
		bh1.BoxWidth / float32(bh2.Ny),
		bh1.BoxWidth / float32(bh2.Nz),
	}
	origin := [3]float32{ ix * width[0], iy * width[1], iz * width[2] }
	
	hd := Header{
		Cosmo: cosmo,
		N: int64(nPart),
		TotalWidth: float64(bh1.BoxWidth),
		Origin: origin,
		Width: width,
	}

	*out = hd
	return nil
}

func (bol *BolshoiPBuffer) MinMass() float32 {
	boxMult := float64(bol.bh1.BoxWidth) / 250.0
	nTot, _ := bol.TotalParticles("")
	nMult := (2048*2048*2048) / float64(nTot)
	return float32(1.35e8*nMult*boxMult*boxMult*boxMult)
}

// TODO: is there any way to figure out if this number changed? I don't think
// so.
func (bol *BolshoiPBuffer) TotalParticles(fname string) (int, error) {
	return 2048*2048*2048, nil
}
