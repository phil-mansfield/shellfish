package io

type VectorBuffer interface {
	Read(fname string) ([][3]float32, error)
	Close() error
	IsOpen() bool
}

type GotetraBuffer struct {
	open bool

	sheet [][3]float32
	out [][3]float32

	sw, gw int

	hd SheetHeader
}

func NewGotetraBuffer(fname string) *GotetraBuffer {
	hd := &SheetHeader{}
	ReadSheetHeaderAt(fname, hd)

	sw, gw := hd.segmentWidth, hd.gridWidth
	buf := &GotetraBuffer{
		sheet: make([][3]float32, gw * gw * gw),
		out: make([][3]float32, sw * sw * sw),
		open: false,
		sw: int(sw), gw: int(gw),
	}

	return buf
}

func (buf *GotetraBuffer) IsOpen() bool { return buf.open }

func (buf *GotetraBuffer) Read(fname string) ([][3]float32, error) {
	if buf.open { panic("Buffer already open.") }

	err := ReadSheetPositionsAt(fname, buf.sheet)
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