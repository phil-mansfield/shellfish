/*package interpolate implements interpolators in various dimensions as
well as filters.
*/
package interpolate

// Interpolator is a 1D interpolator. These interpolators all use caching, so
// they are not thread safe.
type Interpolator interface {
	// Eval evaluates the interpolator at x.
	Eval(x float64) float64
	// EvalAll evaluates a sequeunce of values and returns the result. An
	// optional output array can be supplied to prevent unneeded heap
	// allocations.
	EvalAll(xs []float64, out ...[]float64) []float64
	// Ref creates a shallow copy of the interpolator with its own cache.
	// Each thread using the same interpolator must make a copy with Ref
	// first.
	Ref() Interpolator
}

var (
	_ Interpolator = &Spline{}
	_ Interpolator = &Linear{}
	_ Interpolator = &splineRef{}
	_ Interpolator = &linearRef{}
)

// BiInterpolator is a 2D interpolator. These interpolators all use caching, so
// they are not thread safe.
type BiInterpolator interface {
	// Eval evaluates the interpolator at a point.
	Eval(x, y float64) float64
	// EvalAll evaluates a sequeunce of points and returns the result. An
	// optional output array can be supplied to prevent unneeded heap
	// allocations.
	EvalAll(xs, ys []float64, out ...[]float64) []float64
	// Ref creates a shallow copy of the interpolator with its own cache.
	// Each thread using the saem interpolator must make a copy with Ref
	// first.
	Ref() BiInterpolator
}

var (
	_ BiInterpolator = &BiLinear{}
	_ BiInterpolator = &BiCubic{}
	_ BiInterpolator = &biLinearRef{}
	_ BiInterpolator = &biCubicRef{}
)

// BiInterpolator is a 2D interpolator. These interpolators all use caching, so
// they are not thread safe.
type TriInterpolator interface {
	// Eval evaluates the interpolator at a point.
	Eval(x, y, z float64) float64
	// EvalAll evaluates a sequeunce of points and returns the result. An
	// optional output array can be supplied to prevent unneeded heap
	// allocations.
	EvalAll(xs, ys, zs []float64, out ...[]float64) []float64
	// Ref creates a shallow copy of the interpolator with its own cache.
	// Each thread using the saem interpolator must make a copy with Ref
	// first.
	Ref() TriInterpolator
}

var (
	_ TriInterpolator = &TriLinear{}
	_ TriInterpolator = &TriCubic{}
	_ TriInterpolator = &triLinearRef{}
	_ TriInterpolator = &triCubicRef{}
)
