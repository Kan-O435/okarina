package renderer

import (
	"github.com/Kan-O435/okarina/internal/gltf"
	"github.com/Kan-O435/okarina/internal/vecmath"
)

// LoadGLBMesh はGLBバイナリ(例: Meshy AI/Tripo3Dの出力)をパースしてGPU用の
// Modelを作る。マテリアルのbaseColorFactor・baseColorTexture(あれば)・
// バウンディングボックスもあわせて反映する。
func (c *Context) LoadGLBMesh(data []byte) (*Model, error) {
	prim, err := gltf.Parse(data)
	if err != nil {
		return nil, err
	}
	return c.buildModel(prim)
}

// LoadSkinnedGLBMesh はLoadGLBMeshと同様だが、KayKit等の「体パーツごとに
// 複数メッシュへ分割された」人型キャラクターGLB向けに、スキン付きノードが
// 参照するメッシュをすべて結合してひとつのModelにする。
// スキニング(ボーンによる頂点変形)は未実装のため、アニメーションは反映されず
// エクスポート時の姿勢のまま静止表示になる。
func (c *Context) LoadSkinnedGLBMesh(data []byte) (*Model, error) {
	prim, err := gltf.ParseSkinned(data)
	if err != nil {
		return nil, err
	}
	return c.buildModel(prim)
}

// buildModel はgltf.Primitiveから実際にGPUリソース(Mesh・Texture)を作り、
// Modelにまとめる。baseColorTexture画像が無い場合は共有の白テクスチャを使う
// (Model.Colorだけで色が決まる)。
func (c *Context) buildModel(prim *gltf.Primitive) (*Model, error) {
	mesh := c.NewMesh(prim.Positions, prim.TexCoords, prim.Indices)

	texture := c.WhiteTexture()
	if prim.TextureData != nil {
		decoded, err := c.NewImageTexture(prim.TextureData, prim.TextureMimeType)
		if err != nil {
			return nil, err
		}
		texture = decoded
	}

	color := vecmath.NewVec3(
		float64(prim.BaseColor[0]),
		float64(prim.BaseColor[1]),
		float64(prim.BaseColor[2]),
	)

	return &Model{
		Mesh:    mesh,
		Texture: texture,
		Color:   color,
		Min:     vecmath.NewVec3(float64(prim.Min[0]), float64(prim.Min[1]), float64(prim.Min[2])),
		Max:     vecmath.NewVec3(float64(prim.Max[0]), float64(prim.Max[1]), float64(prim.Max[2])),
	}, nil
}
