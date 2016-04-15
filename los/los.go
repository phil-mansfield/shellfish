package los

import (
	"math"
	"runtime"

	"github.com/phil-mansfield/shellfish/render/io"
	"github.com/phil-mansfield/shellfish/los/geom"
)

type Buffers struct {
	xs [][3]float32
	ts []geom.Tetra
	ss []geom.Sphere
	rhos []float64
	intr []bool
	bufHs []HaloProfiles
	skip int64
}

func (buf *Buffers) ParallelDensity(h *HaloProfiles) {
	workers := runtime.NumCPU()
	out := make(chan int, workers)

	for id := 0; id < workers - 1; id++ {
		go buf.chanIntersect(h, id, workers, out)
	}
	buf.chanIntersect(h, workers - 1, workers, out)
	for i := 0; i < workers; i++ { <-out }

	if workers > len(h.rs) { workers = len(h.rs) }
	for id := 0; id < workers - 1; id++ {
		go buf.chanDensity(h, id, workers, out)
	}
	buf.chanDensity(h, workers - 1, workers, out)
	for i := 0; i < workers; i++ { <-out }
}

func (buf *Buffers) chanDensity(
	h *HaloProfiles, id, workers int, out chan <- int,
) {
	for ri := id; ri < len(h.rs); ri += workers {
		r := &h.rs[ri]
		for ti := 0; ti < len(buf.ts); ti++ {
			if math.IsNaN(buf.rhos[ti]) || math.IsInf(buf.rhos[ti], 0) {
				continue
			}

			if buf.intr[ti] { r.Density(&buf.ts[ti], buf.rhos[ti]) }
		}
	}
	out <- id
}

func (buf *Buffers) chanIntersect(
	h *HaloProfiles, id, workers int, out chan <- int,
) {
	bufLen := len(buf.ts) / workers
	bufStart, bufEnd := id * bufLen, (id + 1) * bufLen
	if id == workers - 1 { bufEnd = len(buf.ts) }
	for i := bufStart; i < bufEnd; i++ {
		buf.intr[i] = h.Sphere.SphereIntersect(&buf.ss[i]) &&
			!h.minSphere.TetraContain(&buf.ts[i])
	}
	out <- id
}

func splits(intr []bool, workers int) (idxs []int, ok bool) {
	n := 0
	for _, ok := range intr {
		if ok { n++ }
	}

	if n == 0 { return nil, false }
	idxs = make([]int, workers + 1)

	spacing := n / workers
	// When m == spacing, insert a split at index j and reset m.
	m, j := 0, 1
	for i, ok := range intr {
		if ok {
			m++
			if m == spacing {
				idxs[j] = i + 1
				j++
				m = 0
				if j == workers { break }
			}
		}
	}

	for j = j; j <= workers ; j++ {
		idxs[j] = len(intr)
	}

	return idxs, true
}

// WrapHalo updates the coordinates of a slice of HaloProfiles so that they
// as close to the given sheet as periodic boundary conditions will allow.
func WrapHalo(hps []*HaloProfiles, hd *io.SheetHeader) {
	tw := float32(hd.TotalWidth)
	newC := &[3]float32{}
	for i := range hps {
		h := hps[i]
		for j := 0; j < 3; j++ {
			if h.cCopy[j] + h.R < hd.Origin[j] {
				newC[j] = h.cCopy[j] + tw
			} else {
				newC[j] = h.cCopy[j]
			}
		}
		h.ChangeCenter(newC)
	}
}
