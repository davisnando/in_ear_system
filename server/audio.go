package main

import (
	"math"
)

func audioMix(a float32, b float32) float32 {
	return a + b
}

func mixLogarithmicRangeCompression(i float32) float32 {
	if i < -1 {
		return float32(-math.Log(-float64(i)-0.85)/14 - 0.75)
	} else if i > 1 {
		return float32(math.Log(float64(i)-0.85)/14 + 0.75)
	} else {
		if i*2 > 1 || i*2 < -2 {
			return i
		}
		return i * 2
	}
}
