package sort


// Percentile calculates the element corresponding to the percentile, p,
// of a non-empty slice. p must be in the range [0, 1]. An optional buffer
// slice of the same size may be supplied to prevent unneeded heap allocations.
// Runs in O(len(xs))
func Percentile(xs []float64, p float64, buf ...[]float64) float64 {
	if len(xs) == 0 {
		panic("xs empty in call to Precentile(xs, ps)")
	} else if p > 1 || p < 0 {
		panic("percentile must be in the range [0, 1]")
	}

	n := int(p * float64(len(xs)))
	// 1-indexing
	if n == 0 {
		n++
	}
	
	return NthLargest(xs, n, buf...)
}

// Median calculates the median of a non-empty slice. An optional buffer
// slice of the same size may be supplied to prevent unneeded heap allocations.
// Runs in O(len(xs)).
func Median(xs []float64, buf ...[]float64) float64 {
	if len(xs) == 0 {
		panic("xs empty in call to Median(xs)")
	}

	return NthLargest(xs, len(xs)/2, buf...)
}

// Remember that this function is 1-indexed.
// NthLargest the nth largest element of a non-empty slice, xs. An optional
// buffer slice of the same size may be supplied  to prevent unneeded heap
// allocations. Runs in O(len(xs)).
//
// n is 1-indexed, meaning that n=2 (not n=1) corresponds to the second largest
// element.
func NthLargest(xs []float64, n int, buf ...[]float64) float64 {
	if len(xs) == 0 {
		panic("xs empty in call to NthHighest(xs, ns)")
	}

	var medSlice []float64
	if len(buf) == 0 {
		medSlice = make([]float64, len(xs))
	} else {
		medSlice = buf[0]
		if len(medSlice) != len(xs) {
			panic("Length of buffer does not equal length of input array.")
		}
	}
	copy(medSlice, xs)

	// This function is just a wrapper around the less user-friendly
	// nthLargest.
	return nthLargest(medSlice, n)
}

// nthLargest is a helper function which recursively calculates the nth
// largest element of the slice xs. It is essentially the same as quicksort
// except that at each level of recursion, one of the two partition halves
// is discarded.
func nthLargest(xs []float64, n int) float64 {
	switch len(xs) {
	case 1:
		return xs[0]
	case 2:
		if xs[0] > xs[1] {
			xs[0], xs[1] = xs[1], xs[0]
		}

		if n <= 2 {
			return xs[2-n]
		} else {
			panic("n in nthLargest(xs, n) too large")
		}
	case 3:
		high, mid, low := sort3(xs[0], xs[1], xs[2])
		switch n {
		case 1:
			return high
		case 2:
			return mid
		case 3:
			return low
		default:
			panic("n in nthLargest(xs, n) too large or too small")
		}
	default:
		pivIdx := partition(xs)
		nPiv := len(xs) - pivIdx
		if nPiv > n {
			return nthLargest(xs[pivIdx:], n)
		} else if nPiv < n {
			return nthLargest(xs[:pivIdx], n-nPiv)
		} else { // nPiv == n
			return xs[pivIdx]
		}
	}
}
