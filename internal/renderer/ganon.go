package renderer

import (
	"github.com/Kan-O435/okarina/internal/assets"
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

// ganonLinkZ は、Linkをカメラのすぐ手前に立たせておく位置。
const ganonLinkZ = -20.0

// GanonBackgroundVariant は、ガノンフィールドの背景として比較中の
// 2つのデザイン案を切り替えるための識別子。
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

	linkObj, linkLocal, err := ganonLinkObject(c)
	if err != nil {
		return nil, LinkPlacement{}, err
	}
	objects = append(objects, linkObj)
	link = LinkPlacement{
		Index:          len(objects) - 1,
		LocalTransform: linkLocal,
		SpawnZ:         ganonLinkZ,
	}

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
const (
	ganonHallCameraY = 5.0
	ganonHallCameraZ = -28.8
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
			vecmath.NewVec3(0, ganonHallCameraY, ganonBackgroundZ),
			vecmath.NewVec3(0, 1, 0),
		)
	}

	// 戦場跡案: 幅に対して高さが低い(横長)ため、縦方向を画面いっぱいに
	// しようとすると横方向は画面からはみ出すほど寄る必要がある。カメラの
	// 高さを背景の縦方向の中心(Y=5、高さ10の半分)に合わせたうえで寄せ、
	// 上下方向の隙間が出ないようにしている。
	return vecmath.LookAt(
		vecmath.NewVec3(0, 5, -17),
		vecmath.NewVec3(0, 5, -35),
		vecmath.NewVec3(0, 1, 0),
	)
}

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

// ganonLinkObject は、神殿・草原フィールドと同じKayKit Knightモデル
// (internal/assets.LinkKnight)を配置する。他のフィールドのlinkObjectと
// 同様、localTransformを別に返し、プレイヤーの位置・向きと組み合わせて
// 毎フレームTransformを再構築できるようにする。
func ganonLinkObject(c *Context) (obj Object, localTransform vecmath.Mat4, err error) {
	model, err := c.LoadSkinnedGLBMesh(assets.LinkKnight)
	if err != nil {
		return Object{}, vecmath.Mat4{}, err
	}

	localTransform = model.GroundTransform(0, 0, linkTargetHeight)
	transform := vecmath.Translate(vecmath.NewVec3(0, 0, ganonLinkZ)).Mul(localTransform)
	obj = Object{Mesh: model.Mesh, Texture: model.Texture, Transform: transform, Color: model.Color}
	return obj, localTransform, nil
}
