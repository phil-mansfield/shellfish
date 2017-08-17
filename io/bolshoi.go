package io

import (
	"encoding/binary"
	"os"
	"io"
	"fmt"
)

// Unfortunately, the Bolshoi boundary region is too small for us to get away
// with only needing to read a single segment file. Thus, we must go through
// the whole song-and-dance

type BolshoiBuffer struct {
	order binary.ByteOrder
	bh1 bolshoiHeader1
	bh2 bolshoiHeader2

	xsBuf, vsBuf [][3]float32
	msBuf []float32
	idsBuf []int64
	pBuf []bolshoiParticle
}

type bolshoiHeader1 struct {
	AExpn, AStep float32
	IStep int32
	NRowCells, NGridCells int32
	NSpecies, NSeed int32
	OmegaM, OmegaL, H100, BoxWidth float32
}

type bolshoiHeader2 struct {
	K int32
	Nx, Ny, Nz int32
	DR float32
}

type bolshoiParticle struct {
	X [3]float32
	V [3]float32
	ID int64
}

func NewBolshoiBuffer(
	path string, order binary.ByteOrder, context Context,
) (VectorBuffer, error) {

	f, err := os.Open(path)
	if err != nil { return nil, err }
	defer f.Close()

	buf := &BolshoiBuffer{ order: order }
	comment := [45]byte{}
	
	fortranRead(f, order, &comment)
	fortranRead(f, order, &buf.bh1)
	fortranRead(f, order, &buf.bh2)
	
	return buf, nil
}

func (bol *BolshoiBuffer) Read(fname string) (
	xs, vs [][3]float32, ms []float32, ids []int64, err error,
) {
	f, err := os.Open(fname)
	if err != nil { return nil, nil, nil, nil, err }
	defer f.Close()
	
	comment := [45]byte{}
	bh1 := bolshoiHeader1{}
	bh2 := bolshoiHeader2{}
	boundary := [6]float32{}
	var nPart, m int32
	
	fortranRead(f, bol.order, &comment)
	fortranRead(f, bol.order, &bh1)
	fortranRead(f, bol.order, &bh2)
	fortranRead(f, bol.order, &boundary)
	fortranRead(f, bol.order, &nPart)

	bol.pBuf = expandBolshoiParticles(bol.pBuf[:0], int(nPart))
	
	for nRead := int32(0); nRead < nPart; {
		fortranRead(f, bol.order, &m)
		//fortranRead(f, bol.order, bol.pBuf[nRead: nRead+m])
		_ = readInt32(f, bol.order)
		readBolshoiParticleAsByte(f, bol.order, bol.pBuf[nRead: nRead+m])
		_ = readInt32(f, bol.order)
		nRead += m
	}

	hd := &Header{}
	err = bol.ReadHeader(fname, hd)
	if err != nil { return nil, nil, nil, nil, err }
	
	bol.pBuf = filterBolshoiParticles(bol.pBuf, hd)
	xs, vs, ms, ids = postprocessBolshoi(bol.pBuf, hd, bol)
	
	return xs, vs, ms, ids, nil
}

func expandBolshoiParticles(p []bolshoiParticle, n int) []bolshoiParticle {
	switch {
	case cap(p) >= n:
		return p[:n]
	case int(float64(cap(p))*1.5) > n:
		return append(p[:cap(p)], make([]bolshoiParticle, n-cap(p))...)
	default:
		return make([]bolshoiParticle, n)
	}
}

func filterBolshoiParticles(p []bolshoiParticle, hd *Header) []bolshoiParticle {
	j := 0
	for _, pp := range p {
		if bolshoiInRange(pp, hd) {
			p[j] = pp
			j++
		}
	}
	return p[:j]
}

func bolshoiInRange(p bolshoiParticle, hd *Header) bool {
	return p.X[0] > hd.Origin[0] &&
		p.X[1] > hd.Origin[1] &&
		p.X[2] > hd.Origin[2] &&
		p.X[0] <= hd.Origin[0] + hd.Width[0] &&
		p.X[1] <= hd.Origin[1] + hd.Width[1] &&
		p.X[2] <= hd.Origin[2] + hd.Width[2]
}

func postprocessBolshoi(p []bolshoiParticle, hd *Header, buf *BolshoiBuffer) (
	x, v [][3]float32, m []float32, id []int64,
) {
	buf.xsBuf = expandVectors(buf.xsBuf[:0], len(p))
	buf.vsBuf = expandVectors(buf.vsBuf[:0], len(p))
	buf.msBuf = expandScalars(buf.msBuf[:0], len(p))
	buf.idsBuf = expandInts(buf.idsBuf[:0], len(p))

	scale := float32(1/(1 + hd.Cosmo.Z))
	mp := buf.MinMass()
	
	for i, pp := range p {
		buf.xsBuf[i] = pp.X
		buf.vsBuf[i] = [3]float32{pp.V[0]*scale, pp.V[1]*scale, pp.V[2]*scale }
		buf.idsBuf[i] = pp.ID
		buf.msBuf[i] = mp
	}
	
	return buf.xsBuf, buf.vsBuf, buf.msBuf, buf.idsBuf
}

func (bol *BolshoiBuffer)  Close() { }

func (bol *BolshoiBuffer) IsOpen() bool { return false }

func fortranRead(rd io.Reader, order binary.ByteOrder, buf interface{}) {
	size1 := readInt32(rd, order)
	err := binary.Read(rd, order, buf)
	if err != nil { panic(err.Error()) }
	size2 := readInt32(rd, order)

	if size1 != size2 {
		panic(fmt.Sprintf(
			"Fortran binary header is %d, but footer is %d.", size1, size2,
		))
	}
}

func (bol *BolshoiBuffer) ReadHeader(fname string, out *Header) error {
	f, err := os.Open(fname)
	if err != nil { return err }
	defer f.Close()
	
	comment := [45]byte{}
	bh1 := bolshoiHeader1{}
	bh2 := bolshoiHeader2{}
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

func (bol *BolshoiBuffer) MinMass() float32 {
	boxMult := float64(bol.bh1.BoxWidth) / 250.0
	nTot, _ := bol.TotalParticles("")
	nMult := (2048*2048*2048) / float64(nTot)
	return float32(1.35e8*nMult*boxMult*boxMult*boxMult)
}

// TODO: is there any way to figure out if this number changed? I don't think
// so.
func (bol *BolshoiBuffer) TotalParticles(fname string) (int, error) {
	return 2048*2048*2048, nil
}
