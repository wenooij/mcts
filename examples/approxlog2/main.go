package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/wenooij/mcts"
	"github.com/wenooij/mcts/model"
)

// best 1.3453431065831190, -0.3531000000000000
const (
	eps  = 1e-12
	c0Lo = 1.345342
	c0Hi = 1.345344
	c1Lo = -0.35309
	c1Hi = -0.35311
)

type constRange struct {
	lo float64
	hi float64
}

func (c constRange) Range() float64              { return c.hi - c.lo }
func (c constRange) Mid() float64                { return c.Range()/2 + c.lo }
func (c constRange) Sample(r *rand.Rand) float64 { return r.Float64() * c.Mid() }

func (c constRange) Minimized() bool { return c.Range() <= eps }

func (c constRange) Subranges() (lo, hi constRange, ok bool) {
	if c.Minimized() {
		return constRange{}, constRange{}, false
	}
	mid := c.Mid()
	return constRange{c.lo, mid}, constRange{mid, c.hi}, true
}
func (c constRange) String() string { return fmt.Sprintf("[%.16f,%.16f)", c.lo, c.hi) }

type c0 struct{ constRange }

type c1 struct{ constRange }

type search struct {
	r  *rand.Rand
	c0 c0
	c1 c1
}

func (s *search) FastLog2(x float32) float32 {
	tmp := math.Float32bits(x)
	expb := uint64(tmp) >> 23
	tmp = (tmp & 0x7fffff) | (0x7f << 23)
	out := math.Float32frombits(tmp) - 1
	return out*(float32(s.c0.Sample(s.r))+float32(s.c1.Sample(s.r))*out) - 127 + float32(expb)
}

func (s *search) Root() {
	s.c0, s.c1 = c0{constRange{c0Lo, c0Hi}}, c1{constRange{c1Lo, c1Hi}}
}
func (s *search) Expand(int) []mcts.FrontierAction {
	var actions []mcts.FrontierAction
	c0lo, c0hi, ok := s.c0.Subranges()
	if ok {
		actions = append(actions, []mcts.FrontierAction{{
			Action: c0{c0lo},
		}, {
			Action: c0{c0hi},
		}}...)
	}
	c1lo, c1hi, ok := s.c1.Subranges()
	if ok {
		actions = append(actions, []mcts.FrontierAction{{
			Action: c1{c1lo},
		}, {
			Action: c1{c1hi},
		}}...)
	}
	return actions
}
func (s *search) Select(a mcts.Action) {
	switch a := a.(type) {
	case c0:
		s.c0 = a
	case c1:
		s.c1 = a
	default:
		panic("bad action")
	}
}

var logSuite = []float32{
	1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
	17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
	33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48,
	49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64,
	65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 80,
	81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 96,
	97, 98, 99, 100,
	1e0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8,
	1e9, 1e10, 1e11, 1e12, 1e13, 1e14, 1e15, 1e16,
	1e17, 1e18, 1e19, 1e20, 1e21, 1e22, 1e23, 1e24,
	1e25, 1e26, 1e27, 1e28, 1e29, 1e30, 1e31, 1e32,
	1592711687, 2574516021, 1305384868, 1478721626,
	684071397, 38303249, 835903135, 1194741872,
	282137336, 2726605093, 3719867068, 3447891336,
	1009042351, 1439832450, 801825004, 1390806701,
	3706879091047852172, 695917195057204076, 8647907819143847933, 6336591080315746654,
	6572862327466004647, 2628158281871773928, 3917601317851761974, 7208026762471721904,
	9065284722760755584, 1871230148511534083, 10145437163404758739, 6491611064124502363,
	7971335130788603544, 6614297601711074061, 8513686800572735716, 16121352637960683916,
}

func (s *search) Rollout() (mcts.Score, int) {
	var mse float64
	var trials int
	for _, x := range logSuite {
		actual := math.Log2(float64(x))
		exp := float64(s.FastLog2(x))
		mse -= (actual - exp) * (actual - exp)
		trials++
	}
	return mcts.Score{[]float64{mse}, model.MinimizeObjective}, trials
}

func (s *search) Score() mcts.Score { return mcts.Score{[]float64{0}, model.MinimizeObjective} }

func main() {
	s := mcts.Search{
		SearchInterface: &search{r: rand.New(rand.NewSource(time.Now().UnixNano()))},
		NumEpisodes:     10000,
		ExploreFactor:   mcts.DefaultExploreFactor,
	}

	s.ExploreTemperature = 1000
	for printNSec := time.Now(); ; {
		s.Search()
		if time.Since(printNSec) > time.Second {
			sCopy := new(search)
			sCopy.Root()
			pv := s.PV()
			for _, e := range pv.TrimRoot() {
				sCopy.Select(e.Action())
			}
			fmt.Printf("[%.16f] %.16f, %.16f (%f)\n", pv.Last().Score(), sCopy.c0.Mid(), sCopy.c1.Mid(), pv.Last().NumRollouts())
			printNSec = time.Now()
		}
		if s.ExploreTemperature >= 1 {
			s.ExploreTemperature--
		}
	}
}
