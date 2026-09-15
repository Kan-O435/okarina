package renderer

import (
	"math"
	"math/rand"

	"github.com/Kan-O435/okarina/internal/assets"
	"github.com/Kan-O435/okarina/internal/gltf"
	"github.com/Kan-O435/okarina/internal/vecmath"
)

// このファイルは、エンディング画面のSceneを組み立てる。花畑の上でLinkと
// ゼルダ姫が向かい合って立ち、その2人をまとめて1つの塊とみなして
// ゆっくり回転させる(cmd/ending/main.go参照。タイトル画面のオカリナと
// 同様、位置を含まないローカル変換を呼び出し側が毎フレーム
// RotateY(angle)と組み合わせて使う)。

// endingCharacterHeight は、Link・ゼルダ姫をこの高さになるようスケール
// する(神殿・草原フィールドのlinkTargetHeightと同じ)。
// endingCharacterOffsetXは、中心から左右に離して向かい合わせる距離
// (2人の間隔はこの2倍になる)。
const (
	endingCharacterHeight  = 1.4
	endingCharacterOffsetX = 0.55
)

// endingGroundHalfExtent は、花畑の地面(1枚の板)の半径。
const endingGroundHalfExtent = 12.0

// endingCameraDistance/EyeX/EyeY/LookYは、向かい合う2人をやや上から
// 見下ろすカメラのパラメータ。EyeYをLookYより高くすることで斜め上からの
// アングルになり、EyeXを0からずらすことで真正面ではなく少し斜めから
// 覗き込むような構図にしている。
const (
	endingCameraDistance = 5.5
	endingCameraEyeX     = 1.1
	endingCameraEyeY     = 2.6
	endingCameraLookY    = 0.6
)

// endingPetalCount は、舞い散る花びら(花吹雪)の枚数。
// endingPetalHalfSizeは、花びら1枚(正方形の板)の半径(ワールド単位)。
const (
	endingPetalCount    = 40
	endingPetalHalfSize = 0.09
)

// endingPetalAreaHalfX/MinZ/MaxZは、花びらが漂う範囲(X: 中心から左右に
// この幅、Z: この範囲)。Link・ゼルダ姫・カメラの周り全体に舞うよう、
// 花畑の手前側を広めにカバーしている。
const (
	endingPetalAreaHalfX = 8.0
	endingPetalAreaMinZ  = -8.0
	endingPetalAreaMaxZ  = 4.0
)

// EndingPetalMaxHeight は、花びらが落ち始める高さの最大値(0〜この値の
// ランダムな高さから、それぞれ独立して落ち始める)。
const EndingPetalMaxHeight = 6.0

// endingPetalFallSpeedMin/Maxは、花びらが落下する速さの範囲
// (ワールド単位/秒)。実際の桜吹雪のように、ゆっくりひらひらと舞い落ちる
// 速さにしている。
const (
	endingPetalFallSpeedMin = 0.35
	endingPetalFallSpeedMax = 0.75
)

// EndingScene は、エンディング画面のScene本体に加えて、Link・ゼルダ姫を
// 「1つの塊」として回転させるための情報、および舞い散る花びら
// (花吹雪)の情報をまとめたもの。
type EndingScene struct {
	Scene *Scene

	// PairObjectStartは、Scene.Objects内でLink・ゼルダ姫のパーツが始まる
	// インデックス(地面の後に連続して追加されている)。
	PairObjectStart int

	// PairLocalTransforms[i]は、PairObjectStart+iのパーツに対応する、
	// ワールド回転を含まないローカル変換(向かい合う位置・向き・原点補正・
	// スケールは含むが、2人をまとめて回すRotateYだけは含まない)。
	// 呼び出し側は毎フレーム、共通のRotateY(angle)をこの左からかけて
	// Transformを再構成する(RotateY(angle).Mul(PairLocalTransforms[i]))
	// ことで、向かい合う位置関係を保ったまま2人一緒に回転させる。
	PairLocalTransforms []vecmath.Mat4

	// Petalsは、舞い散る花びら(花吹雪)1枚ごとの静的パラメータ
	// (Scene.Objects内のインデックス+揺れ方・落下速度等)の一覧。
	// 呼び出し側(cmd/ending/main.go)が毎フレーム、これらを基に各花びらの
	// 現在のY座標・回転角を計算してTransformを更新する。
	Petals []EndingPetal
}

// EndingPetal は、花吹雪の花びら1枚ぶんの静的パラメータ。位置・回転角の
// 実行時の状態(現在のY座標・累積回転角)はcmd/ending/main.go側で
// 保持する(このstructはBuildEndingSceneが一度だけ決める「動かない」値
// のみを持つ)。
type EndingPetal struct {
	// ObjectIndexは、Scene.Objects内でこの花びらに対応するインデックス。
	ObjectIndex int
	// X, Zは、左右・前後の揺れの中心位置。
	X, Z float64
	// StartYは、最初に落ち始める高さ。
	StartY float64
	// FallSpeedは、Y方向に落下する速さ(ワールド単位/秒)。
	FallSpeed float64
	// SwayAmplitude/Frequency/Phaseは、左右にゆらゆら揺れながら落ちる
	// 動き(X = X + sin(t*Frequency+Phase)*Amplitude)のパラメータ。
	SwayAmplitude, SwayFrequency, SwayPhase float64
	// SpinSpeedは、花びら自身がくるくる回る速さ(ラジアン/秒、正負で
	// 回転方向が変わる)。
	SpinSpeed float64
}

// BuildEndingScene は、花畑の地面+向かい合うLink・ゼルダ姫のSceneを
// 組み立てる。
func BuildEndingScene(c *Context) (*EndingScene, error) {
	program, err := c.NewProgram(basicVertexShaderSrc, basicFragmentShaderSrc)
	if err != nil {
		return nil, err
	}

	width, height := c.CanvasSize()
	aspect := float64(width) / float64(height)
	projection := vecmath.Perspective(vecmath.Radians(45), aspect, 0.1, 100)
	view := vecmath.LookAt(
		vecmath.NewVec3(endingCameraEyeX, endingCameraEyeY, endingCameraDistance),
		vecmath.NewVec3(0, endingCameraLookY, 0),
		vecmath.NewVec3(0, 1, 0),
	)

	groundTexture, err := c.NewImageTexture(assets.FlowerFieldTexture, "image/png")
	if err != nil {
		return nil, err
	}
	groundMesh := c.NewMesh(
		quadVertices(endingGroundHalfExtent, endingGroundHalfExtent, 0),
		quadUVsCropped(0, 0, 1, 1),
		quadIndices(),
	)
	objects := []Object{{
		Mesh:      groundMesh,
		Texture:   groundTexture,
		Transform: vecmath.Identity(),
		Color:     vecmath.NewVec3(1, 1, 1),
	}}

	// 神殿・草原フィールドと同じ雲(internal/assets.TempleCloud)を空に
	// 浮かべる。エンディングのカメラは他フィールドよりずっと近く・下向きの
	// アングルのため、demo.go/grassland.go側の配置(遠景・高空向け)を
	// そのまま使うと画角の外に出てしまう。endingCloudPlacementsで、この
	// カメラに合わせた近く・低めの配置にしている。
	clouds, err := endingCloudObjects(c)
	if err != nil {
		return nil, err
	}
	objects = append(objects, clouds...)

	linkParts, err := c.LoadGLBParts(assets.LinkKnight)
	if err != nil {
		return nil, err
	}
	zeldaParts, err := c.LoadGLBParts(assets.Zelda)
	if err != nil {
		return nil, err
	}

	// x=0, z=0で求めることで、ワールド位置・向きを含まない「原点中心」の
	// 変換(モデル原点補正+スケールのみ)になる(Model.GroundTransform参照)。
	linkCentered := CombinedGroundTransform(linkParts, 0, 0, endingCharacterHeight)
	zeldaCentered := CombinedGroundTransform(zeldaParts, 0, 0, endingCharacterHeight)

	// Linkは中心より-X側に立ち、+X側(ゼルダのいる方)を向く。ゼルダは+X側に
	// 立ち、-X側(Linkのいる方)を向く。実際のモデルの正面の向きに合わせて
	// 実機確認の上で符号を決めている(理論値どおりだと背中合わせになった
	// ため反転させた)。
	linkLocal := vecmath.Translate(vecmath.NewVec3(-endingCharacterOffsetX, 0, 0)).
		Mul(vecmath.RotateY(math.Pi / 2)).
		Mul(linkCentered)
	zeldaLocal := vecmath.Translate(vecmath.NewVec3(endingCharacterOffsetX, 0, 0)).
		Mul(vecmath.RotateY(-math.Pi / 2)).
		Mul(zeldaCentered)

	pairStart := len(objects)
	pairLocalTransforms := make([]vecmath.Mat4, 0, len(linkParts)+len(zeldaParts))
	for _, p := range linkParts {
		objects = append(objects, Object{Mesh: p.Mesh, Texture: p.Texture, Transform: linkLocal, Color: p.Color})
		pairLocalTransforms = append(pairLocalTransforms, linkLocal)
	}
	for _, p := range zeldaParts {
		objects = append(objects, Object{Mesh: p.Mesh, Texture: p.Texture, Transform: zeldaLocal, Color: p.Color})
		pairLocalTransforms = append(pairLocalTransforms, zeldaLocal)
	}

	petalTexture, err := c.NewImageTexture(assets.PetalTexture, "image/png")
	if err != nil {
		return nil, err
	}
	petalMesh := c.NewMesh(
		centeredQuadVertices(endingPetalHalfSize, endingPetalHalfSize),
		quadUVsCropped(0, 0, 1, 1),
		quadIndices(),
	)

	petals := make([]EndingPetal, 0, endingPetalCount)
	for i := 0; i < endingPetalCount; i++ {
		objects = append(objects, Object{Mesh: petalMesh, Texture: petalTexture, Transform: vecmath.Identity(), Color: vecmath.NewVec3(1, 1, 1)})
		petals = append(petals, EndingPetal{
			ObjectIndex:   len(objects) - 1,
			X:             randRange(-endingPetalAreaHalfX, endingPetalAreaHalfX),
			Z:             randRange(endingPetalAreaMinZ, endingPetalAreaMaxZ),
			StartY:        randRange(0, EndingPetalMaxHeight),
			FallSpeed:     randRange(endingPetalFallSpeedMin, endingPetalFallSpeedMax),
			SwayAmplitude: randRange(0.2, 0.6),
			SwayFrequency: randRange(0.5, 1.3),
			SwayPhase:     randRange(0, 2*math.Pi),
			SpinSpeed:     randSignedRange(1.0, 3.0),
		})
	}

	return &EndingScene{
		Scene: &Scene{
			Program:        program,
			ViewProjection: projection.Mul(view),
			Objects:        objects,
		},
		PairObjectStart:     pairStart,
		PairLocalTransforms: pairLocalTransforms,
		Petals:              petals,
	}, nil
}

// randRange は[min, max)の一様乱数を返す。
func randRange(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// randSignedRange は、絶対値が[min, max)の範囲でランダムな正または負の
// 値を返す(花びらの回転方向をランダムにするため)。
func randSignedRange(min, max float64) float64 {
	v := randRange(min, max)
	if rand.Intn(2) == 0 {
		return -v
	}
	return v
}

// endingCloudPlacements は、エンディング画面の空に浮かべる雲の配置一覧。
// demo.go側のtempleCloudPlacementsと同じ考え方(PartIndexで
// internal/assets.TempleCloudのどのメッシュを使うか指定、Heightで高さを
// 揃える)。ユーザー提供の参考画像(神殿フィールドの、地平線に沿って低く
// 横に伸びる雲の帯)に寄せて、高く浮かべるのではなく低め・遠めに置き、
// 遠近感で地平線付近に圧縮されて見えるようにしている。左右に広く
// (X: -12.5〜12.5)散らして、画面の一角だけでなく空全体に雲が見える
// ようにしている。
var endingCloudPlacements = []templeCloudPlacement{
	{PartIndex: 0, X: -9.0, Y: 1.2, Z: -28, Height: 1.3},
	{PartIndex: 3, X: -5.0, Y: 1.6, Z: -34, Height: 1.6},
	{PartIndex: 5, X: -1.5, Y: 1.1, Z: -26, Height: 1.2},
	{PartIndex: 8, X: 2.0, Y: 1.8, Z: -36, Height: 1.5},
	{PartIndex: 11, X: 5.5, Y: 1.3, Z: -29, Height: 1.3},
	{PartIndex: 14, X: 9.0, Y: 1.6, Z: -33, Height: 1.5},
	{PartIndex: 2, X: -12.5, Y: 2.0, Z: -40, Height: 1.7},
	{PartIndex: 6, X: 12.5, Y: 2.1, Z: -41, Height: 1.7},
}

// endingCloudObjects は、internal/assets.TempleCloudから
// endingCloudPlacementsで指定した雲を選んで空に配置する。demo.goの
// templeCloudObjectsと同じ実装(実体のあるメッシュのため、透過テクスチャ
// のような描画順の考慮は不要)。
func endingCloudObjects(c *Context) ([]Object, error) {
	prims, err := gltf.ParseParts(assets.TempleCloud)
	if err != nil {
		return nil, err
	}

	objects := make([]Object, 0, len(endingCloudPlacements))
	for _, p := range endingCloudPlacements {
		if p.PartIndex < 0 || p.PartIndex >= len(prims) {
			continue
		}
		model, err := c.buildModel(&prims[p.PartIndex])
		if err != nil {
			return nil, err
		}
		localTransform := model.GroundTransform(0, 0, p.Height)
		transform := vecmath.Translate(vecmath.NewVec3(p.X, p.Y, p.Z)).Mul(localTransform)
		objects = append(objects, Object{Mesh: model.Mesh, Texture: model.Texture, Transform: transform, Color: model.Color})
	}
	return objects, nil
}
