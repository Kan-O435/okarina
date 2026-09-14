package renderer

import (
	"math/rand"

	"github.com/Kan-O435/okarina/internal/assets"
	"github.com/Kan-O435/okarina/internal/vecmath"
)

// このファイルは「時の神殿」フィールド(docs/fields/temple-of-time.md参照)の
// 土台となるデモシーンを組み立てる。
//
// 自作で完結する部分(地面・道・プール)は実寸で配置する。
// 建物本体+双尖塔はTripo3Dで生成・リメッシュ済みのGLB(internal/assets)を
// go:embedで読み込んで実際に配置している。扉(#3)はまだ3Dモデル自体は無いが、
// CC0のドア画像(internal/assets.DoorTexture)を貼った板で見た目を確保している。

const (
	groundHalfExtent = 500.0 // 地面: カメラのfar(150)より十分大きく、実質無限に見えるサイズにしておく

	pathHalfWidth = 2.0
	pathNearZ     = 22.0  // 道の手前端(スポーン地点)
	pathFarZ      = -16.0 // 道の奥端(神殿の手前)

	poolHalfWidth = 2.0
	poolHalfDepth = 4.0
	poolCenterZ   = -10.0
	poolOffsetX   = 5.0 // 道を挟んで左右に配置

	linkSpawnZ       = poolCenterZ + poolHalfDepth // Linkの足元をプール手前端(カメラ側の辺)に揃える
	linkTargetHeight = 1.4                         // KayKit Knightモデルをこの高さになるようスケールする

	buildingTargetHeight = 10.0 // GLBモデルをこの高さになるようスケールする
	// buildingCenterZ: 道の奥端(pathFarZ=-16)との隙間を詰めるため手前に寄せた
	// (実測の建物前面はbuildingCenterZ+buildingHalfDepthMeasuredになる。
	// 沈める(Yをずらす)のではなく、地面(Y=0)に正しく立てたままZだけ動かす)。
	buildingCenterZ = -21.0

	// 扉画像(512x512)は、中央の木の扉の周りに白っぽい石枠・透過の余白が
	// 写っている。実測したところ、木の扉本体はピクセル座標で
	// x=130〜395(左右の石枠を除く)、y=28〜474(上下の透過の余白を除く、
	// PNGは上端がy=0)の範囲。UV座標をこの範囲に絞ることで、扉の絵だけが
	// quadいっぱいに表示されるようにする(下に透過部分があると、そこから
	// 背景の建物本体の壁の色が透けて「白い余白」のように見えてしまうため)。
	doorTextureU0 = 130.0 / 512.0
	doorTextureU1 = 395.0 / 512.0
	// NewImageTextureはUNPACK_FLIP_Y_WEBGLで画像を上下反転して取り込むため、
	// V座標は「1 - (PNGのy座標/512)」で計算する(V=0がPNGの下端に対応)。
	doorTextureV0 = 1.0 - 474.0/512.0 // PNG y=474(扉の下端)に対応
	doorTextureV1 = 1.0 - 28.0/512.0  // PNG y=28(扉の上端)に対応

	// doorHalfWidth/doorHeight は、上記のトリミング後の扉画像(265x446px)を
	// ピクセル密度が縦横で揃うようスケールした見た目のサイズ。
	// (基準: 512px = doorHeightを出す前のスケール0.005859 units/px)
	doorHalfWidth = 0.56
	doorHeight    = 1.9
	// 建物本体(Tripo3D生成GLB)は高さ10へ自動スケールすると最前面が
	// buildingCenterZ+buildingHalfDepthMeasured ≈ -17.73まで張り出す
	// (bounding boxから逆算した実測値)。扉が建物の中に埋もれないよう、
	// 前面よりさらに手前(道の奥端pathFarZ=-16のすぐ内側)に出す。
	doorCenterZ = -16.0
	doorHingeX  = -doorHalfWidth // 扉が開くときに軸となる蝶番のローカルX座標(左端)

	// DoorPassThroughZ は、Linkが扉を「すり抜け終わった」とみなすZ座標
	// (扉の位置doorCenterZより1.5ユニットさらに奥)。ゲームループ
	// (internal/game)が、自動前進中のLinkがこの位置まで進んだかどうかの
	// 判定に使うため、パッケージ外から参照できるようにエクスポートする。
	DoorPassThroughZ = doorCenterZ - 1.5
)

// doorOpenAngleRad は扉が全開(progress=1)になったときの回転角。
var doorOpenAngleRad = vecmath.Radians(100)

// LinkPlacement は、呼び出し側(ゲームループ)がLinkの位置・向きを毎フレーム
// 書き換えるために必要な情報をまとめたもの。
type LinkPlacement struct {
	// Index は、scene.ObjectsのうちLinkに対応する要素のインデックス。
	Index int
	// LocalTransform は、ワールド上の位置・向きを一切含まない、Link自身の
	// 原点(足元・中心)を基準にした変換(スケール+モデル原点補正のみ)。
	// 毎フレーム Translate(worldPos).Mul(RotateY(yaw)).Mul(LocalTransform) の
	// ように組み立て直すことで、位置と向きを独立に更新できる。
	LocalTransform vecmath.Mat4
	// SpawnZ はLinkの初期スポーン位置(Z座標)。
	SpawnZ float64
}

// BuildFieldDemoScene は「時の神殿」フィールドの土台(地面・道・プール+
// 建物本体・扉・Link)を配置したSceneを組み立てる。
// 戻り値のLinkPlacementは、プレイヤー移動に合わせて呼び出し側がLinkの
// Transformを書き換えるために使う。doorIndexは扉Objectのインデックス
// (毎フレーム扉のTransformを更新するために使う)。
func BuildFieldDemoScene(c *Context) (scene *Scene, link LinkPlacement, doorIndex int, err error) {
	program, err := c.NewProgram(basicVertexShaderSrc, basicFragmentShaderSrc)
	if err != nil {
		return nil, LinkPlacement{}, -1, err
	}

	width, height := c.CanvasSize()
	aspect := float64(width) / float64(height)
	projection := vecmath.Perspective(vecmath.Radians(55), aspect, 0.1, 150)
	view := vecmath.LookAt(
		vecmath.NewVec3(0, 6.24, 6.03),  // カメラ位置: 手前の枠が池の手前端の少し手前(Z=-5、緑が少し残る程度)、左右の枠が木のライン(X=±7.8)に合うよう計算
		vecmath.NewVec3(0, 5.72, -8.96), // 注視点: 神殿の手前あたり
		vecmath.NewVec3(0, 1, 0),
	)

	objects := []Object{
		groundObject(c),
		pathObject(c),
		poolObject(c, -poolOffsetX),
		poolObject(c, poolOffsetX),
	}
	templeBody, err := templeBodyObject(c)
	if err != nil {
		return nil, LinkPlacement{}, -1, err
	}
	objects = append(objects, templeBody)

	linkObj, linkLocal, err := linkObject(c)
	if err != nil {
		return nil, LinkPlacement{}, -1, err
	}
	objects = append(objects, linkObj)
	link = LinkPlacement{
		Index:          len(objects) - 1,
		LocalTransform: linkLocal,
		SpawnZ:         linkSpawnZ,
	}

	trees, err := treeObjects(c)
	if err != nil {
		return nil, LinkPlacement{}, -1, err
	}
	objects = append(objects, trees...)

	// 扉は背景が透過のテクスチャを使うため、後ろにある建物本体などの不透明な
	// オブジェクトがすでに描画された後(=Objectsの最後)に描画する。先に描くと、
	// 扉の透過部分が「まだ何も描かれていない背景色」と合成され、後から描かれる
	// 建物にその部分だけ穴が空いたように見えてしまう。
	doorIndex = len(objects)
	door, err := doorObject(c)
	if err != nil {
		return nil, LinkPlacement{}, -1, err
	}
	objects = append(objects, door)

	return &Scene{
		Program:        program,
		ViewProjection: projection.Mul(view),
		Objects:        objects,
	}, link, doorIndex, nil
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

// doorObject は、隠し扉の見た目として、扉画像(internal/assets.DoorTexture、
// CC0)を貼った板(quad)を配置する。3Dモデル自体はまだ無いプレースホルダーだが、
// マゼンタの箱よりも「扉らしく」見えるようにしている。初期状態は全閉
// (DoorTransform(0))。
func doorObject(c *Context) (Object, error) {
	texture, err := c.NewImageTexture(assets.DoorTexture, "image/png")
	if err != nil {
		return Object{}, err
	}

	verts := verticalQuadVertices(doorHalfWidth, doorHeight)
	mesh := c.NewMesh(verts, quadUVsCropped(doorTextureU0, doorTextureV0, doorTextureU1, doorTextureV1), quadIndices())
	return Object{
		Mesh:      mesh,
		Texture:   texture,
		Transform: DoorTransform(0),
		Color:     vecmath.NewVec3(1, 1, 1),
	}, nil
}

// DoorTransform は、扉の開き具合progress(0=全閉、1=全開)に応じた配置行列を
// 返す。扉の片端(X=doorHingeX)を蝶番としてY軸回転させることで、
// 実際の開き戸のような動きになる。
func DoorTransform(progress float64) vecmath.Mat4 {
	angle := doorOpenAngleRad * progress

	toHinge := vecmath.Translate(vecmath.NewVec3(doorHingeX, 0, 0))
	fromHinge := vecmath.Translate(vecmath.NewVec3(-doorHingeX, 0, 0))
	rotate := vecmath.RotateY(angle)
	toWorld := vecmath.Translate(vecmath.NewVec3(0, doorHeight/2, doorCenterZ))

	return toWorld.Mul(toHinge).Mul(rotate).Mul(fromHinge)
}

// linkObject は、KayKit Adventurers(CC0)のKnightモデル
// (internal/assets.LinkKnight)を読み込み、フィールド手前のスポーン地点に
// 立つよう配置する。体パーツごとに分かれた複数メッシュを結合して1体として
// 表示するため LoadSkinnedGLBMesh を使う。アニメーション(Idle/Walking等)は
// GLB内に含まれているが、スキニングは未実装のため現状は静止表示のみ。
//
// 戻り値のlocalTransformは、ワールド座標(0, 0)に置いた場合の変換
// (=スケール+モデル原点補正のみ)で、呼び出し側が毎フレーム位置・向きを
// 更新する際の土台として使う。
func linkObject(c *Context) (obj Object, localTransform vecmath.Mat4, err error) {
	model, err := c.LoadSkinnedGLBMesh(assets.LinkKnight)
	if err != nil {
		return Object{}, vecmath.Mat4{}, err
	}

	localTransform = model.GroundTransform(0, 0, linkTargetHeight)
	transform := vecmath.Translate(vecmath.NewVec3(0, 0, linkSpawnZ)).Mul(localTransform)
	obj = Object{Mesh: model.Mesh, Texture: model.Texture, Transform: transform, Color: model.Color}
	return obj, localTransform, nil
}

// treePlacement は1本の木の配置(中心のXZ座標と目標の高さ)を表す。
// どの木モデルを使うかは、この位置を使う側(treeObjects)で決める。
type treePlacement struct {
	X, Z, Height float64
}

// treePlacements は、木を「池の縁(横辺)に沿った列」と「神殿の裏側の列」の
// 2種類だけに絞って配置する。それ以外の(参道沿い・神殿側面などの)木は
// 意図的に置かない。見た目の再現性のため乱数シードは固定する。
const (
	treeMinHeight = 4.5
	treeMaxExtra  = 3.0
	// buildingHalfDepthMeasured は、建物本体GLB(高さbuildingTargetHeightへ
	// スケール後)の奥行き方向の半分の実測値(bounding boxから算出、
	// doorCenterZのコメント参照)。裏側の並木をこの分だけ建物中心から
	// 離すことで、建物の裏に接する位置に列を作る。
	buildingHalfDepthMeasured = 3.27
)

func treePlacements() []treePlacement {
	rng := rand.New(rand.NewSource(7))
	var placements []treePlacement

	// 池の縁(横辺、Z方向に伸びる辺)に沿って並ぶ木。左右の池それぞれの
	// 外側の縁(poolHalfWidthのすぐ外)から、神殿の手前(doorCenterZ付近)まで
	// 途切れずに続けることで、手前の池と奥の神殿裏の並木との間に
	// 木の無い区間ができないようにする。
	const poolTreeMargin = 0.8
	const poolTreeCount = 8
	const poolTreeZFront = poolCenterZ + poolHalfDepth // 池の手前端(-6)
	const poolTreeZBack = -19.0                        // 神殿の手前際まで
	for _, side := range []float64{-1, 1} {
		outerX := side * (poolOffsetX + poolHalfWidth + poolTreeMargin)
		for i := 0; i < poolTreeCount; i++ {
			t := float64(i) / float64(poolTreeCount-1)
			z := poolTreeZFront + t*(poolTreeZBack-poolTreeZFront)
			placements = append(placements, treePlacement{
				X:      outerX + rng.Float64()*0.6 - 0.3,
				Z:      z + rng.Float64()*0.6 - 0.3,
				Height: treeMinHeight + rng.Float64()*treeMaxExtra,
			})
		}
	}

	// 神殿の裏側(建物の奥の壁のさらに向こう)に並ぶ木。
	const templeBackMargin = 1.5
	const templeBackHalfWidth = 8.0
	const templeBackCount = 9
	backZ := buildingCenterZ - buildingHalfDepthMeasured - templeBackMargin
	for i := 0; i < templeBackCount; i++ {
		t := float64(i) / float64(templeBackCount-1)
		x := -templeBackHalfWidth + t*(templeBackHalfWidth*2)
		placements = append(placements, treePlacement{
			X:      x + rng.Float64() - 0.5,
			Z:      backZ + rng.Float64()*1.5 - 0.75,
			Height: treeMinHeight + rng.Float64()*treeMaxExtra,
		})
	}

	return placements
}

// treeObjects は、CC0の低ポリ木モデル群(internal/assets.Trees、Gobkit Nature
// Kit)を読み込み、treePlacements()の配置に従って並べる。木の種類は同じ
// メッシュ・テクスチャを使い回し、Transformだけを変えて複製する
// (頂点データを毎回コピーしない)。
func treeObjects(c *Context) ([]Object, error) {
	entries, err := assets.Trees.ReadDir("models/trees")
	if err != nil {
		return nil, err
	}

	models := make([]*Model, 0, len(entries))
	for _, entry := range entries {
		data, err := assets.Trees.ReadFile("models/trees/" + entry.Name())
		if err != nil {
			return nil, err
		}
		model, err := c.LoadGLBMesh(data)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	if len(models) == 0 {
		return nil, nil
	}

	placements := treePlacements()
	objects := make([]Object, 0, len(placements))
	for i, p := range placements {
		model := models[i%len(models)]
		transform := model.GroundTransform(p.X, p.Z, p.Height)
		objects = append(objects, Object{Mesh: model.Mesh, Texture: model.Texture, Transform: transform, Color: model.Color})
	}
	return objects, nil
}
