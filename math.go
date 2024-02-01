package mcts

import "math"

// Adapted from https://github.com/LeelaChessZero/lc0/blob/d89a4565fb440d0694e7fea0b11b368a4ba117a9/src/utils/fastmath.h#L42.

func fastLog2(x float32) float32 {
	tmp := math.Float32bits(x)
	expb := uint64(tmp) >> 23
	tmp = (tmp & 0x7fffff) | (0x7f << 23)
	out := math.Float32frombits(tmp) - 1
	return out*(1.3465552-0.34655523*out) - 127 + float32(expb)
}

func fastLog(x float32) float32 { return 6.93147180369123816490e-01 * fastLog2(x) }
