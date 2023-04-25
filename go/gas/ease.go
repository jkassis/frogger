package gas

import "math"

// Ease adjusts the timing of anims
type Ease func(x float64) float64

var EaseNone = Ease(func(x float64) float64 { return x })
var EaseOutSin = Ease(func(x float64) float64 { return math.Sin(x * 1.57) })
var EaseInOutSin = Ease(func(x float64) float64 { return -(math.Cos(3.14*x) - 1) / 2 })
