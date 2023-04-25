package gas

import (
	math32 "github.com/chewxy/math32"
)

// Ease adjusts the timing of anims
type Ease func(x float32) float32

var EaseNone = Ease(func(x float32) float32 { return x })
var EaseOutSin = Ease(func(x float32) float32 { return math32.Sin(x * 1.57) })
var EaseInOutSin = Ease(func(x float32) float32 { return -(math32.Cos(3.14*x) - 1) / 2 })
var EaseInOutSinInv = Ease(func(x float32) float32 { return 1 - (-(math32.Cos(3.14*x) - 1) / 2) })
