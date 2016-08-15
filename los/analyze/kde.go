package analyze

import (
	"math"

	plt "github.com/phil-mansfield/pyplot"
	intr "github.com/phil-mansfield/shellfish/math/interpolate"
)

// GaussianKDE calculates a gaussian KDE from the points in the array
// xs, which kernel widths of h. This is done by evaluating the KDE
// at n uniformly-spaced points between low and high and then returning a
// cubic Spline fit through these points.
func GaussianKDE(xs []float64, h, low, high float64, n int) *intr.Spline {
	dx := (high - low) / float64(n-1)
	spXs, spYs := make([]float64, n), make([]float64, n)
	for i := 0; i < n-1; i++ {
		spXs[i] = low + dx*float64(i)
	}
	spXs[n-1] = high

	maxDist := h * 3

	for _, x := range xs {
		lowIdx := int((x - maxDist - low) / dx)
		highIdx := int((x+maxDist-low)/dx) + 1
		if lowIdx < 0 {
			lowIdx = 0
		}
		if highIdx >= n {
			highIdx = n - 1
		}
		for i := lowIdx; i <= highIdx; i++ {
			udx := (spXs[i] - x) / h
			spYs[i] += math.Exp(-udx * udx)
		}
	}

	return intr.NewSpline(spXs, spYs)
}

// KDETree is a data structure that maintains all the state required for
// the recursive KDE-based filtering used by Shellfish.
type KDETree struct {
	// h - width of KDE
	// low - minimum radius
	// high - maximum radius
	h, low, high      float64
	// spTree[i][j] - KDE in the jth bin of the ith level of recursion
	spTree            [][]*intr.Spline
	// maxesTree[i][j] - location of each local maxima
	maxesTree         [][][]float64
	// thTree[i][j] - angle of the jth bin at the ith level of recursion
	// connMaxes[i][j] - the radius of KDE maximum in the jth bin of the ith
	//                   level of recursion (after the connection process has
	//                   been used).
	thTree, connMaxes [][]float64
	// spRs - Slice of radial bins shared by every
	spRs              []float64
}

// NewKDETree performs filtering fo points using the recursive KDE-based
// filtering described in section 2.2.3 of Mansfield, Kravtsov, Diemer (2016).
//
// splits is the number of levels of recursion minus one.
func NewKDETree(
	rs, phis []float64, splits int, h float64,
) (*KDETree, bool) {
	kt := new(KDETree)
	rn := 100

	if len(rs) == 0 {
		panic("No input r and phi seuqences to NewKDETree. This can " +
			"sometimes happen if incorrect halo positions are given (e.g. " +
			"using a halo catalog from the wrong simulation suite, mistyping " +
			"a coordinate, etc). " +
			"If you are _sure_ that your input locations correspond to " +
			"actual halo centers, this might also be an internal Shellfish " +
			"error and you should submit a bug report.")
	}

	kt.low, kt.high = 0, rs[0]
	for _, r := range rs {
		if r > kt.high {
			kt.high = r
		}
	}

	kt.spRs = make([]float64, rn)
	dr := (kt.high - kt.low) / float64(rn)
	for i := range kt.spRs {
		kt.spRs[i] = kt.low + dr*float64(i)
	}
	kt.spRs[rn-1] = kt.high

	kt.h = h

	kt.spTree = [][]*intr.Spline{{GaussianKDE(rs, kt.h, kt.low, kt.high, 100)}}
	kt.thTree = [][]float64{{math.Pi}}

	kt.growTrees(rs, phis, splits)
	kt.findMaxes()
	kt.connectMaxes()

	return kt, true
}

// PlotLevel plots the KDEs found at a particular level of a KDETree using the
// pyplot options, opts. It is used purely for debugging purposed.
func (kt *KDETree) PlotLevel(level, spIdx int, opts ...interface{}) {
	sps := kt.spTree[level]
	rs := make([]float64, 100)
	vals := make([]float64, 100)
	dr := (kt.high - kt.low) / float64(len(rs))
	for i := range rs {
		rs[i] = dr*(float64(i)+0.5) + kt.low
	}

	if spIdx < -1 {
		for _, sp := range sps {
			for j := range vals {
				vals[j] = sp.Eval(rs[j])
			}
			args := append([]interface{}{rs, vals}, opts...)
			plt.Plot(args...)
		}
	} else {
		sp := sps[spIdx]
		for j := range vals {
			vals[j] = sp.Eval(rs[j])
		}
		args := append([]interface{}{rs, vals}, opts...)
		plt.Plot(args...)

	}
}

// growTrees populated kt.thTree and kt.spTree with angles and splines,
// respectively.
func (kt *KDETree) growTrees(rs, phis []float64, splits int) {
	for split := 0; split < splits; split++ {
		bins := int(1 << uint((1 + split)))
		rBins, thBins := binByTheta(rs, phis, bins)
		sps := make([]*intr.Spline, bins)

		for i, rBin := range rBins {
			sps[i] = GaussianKDE(rBin, kt.h, kt.low, kt.high, 100)
		}
		kt.thTree = append(kt.thTree, thBins)
		kt.spTree = append(kt.spTree, sps)
	}
}

// binByTheta breaks a group of (theta, r) coordinates into groups based on
// their angular bin.
func binByTheta(
	rs, ths []float64, bins int,
) (rBins [][]float64, thBins []float64) {
	rBins = make([][]float64, bins)
	for i := range rBins {
		rBins[i] = []float64{}
	}
	dth := (2 * math.Pi) / float64(bins)
	for i := range rs {
		idx := int(ths[i] / dth)
		rBins[idx] = append(rBins[idx], rs[i])
	}

	thBins = make([]float64, bins)
	for i := range thBins {
		thBins[i] = 2 * math.Pi * (float64(i) + 0.5) / float64(bins)
	}

	return rBins, thBins
}

// findMaxes populated kt.maxesTree with local maxima.
func (kt *KDETree) findMaxes() {
	kt.maxesTree = [][][]float64{}
	for _, sps := range kt.spTree {
		levelMaxes := make([][]float64, len(sps))
		for j, sp := range sps {
			maxes := localSplineMaxes(kt.spRs, sp)
			levelMaxes[j] = maxes
		}

		kt.maxesTree = append(kt.maxesTree, levelMaxes)
	}
}

// localSplineMaxes returns the radii of every local maximum in a spline when
// evaluated at the the given input points.
func localSplineMaxes(xs []float64, sp *intr.Spline) []float64 {
	prev, curr, next := sp.Eval(xs[0]), sp.Eval(xs[1]), sp.Eval(xs[2])
	maxes := []float64{}
	if curr > next && curr > prev {
		maxes = append(maxes, xs[1])
	}
	for i := 2; i < len(xs)-1; i++ {
		prev, curr, next = curr, next, sp.Eval(xs[i+1])
		if curr > next && curr > prev {
			maxes = append(maxes, xs[i])
		}
	}
	return maxes
}

// connectMaxes is the code that actually creates the filtering curve from
// section 2.2.3 in Mansfield+ (2016). The explanation is a bit involved, so
// I'm just going to reference that section.
func (kt *KDETree) connectMaxes() {
	kt.connMaxes = [][]float64{{kt.maxesTree[0][0][0]}}

	for split, maxes := range kt.maxesTree[1:] {
		prevMaxes := kt.connMaxes[len(kt.connMaxes)-1]
		currMaxes := make([]float64, 2*len(prevMaxes))
		for i := range currMaxes {
			currMaxes[i] = math.NaN()
		}

		for node := range maxes {
			nodeMaxes := maxes[node]
			if len(maxes) == 0 {
				continue
			}
			nodePrevMax := prevMaxes[node/2]

			var connMax float64
			if math.IsNaN(nodePrevMax) || len(nodeMaxes) == 0 {
				connMax = math.NaN()
			} else {
				connIdx, minDist := -1, math.Inf(+1)
				for i := range nodeMaxes {
					dist := math.Abs(nodePrevMax - nodeMaxes[i])
					if dist < minDist {
						connIdx, minDist = i, dist
					}
				}
				connMax = nodeMaxes[connIdx]
			}

			if !math.IsNaN(connMax) && (split == 0 ||
				math.Abs(connMax-nodePrevMax) < kt.h) {
				currMaxes[node] = connMax
			} else {
				for _, max := range nodeMaxes {
					rFunc := kt.GetRFunc(split, Radial)
					spR := rFunc(kt.thTree[split+1][node])
					if math.Abs(max-spR) < kt.h {
						currMaxes[node] = max
					}
				}
			}
		}

		kt.connMaxes = append(kt.connMaxes, currMaxes)
	}
}

// getFinestMax returns the connected maximum which corresponds to the finest
// level of returns in the given bin and level of recursion.
func (kt *KDETree) getFinestMax(idx, level int) float64 {
	for i := 0; i <= level; i++ {
		r := kt.connMaxes[level-i][idx/(1<<uint(i))]
		if !math.IsNaN(r) {
			return r
		}
	}
	panic(":3")
}

// GetConMaxes returns the location of the anchor points at a given level of
// recursion.
func (kt *KDETree) GetConnMaxes(level int) (rs, ths []float64) {
	ths = kt.thTree[level]
	maxes := kt.connMaxes[level]
	retMaxes := make([]float64, len(maxes))
	for i := range maxes {
		retMaxes[i] = kt.getFinestMax(i, level)
	}
	return retMaxes, ths
}

// extendAngularRange extends a series of (r, theta) points through several
// 2 pi repetitions. This is done so that a spline through these points will
// not encounter boundary effects when evaluated in [0, 2 pi].
func extendAngularRange(maxes, ths []float64) (spMaxes, spThs []float64) {
	n := len(maxes)
	buf := 5
	if buf > n {
		buf = n
	}
	spThs, spMaxes = make([]float64, 2*buf+n), make([]float64, 2*buf+n)

	j := n - buf
	for i := 0; i < buf; i++ {
		spThs[i], spMaxes[i] = ths[j]-2*math.Pi, maxes[j]
		j++
	}
	j = 0
	for i := buf; i < n+buf; i++ {
		spThs[i], spMaxes[i] = ths[j], maxes[j]
		j++
	}
	j = 0
	for i := n + buf; i < n+2*buf; i++ {
		spThs[i], spMaxes[i] = ths[j]+2*math.Pi, maxes[j]
		j++
	}
	return spMaxes, spThs
}

type RFuncType int

const (
	Radial RFuncType = iota
	Cartesian
)

// GetRFunc returns the filtering curve at a particular level of recursion.
// rt  should always be set to Cartesian.
func (kt *KDETree) GetRFunc(level int, rt RFuncType) func(float64) float64 {
	switch rt {
	case Radial:
		maxes, ths := kt.GetConnMaxes(level)
		spMaxes, spThs := extendAngularRange(maxes, ths)
		sp := intr.NewSpline(spThs, spMaxes)
		return sp.Eval
	case Cartesian:
		maxes, ths := kt.GetConnMaxes(level)
		spMaxes, spThs := extendAngularRange(maxes, ths)
		spXs, spYs := make([]float64, len(spThs)), make([]float64, len(spThs))
		for i, th := range spThs {
			sin, cos := math.Sincos(th)
			spXs[i], spYs[i] = spMaxes[i]*cos, spMaxes[i]*sin
		}
		xSp, ySp := intr.NewSpline(spThs, spXs), intr.NewSpline(spThs, spYs)
		return func(th float64) float64 {
			x, y := xSp.Eval(th), ySp.Eval(th)
			return math.Sqrt(x*x + y*y)
		}
	default:
		panic(":3")
	}
}

// H Returns the smoothing scale of the KDE tree.
func (kt *KDETree) H() float64 { return kt.h }

// FilterNearby returns all the points which are within dr of the filtering
// curve at the specified level of recursion.
func (kt *KDETree) FilterNearby(
	rs, ths []float64, level int, dr float64,
) (fRs, fThs []float64, idxs []int) {
	//rFunc := kt.GetRFunc(level, Cartesian)
	rFunc := kt.GetRFunc(level, Radial)
	fRs, fThs, idxs = []float64{}, []float64{}, []int{}
	for i := range rs {
		if math.Abs(rFunc(ths[i])-rs[i]) < dr/2 {
			fRs = append(fRs, rs[i])
			fThs = append(fThs, ths[i])
			idxs = append(idxs, i)
		}
	}

	return fRs, fThs, idxs
}
