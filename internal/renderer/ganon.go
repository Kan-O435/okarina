package renderer

import (
	"github.com/Kan-O435/okarina/internal/assets"
	"github.com/Kan-O435/okarina/internal/gltf"
	"github.com/Kan-O435/okarina/internal/vecmath"
)

// このファイルは、草原の先にある「ガノンフィールド」のデモシーンを組み立てる。
// 荒れ果てた戦場跡のジオラマ(internal/assets.GanonBattleScene、Tripo3Dで
// 生成したオリジナルデザイン)を背景に据え、その手前に地面を広く敷いて、
// ボス戦のアリーナらしく自由に動き回れる土台にする(demo.go/grassland.goと
// 同じ、まず地面+背景+Linkの最小構成から始める方針)。

// ganonBackgroundZ/Height は、背景ジオラマ(戦場跡案)の位置・高さ。この
// モデルは横長・低め(幅:高さ:奥行 ≈ 3.1:1:2.7)の「舞台」状の形をしている。
const (
	ganonBackgroundZ      = -40.0
	ganonBackgroundHeight = 10.0
)

// ganonHallHeight は、玉座の間案(internal/assets.GanonHall)をこの高さに
// なるようスケールする。このモデルは戦場跡案より立方体に近い縦横比
// (幅:高さ:奥行 ≈ 1.4:1:0.9)のため、同じZ・同じ高さに置いても見え方は
// かなり異なる(奥行きが浅く、横幅も狭い)。カメラは戦場跡案に合わせて
// 調整済みのため、この案を見る際は画角の余白が大きく空く可能性がある。
const ganonHallHeight = 10.0

// ganonGroundHalfExtent は地面の半径。背景ジオラマの床だけだと手前
// (カメラ側)が手狭になるため、他のフィールドと同様に大きく取り、
// 背景の手前まで地面を広げて自由に動き回れるアリーナにしている。
const ganonGroundHalfExtent = 300.0

// ganonGroundColor は、背景ジオラマの暖色(オレンジ〜赤茶)の岩肌に
// 合わせた地面の色。
var ganonGroundColor = vecmath.NewVec3(0.4, 0.2, 0.14)

// ganonLinkZ/Y は、Linkをカメラの手前・ボスの少し手前に立たせておく
// 位置。
//
// カメラ(Y=6で水平を見る、奥の壁用に画角を狭めてある)は見下ろさない
// ため、地面に立つ低いキャラクターほど近距離では画面下に切れてしまう
// (距離dでの画面下端のYは6-d*tan(垂直画角の半分)で、Link(高さ1.4)を
// 画面に収めるには足元がその値以下である必要がある)。カメラのすぐ
// 手前(Z=-24、距離4)だと必要距離が約19にもなり、Linkがほぼ画面外に
// なってしまうため、ボスと同様にジオラマの盛り上がった地形の上に乗せる
// ことで必要距離を下げている。ボス(Z=-38)の3手前のZ=-35を採用。
// GLBの生データから見積もった高さ(1.78)ではまだ画面下に隠れて見えな
// かったため、実際にレンダリングして確認しながらganonLinkYを4.0まで
// 上げている(この付近の地形は起伏があり、点のサンプリングだけでは
// 正確な高さを見積もりにくいため、最終的には目視で調整した)。
// ganonLinkHeight は、戦場跡案でのLinkの高さ。他フィールド共通の
// linkTargetHeight(1.4)だとボスの近くではやや大きく見えたため、
// 少し小さめの1.1にしている。
const (
	ganonLinkZ      = -26.0
	ganonLinkY      = 4.0
	ganonLinkHeight = 1.1
)

// ganonLinkHallZ は、玉座の間案でのLinkの位置。カメラは背景(建物)全体を
// 映すことを優先し、水平視線(チルト無し、Y=5)のままにしているため、
// 足元(Y=0)がフレームに収まるにはカメラから垂直画角(半分30°)ぶんの
// 距離(5÷tan(30°)≈8.66)以上離す必要がある。この制約の中でできるだけ
// 手前(カメラ寄り)にするため、ちょうどこの境界のZ=-37.5にしている。
// ganonLinkHallY は、Linkの足元を少し持ち上げるオフセット。建物の床
// (装飾等)に埋もれて見えたため、ganonBossHallYと同様に少し浮かせている。
// ganonLinkHallHeight は、玉座の間案でのLinkの高さ。他フィールド共通の
// linkTargetHeight(1.4)より小さめの1.0にしている。
const (
	ganonLinkHallZ      = -37.5
	ganonLinkHallY      = 0.3
	ganonLinkHallHeight = 1.0
)

// ganonBossBattleZ/Y/Height は、戦場跡案でのボス(internal/assets.
// GanonBossBattle、緑色・角のある獣形態)の位置・高さ。このモデルは
// バインドポーズが前傾姿勢(しゃがんで武器を構えたようなポーズ)で
// 書き出されているため、CombinedGroundTransformで足元をY=0に置くと
// 手前のLinkや地形の起伏にほぼ隠れてしまう。ganonLinkY(=4.0)と同様、
// この付近の地形が高く盛り上がっているため、はっきり見える高さまで
// 持ち上げてY=3.05にしている。
const (
	ganonBossBattleZ      = -38.0
	ganonBossBattleY      = 3.05
	ganonBossBattleHeight = 3.6
)

// ganonBossHallZ/Y は、玉座の間案(GanonBackgroundHall)でのボスの位置。
// 玉座の間モデル(ganonHallHeight=10へスケール後)は半奥行4.49で、
// ganonBackgroundZ(-40)を中心にZ≈-35.51(手前の入口側)〜Z≈-44.49
// (奥の壁側)の範囲を占める。ganonBossZ(-33)は手前の面より外(建物の
// 外)にあたるため、玉座の間案では代わりに建物の中に収まるZを使う
// (-41から、もう少し手前(入口側)へ寄せて-38.5)。床は平らだと仮定して
// Y=0から始めたが、建物側の床(装飾等)に埋まって見えたため、少し
// (0.4)持ち上げている。
// ganonBossHallHeight は、玉座の間案でのボスの高さ。戦場跡案(ganonBossHeight
// =3.6)より小さめの2.8にしている(建物の中で見ると3.6は大きすぎたため)。
const (
	ganonBossHallZ      = -38.5
	ganonBossHallY      = 0.9
	ganonBossHallHeight = 2.2
)

// GanonBackgroundVariant は、ガノンフィールドの背景デザイン案(戦場跡/
// 玉座の間)を選ぶための識別子。それぞれ別ページ・別Link
// (cmd/ganon-battle, cmd/ganon-hall)から固定で1つを指定して使う。
type GanonBackgroundVariant int

const (
	// GanonBackgroundBattle は現行案: 荒れ果てた戦場跡のジオラマ
	// (internal/assets.GanonBattleScene)。
	GanonBackgroundBattle GanonBackgroundVariant = iota
	// GanonBackgroundHall は前案: 玉座の間(internal/assets.GanonHall、
	// Tripo3Dで再生成したオリジナルデザイン)。
	GanonBackgroundHall
)

// BuildGanonScene は、ガノンフィールドの土台(地面・背景・Link)を配置した
// Sceneを組み立てる。variantで背景デザイン案(戦場跡/玉座の間)を選べる。
// 戻り値のLinkPlacementは、他のフィールドと同様、プレイヤー移動に合わせて
// 呼び出し側がLinkのTransformを書き換えるために使う。
func BuildGanonScene(c *Context, variant GanonBackgroundVariant) (scene *Scene, link LinkPlacement, err error) {
	program, err := c.NewProgram(basicVertexShaderSrc, basicFragmentShaderSrc)
	if err != nil {
		return nil, LinkPlacement{}, err
	}

	width, height := c.CanvasSize()
	aspect := float64(width) / float64(height)
	projection := vecmath.Perspective(vecmath.Radians(ganonVerticalFOV(variant)), aspect, 0.1, 150)
	view := ganonCameraView(variant)

	objects := []Object{ganonGroundObject(c)}

	background, err := ganonBackgroundObjects(c, variant)
	if err != nil {
		return nil, LinkPlacement{}, err
	}
	objects = append(objects, background...)

	boss, err := ganonBossObject(c, variant)
	if err != nil {
		return nil, LinkPlacement{}, err
	}
	objects = append(objects, boss...)

	linkSpawnZ, linkY, linkHeight := ganonLinkZ, ganonLinkY, ganonLinkHeight
	if variant == GanonBackgroundHall {
		linkSpawnZ, linkY, linkHeight = ganonLinkHallZ, ganonLinkHallY, ganonLinkHallHeight
	}

	linkObjs, linkLocal, err := ganonLinkObject(c, linkSpawnZ, linkY, linkHeight)
	if err != nil {
		return nil, LinkPlacement{}, err
	}
	objects, link = appendLinkObjects(objects, linkObjs, linkLocal, linkSpawnZ)

	return &Scene{
		Program:        program,
		ViewProjection: projection.Mul(view),
		Objects:        objects,
	}, link, nil
}

// ganonHallCameraY/Z は、玉座の間案(GanonBackgroundHall)専用のカメラ位置。
// この案は戦場跡案よりコンパクトな縦横比(幅:高さ:奥行 ≈ 1.4:1:0.9)のため、
// 戦場跡案用のカメラのままだと横幅・手前の面がフレームに収まらない。
//
// 実測(ganonHallHeight=10へスケール後): 半幅6.91、半奥行4.49、
// 手前の面(カメラに最も近い面)はganonBackgroundZ+4.49 ≈ -35.51。
// 横幅を画面いっぱいに収めるだけなら距離8.98で足りるが、より近づいた
// 構図にするため、横幅が画面の外に少しはみ出るのを許容して距離6.7まで
// 詰めている(camZ = 手前の面(-35.51) + 6.7 ≈ -28.8)。
//
// ganonHallCameraTargetY は注視点の高さ。一時的に見下ろすチルトを付けて
// 試したが、それだと建物の上部がフレームからはみ出し「背景全体を映す」
// 条件を満たさなくなるため、カメラと同じY=5(水平視線・チルト無し)に
// 戻している。Linkの距離はganonLinkHallZ側で、この水平視線のまま
// 全身が入る最短距離に調整する。
const (
	ganonHallCameraY       = 5.0
	ganonHallCameraZ       = -28.8
	ganonHallCameraTargetY = 5.0
)

// ganonBattleFOVDegrees は、戦場跡案において「奥の壁の横幅がちょうど画面の
// 横幅になる」よう逆算した垂直画角。
//
// 実測(ganonBackgroundHeight=10へスケール後): 半幅15.49、奥の壁は
// ganonBackgroundZ(-40)からさらに奥へ半奥行13.49進んだZ≈-53.49。
// カメラ(Z=-17)から奥の壁までの距離は36.49。半幅/距離=tan(水平半画角)より
// 水平画角は約46.0°、canvas比4:3から垂直画角は約35.3°(=水平画角÷アスペクト比を
// 逆算)。デフォルトの60°のままだと奥の壁は横幅の半分強しか画面を
// 占めないため、この値まで画角を狭めて奥の壁を画面いっぱいに広げている。
const ganonBattleFOVDegrees = 35.3

// ganonHallFOVDegrees は、玉座の間案のカメラ(ganonHallCameraZ)を導出した
// 際に前提とした垂直画角。戦場跡案とは別の画角を使うため、変更しないこと。
const ganonHallFOVDegrees = 60.0

// ganonVerticalFOV は、variantに応じた垂直画角(度)を返す。戦場跡案・
// 玉座の間案はカメラ距離の前提となる画角が異なるため、案ごとに使い分ける。
func ganonVerticalFOV(variant GanonBackgroundVariant) float64 {
	if variant == GanonBackgroundHall {
		return ganonHallFOVDegrees
	}
	return ganonBattleFOVDegrees
}

// ganonCameraView は、variantに応じたカメラ視点(LookAt行列)を返す。
// 戦場跡案・玉座の間案は縦横比が大きく異なるため、案ごとに専用のカメラを
// 用意している。
func ganonCameraView(variant GanonBackgroundVariant) vecmath.Mat4 {
	if variant == GanonBackgroundHall {
		return vecmath.LookAt(
			vecmath.NewVec3(0, ganonHallCameraY, ganonHallCameraZ),
			vecmath.NewVec3(0, ganonHallCameraTargetY, ganonBackgroundZ),
			vecmath.NewVec3(0, 1, 0),
		)
	}

	// 戦場跡案: 幅に対して高さが低い(横長)ため、縦方向を画面いっぱいに
	// しようとすると横方向は画面からはみ出すほど寄る必要がある。カメラの
	// 高さを背景の縦方向の中心(Y=5、高さ10の半分)に合わせたうえで寄せ、
	// 上下方向の隙間が出ないようにしている。
	//
	// カメラをZ方向に3だけ前(奥へ)寄せている(視点・注視点を同じだけ
	// 動かし、視線の向き自体は変えていない)。X=0のまま(ボスもX=0に
	// 立っているため、これでボスのほぼ正面に来る)。Yも5→6に上げている
	// (視点・注視点とも同じだけ、水平視線のまま)。
	return vecmath.LookAt(
		vecmath.NewVec3(0, ganonBattleCameraY, ganonBattleCameraZ),
		vecmath.NewVec3(0, ganonBattleCameraY, -38),
		vecmath.NewVec3(0, 1, 0),
	)
}

// ganonBattleCameraY/Z は、戦場跡案のカメラ位置(上のganonCameraView参照)。
// GanonHallCameraY/Z/GanonBattleCameraY/Zとして公開しているのは、崩落演出の
// 「wipe岩」(カメラのすぐ正面に置く岩。cmd/ganon-hall, cmd/ganon-battle
// 参照)をどちらのフィールドでもカメラ位置基準で配置できるようにするため。
const (
	ganonBattleCameraY = 6.0
	ganonBattleCameraZ = -20.0
)

// GanonHallCameraY/Z, GanonBattleCameraY/Z は、それぞれのカメラのワールド
// 位置(上のganonCameraView参照)。cmd/ganon-hall, cmd/ganon-battleが、
// 崩落演出の「wipe岩」をカメラのすぐ正面に置くために参照する
// (renderer.DoorPassThroughZと同様、呼び出し側が必要とする1つの座標だけを
// 公開する形)。
const (
	GanonHallCameraY   = ganonHallCameraY
	GanonHallCameraZ   = ganonHallCameraZ
	GanonBattleCameraY = ganonBattleCameraY
	GanonBattleCameraZ = ganonBattleCameraZ
)

func ganonGroundObject(c *Context) Object {
	verts := quadVertices(ganonGroundHalfExtent, ganonGroundHalfExtent, 0)
	mesh := c.NewMesh(verts, zeroUVs(4), quadIndices())
	return Object{Mesh: mesh, Transform: vecmath.Identity(), Color: ganonGroundColor}
}

// ganonBackgroundObjects は、variantに応じた背景(戦場跡案/玉座の間案)を
// ganonBackgroundZの位置に配置する。戦場跡案はTripo3Dのセグメンテーション
// 機能で「中央にいた人物・怪物」を削除した後のモデルのため、パーツごとに
// 別メッシュ・別テクスチャへ分かれており、1つのObjectにまとめられない
// (LoadGLBPartsで読み、パーツ数ぶんのObjectを返す)。玉座の間案は通常の
// 単一メッシュのモデルなので、これまでどおり1つのObjectで済む。
func ganonBackgroundObjects(c *Context, variant GanonBackgroundVariant) ([]Object, error) {
	if variant == GanonBackgroundHall {
		model, err := c.LoadGLBMesh(assets.GanonHall)
		if err != nil {
			return nil, err
		}
		transform := model.GroundTransform(0, ganonBackgroundZ, ganonHallHeight)
		return []Object{{Mesh: model.Mesh, Texture: model.Texture, Transform: transform, Color: model.Color}}, nil
	}

	parts, err := c.LoadGLBParts(assets.GanonBattleScene)
	if err != nil {
		return nil, err
	}

	transform := CombinedGroundTransform(parts, 0, ganonBackgroundZ, ganonBackgroundHeight)
	objects := make([]Object, len(parts))
	for i, part := range parts {
		objects[i] = Object{Mesh: part.Mesh, Texture: part.Texture, Transform: transform, Color: part.Color}
	}
	return objects, nil
}

// ganonBossObject は、背景(戦場跡の地形/玉座の間の建物内)の中にボスを
// 配置する。variantに応じてモデル・Z/Y位置を切り替える(戦場跡案は
// internal/assets.GanonBossBattle(緑色・角のある獣形態)をganonBossBattle
// Z/Y、玉座の間案はinternal/assets.GanonBoss(人型のガノンドロフ)を建物の
// 中に収まるganonBossHallZ/Yに配置する)。どちらもSketchfabのファンアート
// モデルで、スキンは無くパーツごとに別メッシュ・別テクスチャへ分かれて
// いるため、背景ジオラマ(ganonBackgroundObjects参照)と同様にLoadGLBParts
// (ここではgltf.ParseParts+buildModel)/CombinedGroundTransformで読み込む。
// CombinedGroundTransformは足元をY=0に置くため、そのあとにY分だけ
// ワールド空間で持ち上げる。現時点では静止しているだけで、専用の行動・
// アニメーションはまだ無い。
func ganonBossObject(c *Context, variant GanonBackgroundVariant) ([]Object, error) {
	bossAsset := assets.GanonBossBattle
	bossZ, bossY, bossHeight := ganonBossBattleZ, ganonBossBattleY, ganonBossBattleHeight
	if variant == GanonBackgroundHall {
		bossAsset = assets.GanonBoss
		bossZ, bossY, bossHeight = ganonBossHallZ, ganonBossHallY, ganonBossHallHeight
	}

	// gltf.ParseParts()は各パーツの元ノードのワールド変換行列(位置・回転・
	// スケール)を頂点に焼き込んで返すため、パーツごとに軸や縮尺が異なる
	// モデル(装備品パーツが体とは別の行列を持つ等)でも正しく組み上がる。
	prims, err := gltf.ParseParts(bossAsset)
	if err != nil {
		return nil, err
	}

	parts := make([]*Model, len(prims))
	for i := range prims {
		model, err := c.buildModel(&prims[i])
		if err != nil {
			return nil, err
		}
		parts[i] = model
	}

	// gltf.ParseParts()がノードのワールド変換行列を焼き込むようになった
	// ことで、GanonBoss/GanonBossBattleとも元から+Z(カメラ・Linkのいる方)
	// を向いた状態で読み込まれるため、追加の回転は不要になった
	// (以前はここでSketchfabのconverted形式向けの補正回転を入れていた)。
	localTransform := CombinedGroundTransform(parts, 0, 0, bossHeight)
	transform := vecmath.Translate(vecmath.NewVec3(0, bossY, bossZ)).Mul(localTransform)
	objects := make([]Object, len(parts))
	for i, part := range parts {
		objects[i] = Object{Mesh: part.Mesh, Texture: part.Texture, Transform: transform, Color: part.Color}
	}
	return objects, nil
}

// ganonLinkObject は、神殿・草原フィールドと同じLinkモデル
// (internal/assets.LinkKnight)をspawnZの位置に配置する。他のフィールドの
// linkObjectと同様、localTransformを別に返し、プレイヤーの位置・向きと
// 組み合わせて毎フレームTransformを再構築できるようにする。
func ganonLinkObject(c *Context, spawnZ, spawnY, targetHeight float64) (objs []Object, localTransform vecmath.Mat4, err error) {
	parts, err := loadLinkParts(c)
	if err != nil {
		return nil, vecmath.Mat4{}, err
	}

	// spawnYは、戦場跡案でジオラマの盛り上がった地形の上に立たせるための
	// ワールド空間の底上げ(ganonLinkY参照。玉座の間案では0)。player.State.
	// Transform()が組み立てる最終的な配置にも常に反映され続ける必要があるため
	// (ジャンプ・歩行バウンド等のY方向オフセットと同様に一時的なものではなく、
	// 固定の底上げ)、呼び出し側が毎フレーム使うlocalTransform自体に焼き込む。
	localTransform = vecmath.Translate(vecmath.NewVec3(0, spawnY, 0)).Mul(CombinedGroundTransform(parts, 0, 0, targetHeight))
	// 初期姿勢(スポーン時点でカメラの逆を向く)はplayer.Player.SpawnAtが
	// 設定するYaw=πを使ってcmd/ganon-battle, cmd/ganon-hall側が
	// SetLinkTransformで組み立て直すため、ここでは仮の向き(Identity)で
	// 構わない。
	transform := vecmath.Translate(vecmath.NewVec3(0, 0, spawnZ)).Mul(localTransform)
	objs = make([]Object, len(parts))
	for i, part := range parts {
		objs[i] = Object{Mesh: part.Mesh, Texture: part.Texture, Transform: transform, Color: part.Color}
	}
	return objs, localTransform, nil
}

// LoadRockDebrisParts は、崩落演出・ページ遷移のwipe岩で使う岩モデル
// (internal/assets.RockDebris)の全パーツ(19個、大小さまざま)を読み込む。
// 呼び出し側(cmd/ganon-hall, cmd/ganon-battle)は起動時に1回だけ呼び、
// 結果を毎回のGanonHallCollapseRockObject/GanonWipeRockObjectで使い回す。
func LoadRockDebrisParts(c *Context) ([]*Model, error) {
	prims, err := gltf.ParseParts(assets.RockDebris)
	if err != nil {
		return nil, err
	}
	parts := make([]*Model, len(prims))
	for i := range prims {
		model, err := c.buildModel(&prims[i])
		if err != nil {
			return nil, err
		}
		parts[i] = model
	}
	return parts, nil
}

// ganonHallCollapseRockPartIndices は、崩落演出で降らせる岩(下の
// GanonHallCollapseRockPlacements)に割り当てるパーツ(RockDebrisの
// インデックス、大小混ぜている)。Placementsと同じ数だけ順番に対応させる。
var ganonHallCollapseRockPartIndices = []int{4, 8, 12, 1, 9, 16}

// GanonHallCollapseRockPlacement は、崩落演出で降らせる岩1個ぶんの
// X/Z位置(固定)と半サイズ(見た目の高さ=HalfSize*2になるよう
// スケールする)。Yは時間経過で上から下へ動かすため、ここには含めない
// (cmd/ganon-hall/main.goが毎フレーム計算する)。
type GanonHallCollapseRockPlacement struct {
	X, Z, HalfSize float64
}

// GanonHallCollapseRockPlacements は、Ganonの第一形態を倒した演出
// (玉座の間が崩れ、岩が降ってくる → 戦場跡フィールドへページ遷移)で
// 降らせる岩の配置一覧。「特定の演奏でボスを倒す」処理はまだ無いため、
// 現時点ではデバッグボタン(cmd/ganon-hall/main.go、web/ganon-hall.html
// 参照)から直接トリガーする仮実装。玉座周辺(ganonBossHallZ付近)に
// ばらけて降るよう、位置・サイズを少しずつ変えている。
var GanonHallCollapseRockPlacements = []GanonHallCollapseRockPlacement{
	{X: -3.0, Z: -33.0, HalfSize: 1.0},
	{X: 2.2, Z: -35.0, HalfSize: 1.3},
	{X: -1.3, Z: -38.0, HalfSize: 0.9},
	{X: 3.4, Z: -37.0, HalfSize: 1.1},
	{X: 0.2, Z: -34.0, HalfSize: 1.4},
	{X: -3.4, Z: -39.0, HalfSize: 1.0},
}

// GanonHallCollapseRockObject は、崩落演出用の岩1個ぶんのObjectと
// localTransform(原点中心・見た目の高さ=HalfSize*2に正規化・パーツごとに
// 少しずつ向きを変えて単調にならないようにしたもの)を組み立てる。
// 呼び出し側は毎フレーム Translate(p.X, y, p.Z).Mul(localTransform) で
// Transformを更新する(ganonLinkObject等と同じ構成)。
// partsはLoadRockDebrisPartsの結果、seedIndexはPlacements内での通し番号
// (ganonHallCollapseRockPartIndicesの選択・向きのバリエーションに使う)。
func GanonHallCollapseRockObject(parts []*Model, seedIndex int, p GanonHallCollapseRockPlacement) (obj Object, localTransform vecmath.Mat4) {
	part := parts[ganonHallCollapseRockPartIndices[seedIndex%len(ganonHallCollapseRockPartIndices)]]
	ground := part.GroundTransform(0, 0, p.HalfSize*2)
	yaw := vecmath.RotateY(vecmath.Radians(float64(seedIndex) * 53.0))
	localTransform = yaw.Mul(ground)
	obj = Object{Mesh: part.Mesh, Texture: part.Texture, Transform: localTransform, Color: part.Color}
	return obj, localTransform
}

// ganonWipeRockPartIndex/HalfSize は、崩落演出のクライマックスで画面を
// 覆う「wipe岩」に使うパーツと見た目の大きさ(高さ=HalfSize*2)。他の
// 岩より大きめの塊(Big_2、3550頂点)を使い、カメラのすぐ正面(1.5手前)を
// 落下させることで、近距離ゆえに画面全体を覆って見えるようにする
// (実際に画面いっぱいに覆えているかはcmd/ganon-hall, cmd/ganon-battle側で
// 目視確認・調整する)。
const (
	ganonWipeRockPartIndex = 5
	ganonWipeRockHalfSize  = 1.5
)

// GanonWipeRockObject は、玉座の間→戦場跡のページ遷移演出で使う「wipe岩」
// (画面を覆いながら通り過ぎることで、遷移の継ぎ目を岩の動きに紛れさせる
// ための岩)のObjectとlocalTransformを組み立てる。X/Zは呼び出し側が毎フレーム
// Translate(x, y, z).Mul(localTransform)で指定する(GanonHallCollapseRockObjectと
// 同じ構成)。玉座の間側(降ってきて画面を覆う)・戦場跡側(画面を覆った
// 状態から通り過ぎて場面を見せる)の両方で同じパーツ・同じ大きさを使う。
func GanonWipeRockObject(parts []*Model) (obj Object, localTransform vecmath.Mat4) {
	part := parts[ganonWipeRockPartIndex]
	localTransform = part.GroundTransform(0, 0, ganonWipeRockHalfSize*2)
	obj = Object{Mesh: part.Mesh, Texture: part.Texture, Transform: localTransform, Color: part.Color}
	return obj, localTransform
}

// ganonHallSmokeColor は、Ganon撃破演出の煙玉の色(暗めのグレー)。
var ganonHallSmokeColor = vecmath.NewVec3(0.25, 0.25, 0.28)

// GanonHallSmokePuffPlacement は、Ganon撃破演出で立ち上る煙玉1つぶんの
// 開始位置・開始/終了時の半径・立ち上る高さ。呼び出し側(cmd/ganon-hall/
// main.go)は、経過時間の割合t(0→1)から
//
//	size := StartHalfSize + (EndHalfSize-StartHalfSize)*t
//	y    := Y + RiseHeight*t
//
// を計算し、Translate(X, y, Z).Mul(Scale(size, size, 1))でTransformを
// 更新する(GanonHallSmokeObjectが返す単位quadに対して適用する)。
type GanonHallSmokePuffPlacement struct {
	X, Y, Z                    float64
	StartHalfSize, EndHalfSize float64
	RiseHeight                 float64
}

// GanonHallSmokePuffPlacements は、Ganonを倒した際に立ち上る煙玉の配置
// 一覧。ganonBossHallZ付近、Ganonの胴〜頭の高さ(ganonBossHallY〜+Height)
// を包むように、3つを少しずつ左右・前後にずらして配置している。
var GanonHallSmokePuffPlacements = []GanonHallSmokePuffPlacement{
	{X: -0.6, Y: ganonBossHallY + ganonBossHallHeight*0.4, Z: ganonBossHallZ + 0.3, StartHalfSize: 0.4, EndHalfSize: 1.3, RiseHeight: 1.6},
	{X: 0.5, Y: ganonBossHallY + ganonBossHallHeight*0.6, Z: ganonBossHallZ - 0.2, StartHalfSize: 0.3, EndHalfSize: 1.1, RiseHeight: 2.0},
	{X: 0.0, Y: ganonBossHallY + ganonBossHallHeight*0.8, Z: ganonBossHallZ + 0.6, StartHalfSize: 0.35, EndHalfSize: 1.2, RiseHeight: 1.8},
}

// GanonHallSmokeObject は、煙玉用の単位サイズ(半径1)のquadのObjectを
// 組み立てる(テクスチャはinternal/assets.CloudTexture、色は暗めのグレー)。
// 呼び出し側はGanonHallSmokePuffPlacementsの数だけこのObjectをコピーし、
// 別々のTransformを与えて使う(Mesh/Textureは使い回して問題ない)。
func GanonHallSmokeObject(c *Context) (Object, error) {
	texture, err := c.NewImageTexture(assets.CloudTexture, "image/png")
	if err != nil {
		return Object{}, err
	}
	verts := centeredQuadVertices(1, 1)
	mesh := c.NewMesh(verts, quadUVsCropped(0, 0, 1, 1), quadIndices())
	return Object{Mesh: mesh, Texture: texture, Color: ganonHallSmokeColor}, nil
}
