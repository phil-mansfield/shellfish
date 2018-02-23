package halo

// SplitSubhaloFinder finds the halos in group A that are within the radii of
// group B. It is optimized for the case where A is extremely large and search
// radii are also very large.
//
// To do this, it does not allow for host list identifications and does not
// memoize results.
type SplitSubhaloFinder struct {
	g         *Grid
	gBuf      []int
	idxBuf []int
	dr2Buf  []float64
	xs, ys, zs []float64
	bufi int
}



// NewSubhaloFinder creates a new subhalo finder corresponding to the given
// Grid. The Grid contains halos from group A.
func NewSplitSubhaloFinder(g *Grid, xs, ys, zs []float64) *SplitSubhaloFinder {
	i := &SplitSubhaloFinder{
		g: g,
		gBuf: make([]int, g.MaxLength()),
		idxBuf: make([]int, len(g.Next)),
		dr2Buf: make([]float64, len(g.Next)),
		xs: xs, ys: ys, zs: zs,
	}

	return i
}



// FindSubhalos links grid halos (from group A) to a target halo (from group B).
// Returned arrays are internal buffers, so please treat them kindly.
func (sf *SplitSubhaloFinder) FindSubhalos(xh, yh, zh, rh float64) (
	idx []int, dr2 []float64,
) {
	sf.bufi = 0
	sf.idxBuf = sf.idxBuf[:cap(sf.idxBuf)]
	sf.dr2Buf = sf.dr2Buf[:cap(sf.dr2Buf)]

	b := &Bounds{}
	c := sf.g.Cells

	pos := [3]float64{xh, yh, zh}
	b.SphereBounds(pos, rh, sf.g.cw, sf.g.Width)

	for dz := 0; dz < b.Span[2]; dz++ {
		z := b.Origin[2] + dz
		if z >= c {
			z -= c
		}
		zOff := z * c * c
		for dy := 0; dy < b.Span[1]; dy++ {
			y := b.Origin[1] + dy
			if y >= c {
				y -= c
			}
			yOff := y * c
			for dx := 0; dx < b.Span[0]; dx++ {
				x := b.Origin[0] + dx
				if x >= c {
					x -= c
				}
				idx := zOff + yOff + x

				sf.gBuf = sf.g.ReadIndexes(idx, sf.gBuf)
				sf.addSubhalos(sf.gBuf, xh, yh, zh, rh, sf.g.Width)
			}
		}
	}

	return sf.idxBuf[:sf.bufi], sf.dr2Buf[:sf.bufi]
}

func (sf *SplitSubhaloFinder) addSubhalos(
	idxs []int, xh, yh, zh, rh float64, L float64,
) {
	for _, j := range idxs {
		sx, sy, sz := sf.xs[j], sf.ys[j], sf.zs[j]
		dx, dy, dz, dr := xh-sx, yh-sy, zh-sz, rh

		if dx > +L/2 { dx -= L }
		if dx < -L/2 { dx += L }
		if dy > +L/2 { dy -= L }
		if dy < -L/2 { dy += L }
		if dz > +L/2 { dz -= L }
		if dz < -L/2 { dz += L }

		dr2 := dx*dx + dy*dy + dz*dz

		if dr*dr >= dr2 {
			sf.idxBuf[sf.bufi] = j
			sf.dr2Buf[sf.bufi] = dr2
			sf.bufi++
		}
	}
}