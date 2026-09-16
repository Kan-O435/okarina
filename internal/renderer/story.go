package renderer

import (
	"bytes"
	"image"
	"image/draw"
	"image/gif"

	"github.com/Kan-O435/okarina/internal/assets"
	"github.com/Kan-O435/okarina/internal/gltf"
	"github.com/Kan-O435/okarina/internal/vecmath"
)

// このファイルは、ストーリー・操作方法説明画面(cmd/story)の背景に表示する
// SceneCを組み立てる。テキストボックス(web/story.htmlの#story-text)の
// 背後に不穏な存在感が映るよう、玉座の間案(ganon.goのganonBossObject、
// GanonBackgroundHall)と同じガノン(assets.GanonBoss、人型のガノンドロフ
// 姿)を使う。地面・玉座の間の建物・Linkは配置せず、暗い背景にガノン単体を
// 浮かべた固定シーンにする(ガノン自体はアニメーション無しの静止ポーズ。
// 背後の炎だけがちらつくアニメーションを持つ)。

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

// --- ここから、ガノンの背後で燃える炎(ビルボード)の実装。
//
// 最初はGLB(3Dメッシュ、静止画)で実装していたが、「炎が動いて(ちらついて)
// 見えるようにしてほしい」という要望を受けて、アニメーションするスプライト
// シート(assets.StoryFlameAnimation)を貼った板(ビルボード)に置き換えた。
// GanonHallSmokeObject(ganon.go)と同じ「板+テクスチャ」のアプローチだが、
// あちらは1枚絵を拡大縮小するだけなのに対し、こちらはUV座標を切り替えて
// スプライトシートの中の別フレームを表示することでコマ送りアニメーションを
// 実現している。

// storyFlameFrameCols/Rows/W/H は、assets.StoryFlameAnimation
// (Small_Fireball_10x26.png)のグリッド構成。1フレーム10x26px、
// 横10列×縦6行(6行は色違い等のバリエーションと思われる)。このうち
// 先頭の1行(10フレーム)だけをアニメーションループとして使う。
const (
	storyFlameFrameCols = 10
	storyFlameFrameRows = 6
	storyFlameFrameW    = 10
	storyFlameFrameH    = 26
	storyFlameSheetW    = storyFlameFrameCols * storyFlameFrameW
	storyFlameSheetH    = storyFlameFrameRows * storyFlameFrameH
)

// storyFlameFPS は、炎アニメーションの再生速度(フレーム/秒)。
const storyFlameFPS = 10.0

// storyFlamePlacement は、ガノンの背後に浮かべる炎1個分の配置(ワールドX/Z、
// 底辺のY、底辺からの高さ)。
type storyFlamePlacement struct {
	X, Y, Z, Height float64
}

// storyFlamePlacements は、ガノンドロフ(storyGanonX、Z=0付近)の奥に
// 炎が立ち上っているように見せるための配置。Zをガノンより奥(カメラから
// 遠い、より負の値)にすることで、ガノンのシルエットの背後に炎が回り込んで
// 見える。地面(Y=0)付近はテキストボックス(web/story.htmlの#story-text)に
// 隠れてしまうため、Yをテキストボックスの上端より高い位置(ガノンの胸
// あたり)まで持ち上げ、そこから頭上にかけて炎が見えるようにしている
// (足元から生やすと、素材のドット絵解像度に対して見た目上必要な高さが
// 大きくなりすぎ、拡大率が上がってぼやけて見えるため)。中央に一番高い炎、
// 左右に少し低い炎を置き、単調な1本立てにならないようにしている。
var storyFlamePlacements = []storyFlamePlacement{
	{X: storyGanonX, Y: 3.6, Z: -2.0, Height: 4.2},
	{X: storyGanonX - 2.2, Y: 3.3, Z: -1.5, Height: 3.6},
	{X: storyGanonX + 2.4, Y: 3.3, Z: -1.5, Height: 3.8},
}

// StoryFlameAnimation は、story画面背景の炎(ビルボード)をアニメーション
// させるための状態。BuildStorySceneが返すsceneのうち、炎に対応する
// Objectだけを毎フレームUpdate(dt)で切り替える。
type StoryFlameAnimation struct {
	scene         *Scene
	frameMeshes   []*Mesh
	objectIndices []int
	elapsed       float64
}

// Update は経過時間dtを進め、現在の再生位置に応じたフレームのMeshを
// 炎のObjectへ反映する(全ての炎で同じフレームを共有し、同期して揺れる)。
func (a *StoryFlameAnimation) Update(dt float64) {
	if a == nil || len(a.frameMeshes) == 0 {
		return
	}
	a.elapsed += dt
	frame := int(a.elapsed*storyFlameFPS) % len(a.frameMeshes)
	mesh := a.frameMeshes[frame]
	for _, idx := range a.objectIndices {
		a.scene.Objects[idx].Mesh = mesh
	}
}

// BuildStoryScene は、ストーリー画面のSceneを組み立てる。カメラは
// storyGanonHeightの6割の高さ(胸のあたり)を水平に見る固定視点。
// 戻り値のStoryFlameAnimationは、呼び出し側(cmd/story)がRunLoop内で
// 毎フレームUpdate(dt)を呼ぶことで、背後の炎がちらつくアニメーションになる。
func BuildStoryScene(c *Context) (*Scene, *StoryFlameAnimation, error) {
	program, err := c.NewProgram(basicVertexShaderSrc, basicFragmentShaderSrc)
	if err != nil {
		return nil, nil, err
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
		return nil, nil, err
	}
	parts := make([]*Model, len(prims))
	for i := range prims {
		model, err := c.buildModel(&prims[i])
		if err != nil {
			return nil, nil, err
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

	// 炎エフェクトはいったん無効化(コメントアウト)。「一旦削除/コメント
	// アウトしてほしい」という要望のため。再度有効化する場合は、この
	// ブロックのコメントを外し、下のreturnを差し替える。
	//
	// flameTexture, err := c.NewPixelArtTexture(assets.StoryFlameAnimation, "image/png")
	// if err != nil {
	// 	return nil, nil, err
	// }
	// flameFrameMeshes := buildStoryFlameFrameMeshes(c)
	//
	// // ガノン(不透明)を先に描画してから炎(半透明ビルボード)を描画する。
	// // 炎はガノンより奥(Zがより負)にあるため、深度テストによりガノンに
	// // 隠れる部分は正しく描画されない(炎同士は画面上でほぼ重ならない
	// // 配置のため、描画順による見た目への影響は無い)。
	// flameIndices := make([]int, len(storyFlamePlacements))
	// for i, p := range storyFlamePlacements {
	// 	flameIndices[i] = len(objects)
	// 	objects = append(objects, Object{
	// 		Mesh:      flameFrameMeshes[0],
	// 		Texture:   flameTexture,
	// 		Transform: vecmath.Translate(vecmath.NewVec3(p.X, p.Y, p.Z)).Mul(vecmath.Scale(vecmath.NewVec3(p.Height, p.Height, p.Height))),
	// 		Color:     vecmath.NewVec3(1, 1, 1),
	// 	})
	// }

	scene := &Scene{
		Program:        program,
		ViewProjection: projection.Mul(view),
		Objects:        objects,
	}

	// return scene, &StoryFlameAnimation{scene: scene, frameMeshes: flameFrameMeshes, objectIndices: flameIndices}, nil
	return scene, nil, nil
}

// buildStoryFlameFrameMeshes は、assets.StoryFlameAnimationの先頭行
// (storyFlameFrameCols枚)を、それぞれ1コマ分のUVを持つ板(縦長の
// verticalQuadVertices、足元中央が原点・高さ1の単位サイズ)として
// 作成する。単位サイズにしているのは、実際の高さ(storyFlamePlacementの
// Height)はObject.Transformのスケールで調整するため、Meshそのものは
// 全配置・全フレームで共有できるようにするため(GPUバッファの使い回し)。
func buildStoryFlameFrameMeshes(c *Context) []*Mesh {
	halfWidth := float32(storyFlameFrameW) / float32(storyFlameFrameH) / 2
	verts := verticalQuadVertices(halfWidth, 1)

	meshes := make([]*Mesh, storyFlameFrameCols)
	for col := 0; col < storyFlameFrameCols; col++ {
		u0 := float32(col*storyFlameFrameW) / float32(storyFlameSheetW)
		u1 := float32((col+1)*storyFlameFrameW) / float32(storyFlameSheetW)
		// 先頭行(row=0、画像の一番上)だけを使う。v0/v1はquadUVsCropped経由で
		// Objectのメッシュ下端/上端に対応する(ganon.goのdoorTextureV0/V1と
		// 同じ考え方)。素材の炎の見た目を上下反転させたいという要望のため、
		// 通常(v0=画像下側、v1=画像上側)とは逆に、v0=画像上側、v1=画像下側を
		// 割り当てている。
		v0 := float32(1)
		v1 := float32(1) - float32(storyFlameFrameH)/float32(storyFlameSheetH)
		meshes[col] = c.NewMesh(verts, quadUVsCropped(u0, v0, u1, v1), quadIndices())
	}
	return meshes
}

// --- ここから、ストーリー画面の背景全体を覆う炎アニメーションの実装。
//
// 上のstoryFlame*(ガノンの脇に立てる小さな炎のビルボード数本)は
// コメントアウトして無効化し、代わりにキャンバス全体を覆う1枚の
// 炎アニメーションGIF(assets.StoryBackgroundFire、ユーザー提供)に
// 置き換えた。WhiteFadeOverlay(overlay.go)と同じ「スクリーン座標に
// 固定したフルスクリーンの板」の仕組みを使うが、白一色ではなく、
// GIFの各フレームを順番に貼り替える点が異なる。
//
// ブラウザのcreateImageBitmapはアニメーションGIFを渡しても最初の
// 1コマしかビットマップ化できないため、NewImageTexture(ブラウザの
// デコーダ経由)は使えない。代わりにGo標準ライブラリのimage/gifで
// GIFバイト列自体を全フレームデコードし、各フレームをNewRGBATexture
// (texture_js.go)でそのままテクスチャ化する。

// storyBackgroundFPS は、背景炎アニメーションの再生速度(フレーム/秒)。
// 元のGIFのフレーム間隔(5/100秒)から算出した値。
const storyBackgroundFPS = 20.0

// StoryBackground は、story画面の背景全体を覆う炎アニメーションの状態。
type StoryBackground struct {
	program *Program
	mesh    *Mesh
	frames  []*Texture
	elapsed float64
}

// BuildStoryBackground は、assets.StoryBackgroundFireの全フレームを
// デコードしてテクスチャ化し、キャンバス全体を覆う板を1つ組み立てる。
// widthPx/heightPxはcanvasの実サイズ(Context.CanvasSize())。
func BuildStoryBackground(c *Context, widthPx, heightPx int) (*StoryBackground, error) {
	program, err := c.NewProgram(basicVertexShaderSrc, basicFragmentShaderSrc)
	if err != nil {
		return nil, err
	}
	// v座標をWhiteFadeOverlayの(0, 1, 1, 0)から(0, 0, 1, 1)へ入れ替えて、
	// 背景画像を上下反転させている(「背景の上下を逆にしてほしい」という要望のため)。
	mesh := c.NewMesh(fullscreenQuadVertices(widthPx, heightPx), quadUVsCropped(0, 0, 1, 1), quadIndices())

	g, err := gif.DecodeAll(bytes.NewReader(assets.StoryBackgroundFire))
	if err != nil {
		return nil, err
	}
	frames := make([]*Texture, len(g.Image))
	for i, frame := range g.Image {
		bounds := frame.Bounds()
		rgba := image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, frame, bounds.Min, draw.Src)
		frames[i] = c.NewRGBATexture(rgba.Pix, bounds.Dx(), bounds.Dy())
	}

	return &StoryBackground{program: program, mesh: mesh, frames: frames}, nil
}

// Update は経過時間dtを進める。
func (b *StoryBackground) Update(dt float64) {
	if b == nil {
		return
	}
	b.elapsed += dt
}

// Render はキャンバス全体を現在のフレームの炎画像で塗りつぶす。深度テストを
// 一時的に無効化して描画するため、呼び出し側(cmd/story)はGanonのScene.Render
// より先にこれを呼ぶこと(3Dシーンの手前に炎が被さらないようにするため)。
func (b *StoryBackground) Render(c *Context, canvasWidth, canvasHeight int) {
	if b == nil || len(b.frames) == 0 {
		return
	}
	frame := int(b.elapsed*storyBackgroundFPS) % len(b.frames)

	c.DisableDepthTest()

	ortho := vecmath.Ortho(0, float64(canvasWidth), float64(canvasHeight), 0, -1, 1)

	b.program.Use(c)
	b.program.SetUniformSampler(c, "uTexture", 0)
	b.frames[frame].Bind(c)
	b.program.SetUniformMat4(c, "uMVP", ortho)
	b.program.SetUniformVec3(c, "uColor", vecmath.NewVec3(1, 1, 1))
	b.program.SetUniformFloat(c, "uAlpha", 1.0)
	b.mesh.Draw(c, b.program)

	c.EnableDepthTest()
}
