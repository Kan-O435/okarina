package vecmath

import "math"

// Mat4 は4x4行列。WebGL(uniformMatrix4fv)に合わせて列優先(column-major)で
// 16要素の配列に格納する。要素 m[col*4+row] がcol列row行の値にあたる。
type Mat4 [16]float64

// Identity は単位行列を返す。
func Identity() Mat4 {
	return Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

// Translate は平行移動行列を返す。
func Translate(v Vec3) Mat4 {
	m := Identity()
	m[12], m[13], m[14] = v.X, v.Y, v.Z
	return m
}

// Scale は拡大縮小行列を返す。
func Scale(v Vec3) Mat4 {
	return Mat4{
		v.X, 0, 0, 0,
		0, v.Y, 0, 0,
		0, 0, v.Z, 0,
		0, 0, 0, 1,
	}
}

// RotateX はX軸周りにrad(ラジアン)回転する行列を返す。
func RotateX(rad float64) Mat4 {
	c, s := math.Cos(rad), math.Sin(rad)
	return Mat4{
		1, 0, 0, 0,
		0, c, s, 0,
		0, -s, c, 0,
		0, 0, 0, 1,
	}
}

// RotateY はY軸周りにrad(ラジアン)回転する行列を返す。
func RotateY(rad float64) Mat4 {
	c, s := math.Cos(rad), math.Sin(rad)
	return Mat4{
		c, 0, -s, 0,
		0, 1, 0, 0,
		s, 0, c, 0,
		0, 0, 0, 1,
	}
}

// RotateZ はZ軸周りにrad(ラジアン)回転する行列を返す。
func RotateZ(rad float64) Mat4 {
	c, s := math.Cos(rad), math.Sin(rad)
	return Mat4{
		c, s, 0, 0,
		-s, c, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

// Mul は行列積 m * other を返す(mを先に適用したい場合は other.Mul(m) ではなく
// m.Mul(other) が「other の後に m を適用する」変換になる点に注意。
// 例: MVP = Projection.Mul(View).Mul(Model))
func (m Mat4) Mul(other Mat4) Mat4 {
	var result Mat4
	for col := 0; col < 4; col++ {
		for row := 0; row < 4; row++ {
			var sum float64
			for k := 0; k < 4; k++ {
				sum += m[k*4+row] * other[col*4+k]
			}
			result[col*4+row] = sum
		}
	}
	return result
}

// MulVec4 は同次座標ベクトル(x, y, z, w)に行列を適用する。
func (m Mat4) MulVec4(v [4]float64) [4]float64 {
	var result [4]float64
	for row := 0; row < 4; row++ {
		var sum float64
		for k := 0; k < 4; k++ {
			sum += m[k*4+row] * v[k]
		}
		result[row] = sum
	}
	return result
}

// Perspective は透視投影行列を返す。
// fovYRad: 縦方向の視野角(ラジアン), aspect: 画面の縦横比(width/height),
// near, far: 描画するZ範囲(共に正の値、near < far)。
func Perspective(fovYRad, aspect, near, far float64) Mat4 {
	f := 1 / math.Tan(fovYRad/2)
	m := Mat4{}
	m[0] = f / aspect
	m[5] = f
	m[10] = (far + near) / (near - far)
	m[11] = -1
	m[14] = (2 * far * near) / (near - far)
	return m
}

// Ortho は平行投影(正射影)行列を返す。遠近感を付けたくない描画
// (HUDなど、スクリーン座標にそのまま貼り付ける2D要素)に使う。
// left/right/bottom/topは描画範囲、near/farはZ範囲。
func Ortho(left, right, bottom, top, near, far float64) Mat4 {
	m := Identity()
	m[0] = 2 / (right - left)
	m[5] = 2 / (top - bottom)
	m[10] = -2 / (far - near)
	m[12] = -(right + left) / (right - left)
	m[13] = -(top + bottom) / (top - bottom)
	m[14] = -(far + near) / (far - near)
	return m
}

// LookAt はeye(カメラ位置)からtarget(注視点)を見るビュー行列を返す。upは上方向。
func LookAt(eye, target, up Vec3) Mat4 {
	zAxis := eye.Sub(target).Normalize()
	xAxis := up.Cross(zAxis).Normalize()
	yAxis := zAxis.Cross(xAxis)

	return Mat4{
		xAxis.X, yAxis.X, zAxis.X, 0,
		xAxis.Y, yAxis.Y, zAxis.Y, 0,
		xAxis.Z, yAxis.Z, zAxis.Z, 0,
		-xAxis.Dot(eye), -yAxis.Dot(eye), -zAxis.Dot(eye), 1,
	}
}
