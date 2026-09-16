package renderer

import (
	"github.com/Kan-O435/okarina/internal/assets"
	"github.com/Kan-O435/okarina/internal/gltf"
	"github.com/Kan-O435/okarina/internal/vecmath"
)

// このファイルは、ストーリー・操作方法説明画面(cmd/story)の背景に表示する
// SceneCを組み立てる。テキストボックス(web/story.htmlの#story-text)の
// 背後に不穏な存在感が映るよう、玉座の間案(ganon.goのganonBossObject、
// GanonBackgroundHall)と同じガノン(assets.GanonBoss、人型のガノンドロフ
// 姿)を使う。地面・玉座の間の建物・Linkは配置せず、暗い背景にガノン単体を
// 浮かべた固定シーンにする(アニメーションは無い静止画として使う想定)。

// storyGanonHeight/storyGanonCameraDistance は、ストーリー画面での
// ガノンの見た目の高さ・カメラからの距離。ganon-hall本編でのガノン
// (ganonBossHallHeight=2.2、他の建物・ボスとのバランスを取った値)より
// 大きめにして、テキストボックスの上に頭〜上半身が大きく覗く存在感を
// 出している。distanceは、45°の垂直画角でこの高さがフレームの7割
// 程度を占めるよう逆算した値。
const (
	storyGanonHeight         = 5.5
	storyGanonCameraDistance = 9.0
)

// storyGanonYawDegrees は、ストーリー画面のガノンを既定の向き(+Z、カメラ側)
// からさらに左へ回転させる角度。ganon.goのganonBossHallYawDegrees(-25°、
// 玉座の間本編用)とは別に、このページ専用の値として独立させている
// (本編側の見た目に影響を与えないため)。-25°では中途半端な傾きに見えた
// ため、より左を向かせて-40°にしている。
const storyGanonYawDegrees = -40.0

// storyGanonX は、ガノンのワールドX座標。カメラはX=0を中心に見ているため
// (BuildStoryScene参照)、負の値にするほど画面上で左寄りに配置される。
// テキストボックスの左右中央よりガノンが少し左に見えるよう調整している。
const storyGanonX = -1.1

// BuildStoryScene は、ストーリー画面のSceneを組み立てる。カメラは
// storyGanonHeightの6割の高さ(胸のあたり)を水平に見る固定視点。
func BuildStoryScene(c *Context) (*Scene, error) {
	program, err := c.NewProgram(basicVertexShaderSrc, basicFragmentShaderSrc)
	if err != nil {
		return nil, err
	}

	width, height := c.CanvasSize()
	aspect := float64(width) / float64(height)
	projection := vecmath.Perspective(vecmath.Radians(45), aspect, 0.1, 100)
	eyeY := storyGanonHeight * 0.6
	view := vecmath.LookAt(
		vecmath.NewVec3(0, eyeY, storyGanonCameraDistance),
		vecmath.NewVec3(0, eyeY, 0),
		vecmath.NewVec3(0, 1, 0),
	)

	// gltf.ParseParts()がノードのワールド変換行列を焼き込むため、
	// ganon.goのganonBossObject(玉座の間案)と同様にそのままCombined
	// GroundTransformで組み上げられる。バインドポーズがわずかに傾いて
	// 見えるため、storyGanonYawDegreesの回転を掛けて見た目を整えている。
	prims, err := gltf.ParseParts(assets.GanonBoss)
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

	localTransform := CombinedGroundTransform(parts, 0, 0, storyGanonHeight)
	facing := vecmath.RotateY(vecmath.Radians(storyGanonYawDegrees))
	transform := vecmath.Translate(vecmath.NewVec3(storyGanonX, 0, 0)).Mul(facing).Mul(localTransform)

	objects := make([]Object, len(parts))
	for i, part := range parts {
		objects[i] = Object{Mesh: part.Mesh, Texture: part.Texture, Transform: transform, Color: part.Color}
	}

	return &Scene{
		Program:        program,
		ViewProjection: projection.Mul(view),
		Objects:        objects,
	}, nil
}
