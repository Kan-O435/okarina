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

// GrasslandTreeTriggerZ は、Linkが右奥の木(grasslandTreeObjectsの
// X:16の木)のあたりまで進んだら、次のフィールド(ガノン)へページ遷移
// するトリガーとして使うZ座標。最後の柵(grasslandObstacleZs末尾)を
// 越えた後、少し走ってから次のフィールドへ着くよう余裕を持たせている。
const GrasslandTreeTriggerZ = -67.0

// grasslandCastleZ/Height は、道の先端(遠景、far=150に収まる範囲でできる
// だけ奥)に置く白い城(internal/assets.GrasslandCastle)の位置・高さ。
// 高さ11だと上端の仰角はカメラ中心からおよそ21.5°(垂直画角の半分27.5°の
// うち、画面上端まで約6°分の余白が残る)。この余白に手段A(放射状
// グラデーションのハロー)を後から重ねられる。
const (
	grasslandCastleZ      = -85.0
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

// grasslandObstacleHalfDepth/HalfWidth/Height は、道の途中に置く柵(丸太)の
// 大きさ。馬に乗ってジャンプしないと越えられない障害物として使う
// (grasslandObstacleZs、GrasslandObstacleBlocks参照)。
const (
	grasslandObstacleHalfDepth = 0.4
	grasslandObstacleHalfWidth = 2.4 // 道(半幅1.8)より広く取り、道を完全にふさぐ
	grasslandObstacleHeight    = 1.0
)

// grasslandObstacleZs は、道に沿って並べる柵の中心Z座標。プレイヤーは
// スポーン地点(grasslandLinkZ)から城に向かってZが減る方向へ進むため、
// 手前から奥の順に並んでいる。馬に乗って加速しながら連続してジャンプする
// アスレチックのような区間にするため、単発の障害物ではなく複数個を
// 間隔を空けて配置している(1つ目は徒歩でも見える距離に置き、馬の歌を
// 演奏する必然性を作る)。
var grasslandObstacleZs = []float64{-12.0, -27.0, -42.0, -57.0}

// grasslandObstacleClampMargin は、GrasslandObstacleBlocksが移動を止める際、
// 境界からほんの少しだけ外側(通行可能な側)に押し戻す量。ちょうど境界の
// 座標に止めてしまうと、浮動小数点の境界判定が「まだ障害物の範囲内」と
// 誤判定し続け、そこから離れる方向へも進めなくなる(スタックする)バグが
// あったため、これを避けるための余白。
const grasslandObstacleClampMargin = 0.01

// grasslandCameraEye/Target/Up は、馬に乗る前の見下ろし気味の固定カメラ。
// GrasslandCameraTransitionDuration秒かけて、馬に乗った後は横視点カメラ
// (GrasslandSideCameraEyeTarget)へ滑らかに遷移する(cmd/grassland/main.go
// 参照)。
var (
	grasslandCameraEye    = vecmath.NewVec3(0, 8, 12)
	grasslandCameraTarget = vecmath.NewVec3(0, 0, -10)
	grasslandCameraUp     = vecmath.NewVec3(0, 1, 0)
)

// GrasslandCameraTransitionDuration は、馬に乗ってから横視点カメラへの
// 切り替えが完了するまでの時間(秒)。瞬間的に切り替わると酔いやすい・
// 状況が飲み込みにくいため、この時間をかけてカメラ位置・注視点を線形補間
// する。
const GrasslandCameraTransitionDuration = 1.5

// GrasslandSideCameraDistance/Height/LookHeight は、馬に乗った後のマリオの
// ような横視点カメラのパラメータ。X軸の正方向から見下ろすことで、
// プレイヤーの前後移動(Z軸)がそのまま画面の左右移動に見える構図になり、
// 障害物のジャンプ(Y方向)が縦の動きとしてはっきり見えるようになる
// (アスレチック的な難易度を作りやすくするため)。
const (
	GrasslandSideCameraDistance   = 10.0
	GrasslandSideCameraHeight     = 3.0
	GrasslandSideCameraLookHeight = 1.3
)

// GrasslandDefaultCameraEyeTarget は、馬に乗る前の固定カメラのeye・target
// を返す。カメラ遷移の補間の開始値として使う。
func GrasslandDefaultCameraEyeTarget() (eye, target vecmath.Vec3) {
	return grasslandCameraEye, grasslandCameraTarget
}

// GrasslandSideCameraEyeTarget は、プレイヤーの現在のZ座標(playerZ)に
// 追従する、横視点カメラのeye・targetを返す。
func GrasslandSideCameraEyeTarget(playerZ float64) (eye, target vecmath.Vec3) {
	eye = vecmath.NewVec3(GrasslandSideCameraDistance, GrasslandSideCameraHeight, playerZ)
	target = vecmath.NewVec3(0, GrasslandSideCameraLookHeight, playerZ)
	return eye, target
}

// GrasslandCameraUp は、草原フィールドのカメラが常に使う上方向ベクトル。
func GrasslandCameraUp() vecmath.Vec3 {
	return grasslandCameraUp
}

// grasslandProjection は、草原フィールドの透視投影行列を組み立てる。
// FOV: 55°では城をこれ以上手前・大きくする余白が無くなったため65°に
// 広げ、画面上端までの余白を確保している(木の配置はまだ十分内側にある
// ため、広げても画面外に出ることはない)。
func grasslandProjection(aspect float64) vecmath.Mat4 {
	return vecmath.Perspective(vecmath.Radians(65), aspect, 0.1, 150)
}

// BuildGrasslandScene は、草原フィールドの土台(地面・道・Link)を配置した
// Sceneを組み立てる。戻り値のLinkPlacementは、demo.goの神殿フィールドと
// 同様、プレイヤー移動に合わせて呼び出し側がLinkのTransformを書き換える
// ために使う。projectionは、呼び出し側が馬に乗った後のカメラ遷移で
// 毎フレームScene.ViewProjectionを組み直すために別途返す。
func BuildGrasslandScene(c *Context) (scene *Scene, link LinkPlacement, projection vecmath.Mat4, err error) {
	program, err := c.NewProgram(basicVertexShaderSrc, basicFragmentShaderSrc)
	if err != nil {
		return nil, LinkPlacement{}, vecmath.Mat4{}, err
	}

	width, height := c.CanvasSize()
	aspect := float64(width) / float64(height)
	projection = grasslandProjection(aspect)
	view := vecmath.LookAt(grasslandCameraEye, grasslandCameraTarget, grasslandCameraUp)

	// まず不透明なオブジェクトをすべて配置する(深度テストがあるため、
	// 互いの前後関係は描画順によらず正しく処理される)。
	objects := []Object{grasslandGroundObject(c), grasslandPathObject(c)}

	castle, err := grasslandCastleObject(c)
	if err != nil {
		return nil, LinkPlacement{}, vecmath.Mat4{}, err
	}
	objects = append(objects, castle)

	trees, err := grasslandTreeObjects(c)
	if err != nil {
		return nil, LinkPlacement{}, vecmath.Mat4{}, err
	}
	objects = append(objects, trees...)

	objects = append(objects, grasslandObstacleObjects(c)...)

	linkObj, linkLocal, err := grasslandLinkObject(c)
	if err != nil {
		return nil, LinkPlacement{}, vecmath.Mat4{}, err
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
		return nil, LinkPlacement{}, vecmath.Mat4{}, err
	}
	objects = append(objects, skyObjects...)

	return &Scene{
		Program:        program,
		ViewProjection: projection.Mul(view),
		Objects:        objects,
	}, link, projection, nil
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

// grasslandObstacleColor は、道をふさぐ丸太っぽい茶色。
var grasslandObstacleColor = vecmath.NewVec3(0.45, 0.32, 0.18)

// GrasslandObstacleBoundsAt は、grasslandObstacleZs[index]の柵のZ方向の
// 手前端(near、スポーン地点側)・奥端(far、城側)を返す。GrasslandObstacle
// Blocksの判定・テストで使う。
func GrasslandObstacleBoundsAt(index int) (near, far float64) {
	z := grasslandObstacleZs[index]
	return z + grasslandObstacleHalfDepth, z - grasslandObstacleHalfDepth
}

// GrasslandObstacleCount は、道に配置した柵の数を返す。
func GrasslandObstacleCount() int {
	return len(grasslandObstacleZs)
}

// grasslandObstacleObjects は、grasslandObstacleZsの各位置に、馬に乗って
// ジャンプしないと越えられない柵(直方体のプレースホルダー)を組み立てる。
func grasslandObstacleObjects(c *Context) []Object {
	objects := make([]Object, 0, len(grasslandObstacleZs))
	for _, z := range grasslandObstacleZs {
		verts := boxVertices(grasslandObstacleHalfWidth, grasslandObstacleHeight, grasslandObstacleHalfDepth)
		mesh := c.NewMesh(verts, zeroUVs(8), boxIndices())
		transform := vecmath.Translate(vecmath.NewVec3(0, 0, z))
		objects = append(objects, Object{Mesh: mesh, Transform: transform, Color: grasslandObstacleColor})
	}
	return objects
}

// GrasslandObstacleBlocks は、Linkが1フレームでprevZからnewZへ移動した際に、
// いずれかの柵(grasslandObstacleZs、GrasslandObstacleBoundsAt参照)を
// 横切ったかどうかを判定する。jumpingがtrue(ジャンプ中)の場合は常に通行を
// 許可する。横切っていて、かつジャンプ中でなければ、移動を止めるべき境界の
// 少し外側(grasslandObstacleClampMargin分)のZ座標をblockedZとして返す。
// 境界ちょうどに止めると、次のフレームでも「まだ範囲内」と判定され続けて
// 離れる方向にも進めなくなる(スタックする)ため、必ず範囲の外側に出す。
func GrasslandObstacleBlocks(prevZ, newZ float64, jumping bool) (blockedZ float64, blocked bool) {
	if jumping {
		return 0, false
	}

	lo, hi := newZ, prevZ
	if lo > hi {
		lo, hi = hi, lo
	}

	for i := range grasslandObstacleZs {
		near, far := GrasslandObstacleBoundsAt(i)
		if hi < far || lo > near {
			continue
		}
		if newZ < prevZ {
			// 前進(Zが減る方向)して障害物に入った → 手前の境界の外側で止める。
			return near + grasslandObstacleClampMargin, true
		}
		// 後退(Zが増える方向)して障害物に入った → 奥の境界の外側で止める。
		return far - grasslandObstacleClampMargin, true
	}
	return 0, false
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
		{X: 16, Z: GrasslandTreeTriggerZ, Height: grasslandTreeHeight}, // 道の右奥(次のフィールドへのトリガー地点の目印)
		{X: -5, Z: 2.5, Height: grasslandTreeHeight},                   // 道の左手前(画面左下寄り、視錐台に収まる位置)
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

// horseTargetHeight は、馬(internal/assets.Horse)をこの高さになるよう
// スケールする。
const horseTargetHeight = 2.0

// LinkMountHeight は、馬に乗っているLinkを、馬の背に座っているように見せる
// ための上方向オフセット(ワールド単位)。アニメーション・スキニングが
// 未実装のため、専用の騎乗ポーズは組めず、Linkをこの高さだけ持ち上げる
// 簡易的な表現にとどめている(docs/assets/horse/README.md参照)。
const LinkMountHeight = 1.3

// BuildHorseObject は、草原フィールドで「馬の歌」が演奏された際にSceneへ
// 追加する馬のObjectを組み立てる。Linkと同じ場所(player.Player.Zの位置)に
// 呼び出し、Linkが馬に乗っているように見せるため。アニメーション・
// スキニングは未実装のため、静止ポーズでの表示になる
// (docs/assets/horse/README.md参照)。
//
// 戻り値のlocalTransformは、ワールド位置を含まないモデル原点補正+スケール
// のみの変換で、呼び出し側が毎フレームplayer.Player.Transform(localTransform)
// を使って、Linkと同じ位置・向きに馬を追従させるためのもの。
func BuildHorseObject(c *Context) (obj Object, localTransform vecmath.Mat4, err error) {
	model, err := c.LoadSkinnedGLBMesh(assets.Horse)
	if err != nil {
		return Object{}, vecmath.Mat4{}, err
	}

	localTransform = model.GroundTransform(0, 0, horseTargetHeight)
	obj = Object{Mesh: model.Mesh, Texture: model.Texture, Color: model.Color}
	return obj, localTransform, nil
}
