package mcts

import (
	"fmt"
	"unsafe"
)

type Counter interface{}

type Score[T Counter] struct {
	Counter   T
	Objective func(T) float64
}

func (s Score[T]) Apply() float64 { return s.Objective(s.Counter) }

func (s Score[T]) SafeApply() (float64, bool) {
	if s.Objective == nil {
		return 0, false
	}
	return s.Objective(s.Counter), true
}

type CounterInterface[T Counter] struct {
	// Add counters and store the result in x.
	//
	// For common values of T this will be populated automatically.
	// The common values of T are: float32, float64, [2]float64, []float64, int, and int64.
	// Custom counters or those in the model package need to be supplied manually.
	Add func(x *T, y T)
}

func patchBuiltinAdd[T Counter](c *CounterInterface[T]) {
	var t T
	switch any(t).(type) {
	case int:
		f := func(x1 *int, x2 int) { *x1 += x2 }
		c.Add = *(*func(*T, T))(unsafe.Pointer(&f))
	case int64:
		f := func(x1 *int64, x2 int64) { *x1 += x2 }
		c.Add = *(*func(*T, T))(unsafe.Pointer(&f))
	case float32:
		f := func(x1 *float32, x2 float32) { *x1 += x2 }
		c.Add = *(*func(*T, T))(unsafe.Pointer(&f))
	case float64:
		f := func(x1 *float64, x2 float64) { *x1 += x2 }
		c.Add = *(*func(*T, T))(unsafe.Pointer(&f))
	case [2]int:
		f := func(c1 *[2]int, c2 [2]int) { c1[0] += c2[0]; c1[1] += c2[1] }
		c.Add = *(*func(*T, T))(unsafe.Pointer(&f))
	case [2]int64:
		f := func(c1 *[2]int64, c2 [2]int64) { c1[0] += c2[0]; c1[1] += c2[1] }
		c.Add = *(*func(*T, T))(unsafe.Pointer(&f))
	case [2]float32:
		f := func(c1 *[2]float32, c2 [2]float32) { c1[0] += c2[0]; c1[1] += c2[1] }
		c.Add = *(*func(*T, T))(unsafe.Pointer(&f))
	case [2]float64:
		f := func(c1 *[2]float64, c2 [2]float64) { c1[0] += c2[0]; c1[1] += c2[1] }
		c.Add = *(*func(*T, T))(unsafe.Pointer(&f))
	case []int:
		f := func(c1 *[]int, c2 []int) {
			for i, v := range c2 {
				(*c1)[i] += v
			}
		}
		c.Add = *(*func(*T, T))(unsafe.Pointer(&f))
	case []int64:
		f := func(c1 *[]int64, c2 []int64) {
			for i, v := range c2 {
				(*c1)[i] += v
			}
		}
		c.Add = *(*func(*T, T))(unsafe.Pointer(&f))
	case []float32:
		f := func(c1 *[]float32, c2 []float32) {
			for i, v := range c2 {
				(*c1)[i] += v
			}
		}
		c.Add = *(*func(*T, T))(unsafe.Pointer(&f))
	case []float64:
		f := func(c1 *[]float64, c2 []float64) {
			for i, v := range c2 {
				(*c1)[i] += v
			}
		}
		c.Add = *(*func(*T, T))(unsafe.Pointer(&f))
	default:
		panic(fmt.Errorf("could not automatically set the implementation for Add because type %T is not a supported type", t))
	}
}
