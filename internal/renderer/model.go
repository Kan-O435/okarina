package renderer

import "github.com/Kan-O435/okarina/internal/vecmath"

// Model はGLB等から読み込んだ1メッシュ分のデータ(GPU用Mesh・テクスチャ・色・
// ローカル座標系でのバウンディングボックス)をまとめたもの。
type Model struct {
	Mesh    *Mesh
	Texture *Texture
	Color   vecmath.Vec3
	Min     vecmath.Vec3
	Max     vecmath.Vec3
}

// GroundTransform は、このModelを「底面がY=0(地面)に接し、X/Z中心が
// ワールド座標(x, z)に来る」ように配置する変換行列を返す。
// targetHeightに正の値を渡すと、モデルの高さがtargetHeightになるよう
// 均一スケールする(0以下ならスケールしない)。
//
// GLBモデルは原点や向きがエクスポート元によってまちまちなため、
// バウンディングボックスから逆算して「置きたい場所にちょうど立つ」
// 変換を機械的に求める。
func (m *Model) GroundTransform(x, z, targetHeight float64) vecmath.Mat4 {
	height := m.Max.Y - m.Min.Y
	scale := 1.0
	if targetHeight > 0 && height > 0 {
		scale = targetHeight / height
	}

	centerX := (m.Min.X + m.Max.X) / 2
	centerZ := (m.Min.Z + m.Max.Z) / 2

	toOrigin := vecmath.Translate(vecmath.NewVec3(-centerX, -m.Min.Y, -centerZ))
	scaleMat := vecmath.Scale(vecmath.NewVec3(scale, scale, scale))
	toWorld := vecmath.Translate(vecmath.NewVec3(x, 0, z))

	return toWorld.Mul(scaleMat).Mul(toOrigin)
}
