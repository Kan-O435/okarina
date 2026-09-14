package renderer

import (
	"sort"

	"github.com/Kan-O435/okarina/internal/assets"
	"github.com/Kan-O435/okarina/internal/vecmath"
)

// このファイルは、扉を抜けた先の「草原フィールド」のデモシーンを組み立てる。
// 神殿フィールド(demo.go)と同じく、まずは地面だけの最小構成から始め、
// 必要になった要素を後から追加していく(CLAUDE.md 27章の方針)。

// grasslandGroundHalfExtent は地面の半径(1辺の半分)。神殿フィールドと
// 同様、カメラのfarより十分大きくして地平線まで緑が続くようにしておく。
const grasslandGroundHalfExtent = 500.0

// grasslandPathHalfWidth は、地面の真ん中を通る道の半幅。神殿フィールドの
// 参道(石畳)より狭く、人が踏み固めただけの細い道をイメージしている。
const grasslandPathHalfWidth = 1.8

// grasslandLinkZ は、神殿フィールドの扉を抜けてきたLinkを道のいちばん手前
// (カメラに近い側、画面下寄り)に立たせておく位置。
const grasslandLinkZ = 3.5

// grasslandCastleZ/Height は、道の先端(遠景、far=150に収まる範囲でできる
// だけ奥)に置く白い城(internal/assets.GrasslandCastle)の位置・高さ。
// 高さ11だと上端の仰角はカメラ中心からおよそ21.5°(垂直画角の半分27.5°の
// うち、画面上端まで約6°分の余白が残る)。この余白に手段A(放射状
// グラデーションのハロー)を後から重ねられる。
const (
	grasslandCastleZ      = -58.0
	grasslandCastleHeight = 21.0
)

// grasslandHalo* は、城の背後に浮かべる白い後光(放射状グラデーションの板)
// の位置・大きさ。城よりさらに奥(Z)・城の頂上あたりの高さ(Y)に置くことで、
// 城のシルエットの背後から光が差しているように見せる。
const (
	grasslandHaloZ          = grasslandCastleZ - 15 // 城よりさらに奥
	grasslandHaloCenterY    = grasslandCastleHeight // 城の頂上あたりの高さ
	grasslandHaloHalfWidth  = 15.0
	grasslandHaloHalfHeight = 12.0
)

// BuildGrasslandScene は、草原フィールドの土台(地面・道・Link)を配置した
// Sceneを組み立てる。戻り値のLinkPlacementは、demo.goの神殿フィールドと
// 同様、プレイヤー移動に合わせて呼び出し側がLinkのTransformを書き換える
// ために使う。
func BuildGrasslandScene(c *Context) (scene *Scene, link LinkPlacement, err error) {
	program, err := c.NewProgram(basicVertexShaderSrc, basicFragmentShaderSrc)
	if err != nil {
		return nil, LinkPlacement{}, err
	}

	width, height := c.CanvasSize()
	aspect := float64(width) / float64(height)
	// FOV: 55°では城をこれ以上手前・大きくする余白が無くなったため65°に
	// 広げ、画面上端までの余白を確保している(木の配置はまだ十分内側にある
	// ため、広げても画面外に出ることはない)。
	projection := vecmath.Perspective(vecmath.Radians(65), aspect, 0.1, 150)
	view := vecmath.LookAt(
		vecmath.NewVec3(0, 8, 12), // カメラ位置: 地面を見渡せる高さ・距離
		vecmath.NewVec3(0, 0, -10),
		vecmath.NewVec3(0, 1, 0),
	)

	// まず不透明なオブジェクトをすべて配置する(深度テストがあるため、
	// 互いの前後関係は描画順によらず正しく処理される)。
	objects := []Object{grasslandGroundObject(c), grasslandPathObject(c)}

	castle, err := grasslandCastleObject(c)
	if err != nil {
		return nil, LinkPlacement{}, err
	}
	objects = append(objects, castle)

	trees, err := grasslandTreeObjects(c)
	if err != nil {
		return nil, LinkPlacement{}, err
	}
	objects = append(objects, trees...)

	linkObj, linkLocal, err := grasslandLinkObject(c)
	if err != nil {
		return nil, LinkPlacement{}, err
	}
	objects = append(objects, linkObj)
	link = LinkPlacement{
		Index:          len(objects) - 1,
		LocalTransform: linkLocal,
		SpawnZ:         grasslandLinkZ,
	}

	// 透過オブジェクト(後光・雲)は、不透明なオブジェクトがすべて描かれた
	// 後に描画する。先に描くと、まだ何も描かれていない背景色と合成されて
	// しまい、後から描く不透明なオブジェクトにその形の穴が空いたように
	// 見えてしまう。また、透過オブジェクト同士も奥から手前の順(back-to-
	// front)で描く必要がある。このシェーダーはアルファに応じたフラグメント
	// の破棄(alpha discard)をしないため、深度バッファには透明な部分にも
	// 深度が書き込まれる。手前のものを先に描くと、その「見えない」はずの
	// 透明な部分が奥のものを誤って隠してしまう。
	skyObjects, err := grasslandSkyObjects(c)
	if err != nil {
		return nil, LinkPlacement{}, err
	}
	objects = append(objects, skyObjects...)

	return &Scene{
		Program:        program,
		ViewProjection: projection.Mul(view),
		Objects:        objects,
	}, link, nil
}

func grasslandGroundObject(c *Context) Object {
	verts := quadVertices(grasslandGroundHalfExtent, grasslandGroundHalfExtent, 0)
	mesh := c.NewMesh(verts, zeroUVs(4), quadIndices())
	return Object{Mesh: mesh, Transform: vecmath.Identity(), Color: vecmath.NewVec3(0.3, 0.65, 0.25)}
}

// grasslandPathObject は、地面の真ん中(X=0)を奥まで貫く、人が通った跡の
// ような黄色っぽい土の道。地面(Y=0)とのZ-fighting防止のためわずかに
// (Y=0.01)持ち上げている。
func grasslandPathObject(c *Context) Object {
	verts := quadVertices(grasslandPathHalfWidth, grasslandGroundHalfExtent, 0.01)
	mesh := c.NewMesh(verts, zeroUVs(4), quadIndices())
	return Object{Mesh: mesh, Transform: vecmath.Identity(), Color: vecmath.NewVec3(0.75, 0.68, 0.35)}
}

// grasslandCastleColor は、テクスチャの無い白い城に付ける色。参考画像の
// ような、夕暮れ・後光に照らされた紫がかった石造りの見た目を狙っている。
var grasslandCastleColor = vecmath.NewVec3(0.24, 0.18, 0.32)

// grasslandCastleObject は、道の先(grasslandCastleZ)に白い城
// (internal/assets.GrasslandCastle、Tripo3Dで生成・リメッシュ済み、
// テクスチャ無しの単色モデル)を配置する。
func grasslandCastleObject(c *Context) (Object, error) {
	model, err := c.LoadGLBMesh(assets.GrasslandCastle)
	if err != nil {
		return Object{}, err
	}

	transform := model.GroundTransform(0, grasslandCastleZ, grasslandCastleHeight)
	return Object{Mesh: model.Mesh, Texture: model.Texture, Transform: transform, Color: grasslandCastleColor}, nil
}

// grasslandCloudColor は、雲の板に付ける色。「濃いめの紫」の不穏な雲を狙う。
var grasslandCloudColor = vecmath.NewVec3(0.32, 0.24, 0.42)

// skyBillboard は、空に浮かべる透過オブジェクト(後光・雲)1枚分の配置情報。
type skyBillboard struct {
	X, Y, Z      float64
	HalfW, HalfH float64
	Texture      *Texture
	Color        vecmath.Vec3
}

// grasslandSkyObjects は、後光・雲(いずれも透過テクスチャの板)をまとめて
// 組み立てる。このシェーダーはalpha discardをしないため、透明な部分にも
// 深度が書き込まれる。手前のものを先に描くと、その「見えない」はずの透明な
// 部分が奥のものを誤って隠してしまうため、奥から手前の順(Zの小さい順)に
// 並べ替えてから描画用Objectを作る。
func grasslandSkyObjects(c *Context) ([]Object, error) {
	haloTexture, err := c.NewImageTexture(assets.HaloTexture, "image/png")
	if err != nil {
		return nil, err
	}
	cloudTexture, err := c.NewImageTexture(assets.CloudTexture, "image/png")
	if err != nil {
		return nil, err
	}

	// 城や後光と重なりすぎない位置に、奥行きを変えて雲をばらけさせている。
	placements := []skyBillboard{
		{X: 0, Y: grasslandHaloCenterY, Z: grasslandHaloZ, HalfW: grasslandHaloHalfWidth, HalfH: grasslandHaloHalfHeight, Texture: haloTexture, Color: vecmath.NewVec3(1, 1, 1)},
		{X: -26, Y: 24, Z: -45, HalfW: 20, HalfH: 11, Texture: cloudTexture, Color: grasslandCloudColor},
		{X: 22, Y: 20, Z: -65, HalfW: 16, HalfH: 9, Texture: cloudTexture, Color: grasslandCloudColor},
		{X: 4, Y: 27, Z: -90, HalfW: 22, HalfH: 12, Texture: cloudTexture, Color: grasslandCloudColor},
		{X: -45, Y: 22, Z: -60, HalfW: 18, HalfH: 10, Texture: cloudTexture, Color: grasslandCloudColor},
	}
	sort.Slice(placements, func(i, j int) bool {
		return placements[i].Z < placements[j].Z
	})

	objects := make([]Object, 0, len(placements))
	for _, p := range placements {
		verts := centeredQuadVertices(float32(p.HalfW), float32(p.HalfH))
		mesh := c.NewMesh(verts, quadUVsCropped(0, 0, 1, 1), quadIndices())
		transform := vecmath.Translate(vecmath.NewVec3(p.X, p.Y, p.Z))
		objects = append(objects, Object{
			Mesh:      mesh,
			Texture:   p.Texture,
			Transform: transform,
			Color:     p.Color,
		})
	}
	return objects, nil
}

// grasslandTreeHeight は、草原フィールドに置く木の高さ。
const grasslandTreeHeight = 5.5

// grasslandTreeObjects は、神殿フィールドと同じ木のモデル(CC0、
// internal/assets.Trees)を1本読み込み、道を挟んで右奥(画面右上寄り)と
// 左手前(画面左下寄り)に1本ずつ配置する。
func grasslandTreeObjects(c *Context) ([]Object, error) {
	entries, err := assets.Trees.ReadDir("models/trees")
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, nil
	}

	data, err := assets.Trees.ReadFile("models/trees/" + entries[0].Name())
	if err != nil {
		return nil, err
	}
	model, err := c.LoadGLBMesh(data)
	if err != nil {
		return nil, err
	}

	placements := []treePlacement{
		{X: 16, Z: -25, Height: grasslandTreeHeight}, // 道の右奥(画面右上寄り)
		{X: -5, Z: 2.5, Height: grasslandTreeHeight}, // 道の左手前(画面左下寄り、視錐台に収まる位置)
	}

	objects := make([]Object, 0, len(placements))
	for _, p := range placements {
		transform := model.GroundTransform(p.X, p.Z, p.Height)
		objects = append(objects, Object{Mesh: model.Mesh, Texture: model.Texture, Transform: transform, Color: model.Color})
	}
	return objects, nil
}

// grasslandLinkObject は、神殿フィールドと同じKayKit Knightモデル
// (internal/assets.LinkKnight)を、道のいちばん手前(grasslandLinkZ)に配置
// する。神殿の扉を抜けてきた直後、という体で置いている。demo.goのlinkObject
// と同様、ワールド位置を含まないlocalTransform(モデル原点補正+スケール)を
// 別に返し、呼び出し側が毎フレームプレイヤーの位置・向きと組み合わせて
// Transformを再構築できるようにする。
func grasslandLinkObject(c *Context) (obj Object, localTransform vecmath.Mat4, err error) {
	model, err := c.LoadSkinnedGLBMesh(assets.LinkKnight)
	if err != nil {
		return Object{}, vecmath.Mat4{}, err
	}

	localTransform = model.GroundTransform(0, 0, linkTargetHeight)
	transform := vecmath.Translate(vecmath.NewVec3(0, 0, grasslandLinkZ)).Mul(localTransform)
	obj = Object{Mesh: model.Mesh, Texture: model.Texture, Transform: transform, Color: model.Color}
	return obj, localTransform, nil
}
