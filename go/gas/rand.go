package gas

import (
	"math/rand"
	"time"
)

// RandDuration returns duration between 0 and 1
func RandDuration(base time.Duration) time.Duration {
	return time.Duration(float64(base) * rand.Float64())
}
