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

func sphereRand() (phi, th float64) {
	return 2*math.Pi*rand.Float64(), math.Acos(2*rand.Float64() - 1)
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

func TestSpherePixel(t *testing.T) {
	plt.Figure(plt.Num(0), plt.FigSize(8, 8))
	plt.Figure(plt.Num(1), plt.FigSize(8, 8))

	lvl := 10

	phi, th := []float64{}, []float64{}
	idx := []int{}
	for i := 0; i < 50*1000; i++ {
		rphi, rth := sphereRand()
		phi, th = append(phi, rphi), append(th, rth)
		idx = append(idx, SpherePixel(rphi, rth, lvl))
	}

	for p := 0; p < SpherePixelNum(lvl); p++ {
		pphi, pth := []float64{}, []float64{}
		for i := range idx {
			if idx[i] == p {
				pphi, pth = append(pphi, phi[i]), append(pth, th[i])
			}
		}
		plt.Figure(plt.Num(0))
		//c := fmt.Sprintf("%g", rand.Float64())
		c := []string{
			"r", "g", "b", "k", "w", "purple", "orange", "c", "y", "m",
		}[p%10]
		plt.Plot(pphi, pth, "o", plt.C(c))
		plt.Figure(plt.Num(1))
		py, pz := []float64{}, []float64{}
		for i := range pphi {
			x := math.Sin(pth[i]) * math.Cos(pphi[i])
			y := math.Sin(pth[i]) * math.Sin(pphi[i])
			z := math.Cos(pth[i])
			if x > 0 {
				py, pz = append(py, y), append(pz, z)
			}
		}
		plt.Plot(py, pz, "o", plt.C(c))
	}
	fmt.Println(SpherePixelNum(lvl))

	plt.Figure(plt.Num(0))
	plt.XLim(0, 2*math.Pi)
	plt.YLim(0, math.Pi)
	plt.Figure(plt.Num(1))
	plt.XLim(-1, +1)
	plt.YLim(-1, +1)

	plt.Show()
}