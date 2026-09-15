package vecmath

import "math"

// Radians は度(degree)をラジアンに変換する。カメラのFOV指定などで使う。
func Radians(deg float64) float64 {
	return deg * math.Pi / 180
}
