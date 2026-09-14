package renderer

import "github.com/Kan-O435/okarina/internal/vecmath"

// このファイルは「地面 + Linkプレースホルダー」を画面に配置するだけの
// 最小構成のデモシーンを組み立てる。GLBモデルやテクスチャはまだ使わず、
// 単色シェーダーで配置・カメラまわりの土台を確認するためのもの。

const (
	groundHalfExtent = 5.0 // 地面quadの中心からの半径(X/Z方向)
	linkWidth         = 1.0
	linkHeight        = 2.0
)

// BuildFieldDemoScene は地面とLinkプレースホルダーを配置したSceneを組み立てる。
func BuildFieldDemoScene(c *Context) (*Scene, error) {
	program, err := c.NewProgram(basicVertexShaderSrc, basicFragmentShaderSrc)
	if err != nil {
		return nil, err
	}

	groundMesh := c.NewMesh(groundVertices(), quadIndices())
	linkMesh := c.NewMesh(linkVertices(), quadIndices())

	width, height := c.CanvasSize()
	aspect := float64(width) / float64(height)

	projection := vecmath.Perspective(vecmath.Radians(60), aspect, 0.1, 100)
	view := vecmath.LookAt(
		vecmath.NewVec3(0, 4, 8), // カメラ位置
		vecmath.NewVec3(0, 1, 0), // 注視点(Linkの胴体あたり)
		vecmath.NewVec3(0, 1, 0), // 上方向
	)

	return &Scene{
		Program:        program,
		ViewProjection: projection.Mul(view),
		Objects: []Object{
			{
				Mesh:      groundMesh,
				Transform: vecmath.Identity(),
				Color:     vecmath.NewVec3(0.35, 0.55, 0.25), // 草地っぽい緑
			},
			{
				Mesh:      linkMesh,
				Transform: vecmath.Identity(),
				Color:     vecmath.NewVec3(0.2, 0.5, 0.85), // Linkの仮配置(青系プレースホルダー)
			},
		},
	}, nil
}

// quadIndices は4頂点(左下, 右下, 右上, 左上)で構成される矩形を
// 三角形2枚として描画するための共通インデックス。
func quadIndices() []uint16 {
	return []uint16{0, 1, 2, 0, 2, 3}
}

// groundVertices はXZ平面上、y=0の正方形の地面を返す。
func groundVertices() []float32 {
	e := float32(groundHalfExtent)
	return []float32{
		-e, 0, -e,
		e, 0, -e,
		e, 0, e,
		-e, 0, e,
	}
}

// linkVertices はY-Z平面ではなくXY平面上に立てた、原点(足元)から
// 高さlinkHeightまでの板(ビルボード)を返す。GLBモデル導入までの仮配置用。
func linkVertices() []float32 {
	hw := float32(linkWidth) / 2
	h := float32(linkHeight)
	return []float32{
		-hw, 0, 0,
		hw, 0, 0,
		hw, h, 0,
		-hw, h, 0,
	}
}
