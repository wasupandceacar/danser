package aop

import (
	"danser/bmath"
	"fmt"
	"github.com/Mempler/rplpa"
	"github.com/cnkei/gospline"
	"math"
	"sort"
)

type xy struct{ x, y []float64 }

func (s *xy) Len() int           { return len(s.x) }
func (s *xy) Less(i, j int) bool { return s.x[i] < s.x[j] }
func (s *xy) Swap(i, j int) {
	s.x[i], s.x[j] = s.x[j], s.x[i]
	s.y[i], s.y[j] = s.y[j], s.y[i]
}
func (s *xy) XY(i int) (float64, float64) { return s.x[i], s.y[i] }

func interpolate2points(x0, y0, x1, y1 float64) func(float64) float64 {
	return func(x float64) float64 {
		if x1 == x0 {
			return (y0 + y1) / 2.
		}
		return y0 + (x-x0)/(x1-x0)*(y1-y0)
	}
}

func Interp1d(x, y []float64) func(x float64) float64 {
	if len(x) < 2 || len(x) != len(y) {
		panic(fmt.Errorf("interp1d lenx:%d leny:%d", len(x), len(y)))
	}

	var both *xy
	if sort.Float64sAreSorted(x) {
		both = &xy{x, y}
	} else {
		both = &xy{make([]float64, len(x)), make([]float64, len(x))}
		copy(both.x, x)
		copy(both.y, y)
		sort.Sort(both)
	}
	return func(x float64) float64 {
		ix := sort.SearchFloat64s(both.x, x) - 1
		if ix < 0 {
			ix = 0
		}
		if ix > len(both.x)-2 {
			ix = len(both.x) - 2
		}
		ix1 := ix + 1
		for ix > 0 && math.IsNaN(both.y[ix]) {
			ix--
		}
		for ix1 < len(both.x)-1 && math.IsNaN(both.y[ix1]) {
			ix1++
		}
		//fmt.Println(x, both.x, both.y, ix, ix1)
		return interpolate2points(both.x[ix], both.y[ix], both.x[ix1], both.y[ix1])(x)
	}
}

func interpolateCubicSpline(rep *rplpa.Replay) func(t int64) bmath.Vector2d {
	var ts []float64
	var xs []float64
	var ys []float64
	if len(rep.ReplayData) > 2 {
		ts = append(ts, float64(rep.ReplayData[0].Time))
		ts[0] += float64(rep.ReplayData[1].Time)
		ts[0] += float64(rep.ReplayData[2].Time)
		xs = append(xs, float64(rep.ReplayData[2].MosueX))
		ys = append(ys, float64(rep.ReplayData[2].MouseY))
	}
	for i := 3; i < len(rep.ReplayData)-1; i++ {
		if rep.ReplayData[i].Time == 0 {
			continue
		}
		ts = append(ts, float64(rep.ReplayData[i].Time))
		xs = append(xs, float64(rep.ReplayData[i].MosueX))
		ys = append(ys, float64(rep.ReplayData[i].MouseY))
	}
	for i := 1; i < len(ts); i++ {
		ts[i] += ts[i-1]
	}
	fx := gospline.NewCubicSpline(ts, xs)
	fy := gospline.NewCubicSpline(ts, ys)
	return func(t int64) bmath.Vector2d {
		ft := float64(t)
		return bmath.Vector2d{
			X: fx.At(ft),
			Y: fy.At(ft),
		}
	}
}

func interpolateLinear(rep *rplpa.Replay) func(t int64) bmath.Vector2d {
	var ts []float64
	var xs []float64
	var ys []float64
	if len(rep.ReplayData) > 2 {
		ts = append(ts, float64(rep.ReplayData[0].Time))
		ts[0] += float64(rep.ReplayData[1].Time)
		ts[0] += float64(rep.ReplayData[2].Time)
		xs = append(xs, float64(rep.ReplayData[2].MosueX))
		ys = append(ys, float64(rep.ReplayData[2].MouseY))
	}
	for i := 3; i < len(rep.ReplayData)-1; i++ {
		if rep.ReplayData[i].Time == 0 {
			continue
		}
		ts = append(ts, float64(rep.ReplayData[i].Time))
		xs = append(xs, float64(rep.ReplayData[i].MosueX))
		ys = append(ys, float64(rep.ReplayData[i].MouseY))
	}
	for i := 1; i < len(ts); i++ {
		ts[i] += ts[i-1]
	}
	fx := Interp1d(ts, xs)
	fy := Interp1d(ts, ys)
	return func(t int64) bmath.Vector2d {
		ft := float64(t)
		return bmath.Vector2d{
			X: fx(ft),
			Y: fy(ft),
		}
	}
}
