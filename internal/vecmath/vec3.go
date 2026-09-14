// Package vecmath は、3Dエンジンの基礎となるベクトル・行列演算を提供する。
// WebGL/GLSLとの対応を意識し、Mat4は列優先(column-major)で扱う。
package vecmath

import "math"

// Vec3 は3次元ベクトル。座標・移動方向・カメラ位置・光源位置などに使う。
type Vec3 struct {
	X, Y, Z float64
}

// NewVec3 はVec3を生成する。
func NewVec3(x, y, z float64) Vec3 {
	return Vec3{X: x, Y: y, Z: z}
}

// Add はベクトルの加算を返す。
func (v Vec3) Add(other Vec3) Vec3 {
	return Vec3{v.X + other.X, v.Y + other.Y, v.Z + other.Z}
}

// Sub はベクトルの減算を返す。
func (v Vec3) Sub(other Vec3) Vec3 {
	return Vec3{v.X - other.X, v.Y - other.Y, v.Z - other.Z}
}

// Scale はスカラー倍を返す。
func (v Vec3) Scale(s float64) Vec3 {
	return Vec3{v.X * s, v.Y * s, v.Z * s}
}

// Dot は内積を返す。
func (v Vec3) Dot(other Vec3) float64 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

// Cross は外積を返す。
func (v Vec3) Cross(other Vec3) Vec3 {
	return Vec3{
		X: v.Y*other.Z - v.Z*other.Y,
		Y: v.Z*other.X - v.X*other.Z,
		Z: v.X*other.Y - v.Y*other.X,
	}
}

// Length はベクトルの長さを返す。
func (v Vec3) Length() float64 {
	return math.Sqrt(v.Dot(v))
}

// Normalize は長さ1に正規化したベクトルを返す。
// ゼロベクトルの場合はゼロベクトルをそのまま返す。
func (v Vec3) Normalize() Vec3 {
	length := v.Length()
	if length == 0 {
		return v
	}
	return v.Scale(1 / length)
}
