package gas

import "math"

// Ease adjusts the timing of anims
type Ease func(x float32) float32

var EaseNone = Ease(func(x float32) float32 { return x })
var EaseOutSin = Ease(func(x float32) float32 { return float32(math.Sin(float64(x) * 1.57)) })
var EaseInOutSin = Ease(func(x float32) float32 { return float32(-(math.Cos(3.14*float64(x)) - 1) / 2) })
var EaseInOutSinInv = Ease(func(x float32) float32 { return float32(1 - (-(math.Cos(3.14*float64(x)) - 1) / 2)) })
