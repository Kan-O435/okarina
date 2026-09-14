package renderer

import (
	"math/rand"

	"github.com/Kan-O435/okarina/internal/assets"
	"github.com/Kan-O435/okarina/internal/vecmath"
)

// このファイルは「時の神殿」フィールド(docs/fields/temple-of-time.md参照)の
// 土台となるデモシーンを組み立てる。
//
// 自作で完結する部分(地面・道・プール・階段)は実寸で配置する。
// 建物本体+双尖塔はTripo3Dで生成・リメッシュ済みのGLB(internal/assets)を
// go:embedで読み込んで実際に配置している。扉(#3)はまだ3Dモデル自体は無いが、
// CC0のドア画像(internal/assets.DoorTexture)を貼った板で見た目を確保している。

const (
	groundHalfExtent = 500.0 // 地面: カメラのfar(150)より十分大きく、実質無限に見えるサイズにしておく

	pathHalfWidth = 2.0
	pathNearZ     = 22.0  // 道の手前端(スポーン地点)
	pathFarZ      = -16.0 // 道の奥端(階段の手前)

	poolHalfWidth = 2.0
	poolHalfDepth = 4.0
	poolCenterZ   = -10.0
	poolOffsetX   = 5.0 // 道を挟んで左右に配置

	linkSpawnZ       = poolCenterZ + poolHalfDepth // Linkの足元をプール手前端(カメラ側の辺)に揃える
	linkTargetHeight = 1.4                         // KayKit Knightモデルをこの高さになるようスケールする

	stepHalfWidth = 3.0
	stepHalfDepth = 0.5

	buildingTargetHeight = 10.0 // GLBモデルをこの高さになるようスケールする
	buildingCenterZ      = -24.0

	// doorHalfWidth は、扉画像から左右の石枠(doorTextureU0〜U1の外側)を
	// 除いてトリミングした後の見た目の幅に合わせた値(ピクセル密度を
	// doorHeightと揃えて算出: (395-130)px ÷ (512px/doorHeight))。
	doorHalfWidth = 0.775
	doorHeight    = 3.0
	// 扉画像(512x512)は、中央の木の扉の両脇に白っぽい石枠が写っている。
	// U座標をこの範囲に絞ることで、左右の石枠を除いて中央の扉部分だけを
	// 表示する(ピクセル座標130〜395を実測して算出)。
	doorTextureU0 = 130.0 / 512.0
	doorTextureU1 = 395.0 / 512.0
	// 建物本体(Tripo3D生成GLB)は高さ10へ自動スケールすると最前面が
	// 世界座標Z=-20.73付近まで張り出す(bounding boxから逆算した実測値)。
	// 旧プレースホルダーboxの寸法を前提にZ=-20.5としていたため、実モデルに
	// 差し替わった際に扉が建物の中に埋もれて見えなくなっていた。
	// 最終段(Z=-18)と建物の間に確実に収まるよう手前に出す。
	doorCenterZ = -19.0
	doorHingeX  = -doorHalfWidth // 扉が開くときに軸となる蝶番のローカルX座標(左端)
)

// doorOpenAngleRad は扉が全開(progress=1)になったときの回転角。
var doorOpenAngleRad = vecmath.Radians(100)

// BuildFieldDemoScene は「時の神殿」フィールドの土台(地面・道・プール・階段+
// 建物/塔/扉のプレースホルダー)を配置したSceneを組み立てる。
// 戻り値の2番目は、Scene.Objects内での扉Objectのインデックス
// (呼び出し側が毎フレーム扉のTransformを更新できるようにするため)。
func BuildFieldDemoScene(c *Context) (*Scene, int, error) {
	program, err := c.NewProgram(basicVertexShaderSrc, basicFragmentShaderSrc)
	if err != nil {
		return nil, 0, err
	}

	width, height := c.CanvasSize()
	aspect := float64(width) / float64(height)
	projection := vecmath.Perspective(vecmath.Radians(55), aspect, 0.1, 150)
	view := vecmath.LookAt(
		vecmath.NewVec3(0, 6.6, 0.7), // カメラ位置: 神殿(塔)に寄せた高さ・距離。前回よりさらに近づけて調整中
		vecmath.NewVec3(0, 3, -15),   // 注視点: 神殿の手前あたり
		vecmath.NewVec3(0, 1, 0),
	)

	objects := []Object{
		groundObject(c),
		pathObject(c),
		poolObject(c, -poolOffsetX),
		poolObject(c, poolOffsetX),
	}
	objects = append(objects, stepObjects(c)...)

	templeBody, err := templeBodyObject(c)
	if err != nil {
		return nil, 0, err
	}
	objects = append(objects, templeBody)

	link, err := linkObject(c)
	if err != nil {
		return nil, 0, err
	}
	objects = append(objects, link)

	trees, err := treeObjects(c)
	if err != nil {
		return nil, 0, err
	}
	objects = append(objects, trees...)

	// 扉は背景が透過のテクスチャを使うため、後ろにある建物本体などの不透明な
	// オブジェクトがすでに描画された後(=Objectsの最後)に描画する。先に描くと、
	// 扉の透過部分が「まだ何も描かれていない背景色」と合成され、後から描かれる
	// 建物にその部分だけ穴が空いたように見えてしまう。
	doorIndex := len(objects)
	door, err := doorObject(c)
	if err != nil {
		return nil, 0, err
	}
	objects = append(objects, door)

	return &Scene{
		Program:        program,
		ViewProjection: projection.Mul(view),
		Objects:        objects,
	}, doorIndex, nil
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
	mesh := c.NewMesh(verts, quadUVsCropped(doorTextureU0, doorTextureU1), quadIndices())
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
	toWorld := vecmath.Translate(vecmath.NewVec3(0, 0, doorCenterZ))

	return toWorld.Mul(toHinge).Mul(rotate).Mul(fromHinge)
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
