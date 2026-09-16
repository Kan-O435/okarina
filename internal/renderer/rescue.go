package renderer

import (
	"math"

	"github.com/Kan-O435/okarina/internal/assets"
	"github.com/Kan-O435/okarina/internal/vecmath"
)

// このファイルは「ゼルダ姫を迎えに行く」フィールドのSceneを組み立てる。
// 戦場跡でGanon最終形態を倒した後、エンディング(向かい合って回転する
// 演出)へ直接飛ぶのではなく、神殿・草原フィールドと同じくオタマトーンの
// ピッチ入力でLinkを前後移動させ、少し離れた場所に立つゼルダ姫まで
// 歩いて近づく短い区間を挟む(cmd/rescue/main.go参照)。草原ステージ以降
// オタマトーンでの移動操作を使う場面が無くなっていたため、エンディング
// 直前でもう一度この操作を使わせる狙い。
//
// 背景はエンディング画面(internal/renderer/ending.go)と同じ花畑+空にして
// 視覚的な繋がりを持たせている(地面のテクスチャ・雲は共通のものを流用)。

const (
	// rescueGroundHalfExtent は、花畑の地面(1枚の板)の半径。
	rescueGroundHalfExtent = 30.0

	// rescueLinkSpawnZ/rescueZeldaZは、Linkの初期スポーン位置とゼルダ姫の
	// 立ち位置(Z座標)。神殿・草原と同じく、Zが減る方向が「前進」。
	rescueLinkSpawnZ = 10.0
	rescueZeldaZ     = -12.0

	// rescueCharacterHeight は、Link・ゼルダ姫をこの高さになるようスケール
	// する(神殿・草原・エンディングと同じ)。
	rescueCharacterHeight = 1.4
)

// RescueZeldaZ/RescueReachRangeZ は、ゲームループ(internal/game)が
// 「プレイヤーがゼルダ姫の近くにいるかどうか」を判定する際に参照するため、
// パッケージ外から参照できるようにエクスポートする。
const (
	RescueZeldaZ      = rescueZeldaZ
	RescueReachRangeZ = 2.5
)

// rescueCameraEye/Target/Upは、Linkのすぐ後ろ上空から、ゼルダ姫の方向を
// 見下ろす固定カメラ(草原フィールドの馬に乗る前のカメラと同じ考え方)。
var (
	rescueCameraEye    = vecmath.NewVec3(0, 4.5, rescueLinkSpawnZ+5)
	rescueCameraTarget = vecmath.NewVec3(0, 1.0, rescueZeldaZ)
	rescueCameraUp     = vecmath.NewVec3(0, 1, 0)
)

// rescueProjection は、このフィールドの透視投影行列を組み立てる。
func rescueProjection(aspect float64) vecmath.Mat4 {
	return vecmath.Perspective(vecmath.Radians(55), aspect, 0.1, 150)
}

// BuildRescueScene は、花畑の地面+ゼルダ姫+LinkのSceneを組み立てる。
func BuildRescueScene(c *Context) (scene *Scene, link LinkPlacement, err error) {
	program, err := c.NewProgram(basicVertexShaderSrc, basicFragmentShaderSrc)
	if err != nil {
		return nil, LinkPlacement{}, err
	}

	width, height := c.CanvasSize()
	aspect := float64(width) / float64(height)
	projection := rescueProjection(aspect)
	view := vecmath.LookAt(rescueCameraEye, rescueCameraTarget, rescueCameraUp)

	groundTexture, err := c.NewImageTexture(assets.FlowerFieldTexture, "image/png")
	if err != nil {
		return nil, LinkPlacement{}, err
	}
	groundMesh := c.NewMesh(
		quadVertices(rescueGroundHalfExtent, rescueGroundHalfExtent, 0),
		quadUVsCropped(0, 0, 1, 1),
		quadIndices(),
	)
	objects := []Object{{
		Mesh:      groundMesh,
		Texture:   groundTexture,
		Transform: vecmath.Identity(),
		Color:     vecmath.NewVec3(1, 1, 1),
	}}

	// エンディング画面と同じ雲(internal/renderer/ending.go参照)を空に
	// 浮かべ、視覚的な繋がりを持たせる。
	clouds, err := endingCloudObjects(c)
	if err != nil {
		return nil, LinkPlacement{}, err
	}
	objects = append(objects, clouds...)

	zeldaParts, err := c.LoadGLBParts(assets.Zelda)
	if err != nil {
		return nil, LinkPlacement{}, err
	}
	zeldaCentered := CombinedGroundTransform(zeldaParts, 0, 0, rescueCharacterHeight)
	// ゼルダ姫はLinkの歩いてくる方向(+Z側)を向いて待つ。実機確認の上で
	// 符号を決めている(internal/renderer/ending.goと同じモデル・読み込み
	// 方法のため、そちらで「理論値通りだと背中合わせになった」経験を
	// 踏まえてRotateY(π)を採用。見た目がおかしい場合は要調整)。
	zeldaLocal := vecmath.Translate(vecmath.NewVec3(0, 0, rescueZeldaZ)).
		Mul(vecmath.RotateY(math.Pi)).
		Mul(zeldaCentered)
	for _, p := range zeldaParts {
		objects = append(objects, Object{Mesh: p.Mesh, Texture: p.Texture, Transform: zeldaLocal, Color: p.Color})
	}

	linkParts, err := loadLinkParts(c)
	if err != nil {
		return nil, LinkPlacement{}, err
	}
	linkLocal := CombinedGroundTransform(linkParts, 0, 0, rescueCharacterHeight)
	linkTransform := vecmath.Translate(vecmath.NewVec3(0, 0, rescueLinkSpawnZ)).Mul(linkLocal)
	linkObjs := make([]Object, len(linkParts))
	for i, part := range linkParts {
		linkObjs[i] = Object{Mesh: part.Mesh, Texture: part.Texture, Transform: linkTransform, Color: part.Color}
	}
	objects, link = appendLinkObjects(objects, linkObjs, linkLocal, rescueLinkSpawnZ)

	return &Scene{
		Program:        program,
		ViewProjection: projection.Mul(view),
		Objects:        objects,
	}, link, nil
}
