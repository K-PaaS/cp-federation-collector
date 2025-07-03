package util

import "math"

func Round(num float64, decimals int) float64 {
	pow10 := math.Pow10(decimals)
	return math.Round(num*pow10) / pow10
}
