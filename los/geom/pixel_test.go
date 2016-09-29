package geom

import(
	"fmt"
	"math"
	"math/rand"
	"testing"
	plt "github.com/phil-mansfield/pyplot"
)

func diskRand() (r, th float64) {
	return math.Sqrt(rand.Float64()), rand.Float64()*2*math.Pi
}

func TestDiskPixel(t *testing.T) {
	plt.Figure(plt.FigSize(8, 8))

	lvl := 3

	r, th := []float64{}, []float64{}
	x, y := []float64{}, []float64{}
	idx := []int{}
	for i := 0; i < 10*1000; i++ {
		rr, rth := diskRand()
		r, th = append(r, rr), append(th, rth)
		x, y = append(x, rr*math.Cos(rth)), append(y, rr*math.Sin(rth))
		idx = append(idx, DiskPixel(rr, rth, lvl))
	}
	
	for p := 0; p < DiskPixelNum(lvl); p++ {
		px, py := []float64{}, []float64{}
		for i := range idx {
			if idx[i] == p {
				px, py = append(px, x[i]), append(py, y[i])
			}
		}
		plt.Plot(px, py, ".", plt.C(fmt.Sprintf("%g", rand.Float64())))
	}

	plt.Show()
}