package renderer

import "github.com/Kan-O435/okarina/internal/vecmath"

// このファイルは、扉を抜けた先の「草原フィールド」のデモシーンを組み立てる。
// 神殿フィールド(demo.go)と同じく、まずは地面だけの最小構成から始め、
// 必要になった要素を後から追加していく(CLAUDE.md 27章の方針)。

// grasslandGroundHalfExtent は地面の半径(1辺の半分)。神殿フィールドと
// 同様、カメラのfarより十分大きくして地平線まで緑が続くようにしておく。
const grasslandGroundHalfExtent = 500.0

// BuildGrasslandScene は、草原フィールドの土台(地面を緑で埋め尽くしただけ)
// を配置したSceneを組み立てる。
func BuildGrasslandScene(c *Context) (*Scene, error) {
	program, err := c.NewProgram(basicVertexShaderSrc, basicFragmentShaderSrc)
	if err != nil {
		return nil, err
	}

	width, height := c.CanvasSize()
	aspect := float64(width) / float64(height)
	projection := vecmath.Perspective(vecmath.Radians(55), aspect, 0.1, 150)
	view := vecmath.LookAt(
		vecmath.NewVec3(0, 8, 12), // カメラ位置: 地面を見渡せる高さ・距離
		vecmath.NewVec3(0, 0, -10),
		vecmath.NewVec3(0, 1, 0),
	)

	return &Scene{
		Program:        program,
		ViewProjection: projection.Mul(view),
		Objects:        []Object{grasslandGroundObject(c)},
	}, nil
}

func grasslandGroundObject(c *Context) Object {
	verts := quadVertices(grasslandGroundHalfExtent, grasslandGroundHalfExtent, 0)
	mesh := c.NewMesh(verts, zeroUVs(4), quadIndices())
	return Object{Mesh: mesh, Transform: vecmath.Identity(), Color: vecmath.NewVec3(0.3, 0.65, 0.25)}
}
