package renderer

import (
	"github.com/Kan-O435/okarina/internal/assets"
	"github.com/Kan-O435/okarina/internal/vecmath"
)

// このファイルは「時の神殿」フィールド(docs/fields/temple-of-time.md参照)の
// 土台となるデモシーンを組み立てる。
//
// 自作で完結する部分(地面・道・プール・階段)は実寸で配置する。
// 建物本体+双尖塔はTripo3Dで生成・リメッシュ済みのGLB(internal/assets)を
// go:embedで読み込んで実際に配置している。扉(#3)はまだ生成中のため、
// マゼンタ色のプレースホルダーボックスで場所だけ確保している。

// placeholderColor は「まだ実アセットに差し替わっていない」ことを示す
// 目印の色(ゲーム開発で言う "missing texture" のマゼンタ表記に倣う)。
var placeholderColor = vecmath.NewVec3(1.0, 0.0, 1.0)

const (
	groundHalfExtent = 500.0 // 地面: カメラのfar(150)より十分大きく、実質無限に見えるサイズにしておく

	linkSpawnZ = 2.0 // Linkの初期位置。カメラ(0,12,18)からの俯角(約15°+FOV55°)だと
	// z=28はカメラの視錐台の外(画面外)になってしまうため、
	// カメラと神殿の間の、画面に収まる位置まで手前に寄せている
	linkTargetHeight = 1.8 // KayKit Knightモデルをこの高さになるようスケールする

	pathHalfWidth = 2.0
	pathNearZ     = 22.0  // 道の手前端(スポーン地点)
	pathFarZ      = -16.0 // 道の奥端(階段の手前)

	poolHalfWidth = 2.0
	poolHalfDepth = 4.0
	poolCenterZ   = -10.0
	poolOffsetX   = 5.0 // 道を挟んで左右に配置

	stepHalfWidth = 3.0
	stepHalfDepth = 0.5

	buildingTargetHeight = 10.0 // GLBモデルをこの高さになるようスケールする
	buildingCenterZ      = -24.0

	doorHalfWidth = 1.0
	doorHalfDepth = 0.15
	doorHeight    = 3.0
	doorCenterZ   = -20.5 // 建物の正面より少し手前に張り出させ、埋もれないようにする
)

// BuildFieldDemoScene は「時の神殿」フィールドの土台(地面・道・プール・階段+
// 建物/塔/扉のプレースホルダー)を配置したSceneを組み立てる。
// 戻り値のlinkIndexは、scene.ObjectsのうちLinkに対応する要素のインデックス
// (プレイヤー移動に合わせて呼び出し側がTransformを書き換えるために使う)。
func BuildFieldDemoScene(c *Context) (scene *Scene, linkIndex int, err error) {
	program, err := c.NewProgram(basicVertexShaderSrc, basicFragmentShaderSrc)
	if err != nil {
		return nil, -1, err
	}

	width, height := c.CanvasSize()
	aspect := float64(width) / float64(height)
	projection := vecmath.Perspective(vecmath.Radians(55), aspect, 0.1, 150)
	view := vecmath.LookAt(
		vecmath.NewVec3(0, 12, 18), // カメラ位置: 神殿(塔)に寄せた高さ・距離
		vecmath.NewVec3(0, 3, -15), // 注視点: 神殿の手前あたり
		vecmath.NewVec3(0, 1, 0),
	)

	objects := []Object{
		groundObject(c),
		pathObject(c),
		poolObject(c, -poolOffsetX),
		poolObject(c, poolOffsetX),
		doorPlaceholderObject(c),
	}
	objects = append(objects, stepObjects(c)...)

	templeBody, err := templeBodyObject(c)
	if err != nil {
		return nil, -1, err
	}
	objects = append(objects, templeBody)

	link, err := linkObject(c)
	if err != nil {
		return nil, -1, err
	}
	objects = append(objects, link)
	linkIndex = len(objects) - 1

	return &Scene{
		Program:        program,
		ViewProjection: projection.Mul(view),
		Objects:        objects,
	}, linkIndex, nil
}

func groundObject(c *Context) Object {
	verts := quadVertices(groundHalfExtent, groundHalfExtent, 0)
	mesh := c.NewMesh(verts, zeroUVs(4), quadIndices())
	return Object{Mesh: mesh, Transform: vecmath.Identity(), Color: vecmath.NewVec3(0.35, 0.55, 0.25)}
}

func pathObject(c *Context) Object {
	halfDepth := float32((pathNearZ - pathFarZ) / 2)
	centerZ := float32((pathNearZ + pathFarZ) / 2)
	verts := quadVertices(pathHalfWidth, halfDepth, 0.01)
	mesh := c.NewMesh(verts, zeroUVs(4), quadIndices())
	transform := vecmath.Translate(vecmath.NewVec3(0, 0, float64(centerZ)))
	return Object{Mesh: mesh, Transform: transform, Color: vecmath.NewVec3(0.6, 0.55, 0.45)}
}

func poolObject(c *Context, offsetX float32) Object {
	verts := quadVertices(poolHalfWidth, poolHalfDepth, 0.01)
	mesh := c.NewMesh(verts, zeroUVs(4), quadIndices())
	transform := vecmath.Translate(vecmath.NewVec3(float64(offsetX), 0, poolCenterZ))
	return Object{Mesh: mesh, Transform: transform, Color: vecmath.NewVec3(0.25, 0.45, 0.65)}
}

// stepObjects は神殿の手前に、奥に向かって徐々に高くなる階段を3段配置する。
func stepObjects(c *Context) []Object {
	heights := []float32{0.4, 0.8, 1.2}
	zPositions := []float32{-16, -17, -18}

	objects := make([]Object, 0, len(heights))
	for i, h := range heights {
		verts := boxVertices(stepHalfWidth, h, stepHalfDepth)
		mesh := c.NewMesh(verts, zeroUVs(8), boxIndices())
		transform := vecmath.Translate(vecmath.NewVec3(0, 0, float64(zPositions[i])))
		objects = append(objects, Object{Mesh: mesh, Transform: transform, Color: vecmath.NewVec3(0.55, 0.53, 0.5)})
	}
	return objects
}

// templeBodyObject は、Tripo3Dで生成・リメッシュ済みの建物本体+双尖塔GLB
// (internal/assets.TempleBody)を読み込み、地面のbuildingCenterZの位置に
// 高さbuildingTargetHeightで立つよう配置する。
func templeBodyObject(c *Context) (Object, error) {
	model, err := c.LoadGLBMesh(assets.TempleBody)
	if err != nil {
		return Object{}, err
	}

	transform := model.GroundTransform(0, buildingCenterZ, buildingTargetHeight)
	return Object{Mesh: model.Mesh, Texture: model.Texture, Transform: transform, Color: model.Color}, nil
}

// doorPlaceholderObject は、まだMeshyで生成中の「扉」(#3)の置き場所を
// 確保するプレースホルダー。生成物が届いたらMeshをLoadGLBMesh()の結果に差し替える。
func doorPlaceholderObject(c *Context) Object {
	verts := boxVertices(doorHalfWidth, doorHeight, doorHalfDepth)
	return Object{
		Mesh:      c.NewMesh(verts, zeroUVs(8), boxIndices()),
		Transform: vecmath.Translate(vecmath.NewVec3(0, 0, doorCenterZ)),
		Color:     placeholderColor,
	}
}

// linkObject は、KayKit Adventurers(CC0)のKnightモデル
// (internal/assets.LinkKnight)を読み込み、フィールド手前のスポーン地点に
// 立つよう配置する。体パーツごとに分かれた複数メッシュを結合して1体として
// 表示するため LoadSkinnedGLBMesh を使う。アニメーション(Idle/Walking等)は
// GLB内に含まれているが、スキニングは未実装のため現状は静止表示のみ。
func linkObject(c *Context) (Object, error) {
	model, err := c.LoadSkinnedGLBMesh(assets.LinkKnight)
	if err != nil {
		return Object{}, err
	}

	transform := model.GroundTransform(0, linkSpawnZ, linkTargetHeight)
	return Object{Mesh: model.Mesh, Texture: model.Texture, Transform: transform, Color: model.Color}, nil
}
