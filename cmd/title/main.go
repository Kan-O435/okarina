package main

import (
	"fmt"
	"math/rand"
	"runtime"

	"github.com/Kan-O435/okarina/internal/bridge"
	"github.com/Kan-O435/okarina/internal/game"
	"github.com/Kan-O435/okarina/internal/renderer"
	"github.com/Kan-O435/okarina/internal/vecmath"
)

// titleSpinSpeed は、手前のメインのオカリナが1秒間に回転する角度
// (ラジアン)。2πよりやや小さくすることで、約7秒で1周する穏やかな
// 速さにしている。
const titleSpinSpeed = 0.9

// 背景で漂う小さなオカリナ群のパラメータ。メインのオカリナ(手前、原点で
// 自転するだけ)とは違い、それぞれ別々の位置・速度・向きでランダムに
// 飛び回らせることで賑やかな背景にする。
const (
	numBackgroundOcarinas = 24

	bgOcarinaMinHeight = 1.0
	bgOcarinaMaxHeight = 2.4

	bgBoundMinX, bgBoundMaxX = -12.0, 12.0
	bgBoundMinY, bgBoundMaxY = -2.0, 10.0
	bgBoundMinZ, bgBoundMaxZ = -35.0, -8.0

	bgSpeedMin = 0.4 // ワールド単位/秒(各軸成分の最小の大きさ)
	bgSpeedMax = 1.6
	bgSpinMin  = 0.5 // ラジアン/秒(各軸周りの回転速度の最小の大きさ)
	bgSpinMax  = 2.5
)

// backgroundOcarina は、背景で飛び回る小さなオカリナ1体分の状態。
type backgroundOcarina struct {
	objectStart int // scene.Objects内で、このオカリナのパーツが始まるインデックス
	local       vecmath.Mat4
	pos         vecmath.Vec3
	vel         vecmath.Vec3
	angle       vecmath.Vec3 // X/Y/Z軸それぞれの現在の回転角
	spin        vecmath.Vec3 // X/Y/Z軸それぞれの回転速度(ラジアン/秒)
}

// randRange は[min, max)の一様乱数を返す。
func randRange(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// randSigned は、絶対値が[min, max)の範囲でランダムな正または負の値を返す。
// 0付近で止まって見えないよう、速度・回転速度の生成に使う。
func randSigned(min, max float64) float64 {
	v := randRange(min, max)
	if rand.Intn(2) == 0 {
		return -v
	}
	return v
}

// newBackgroundOcarina は、objectStartの位置にパーツが並んでいる背景の
// 小さなオカリナ1体分の状態を、ランダムな位置・速度・回転速度で初期化
// する。localは、そのパーツをBuildTitleBackgroundOcarinaObjectsで組み
// 立てた際に使ったのと同じ変換(大きさ)を渡す。
func newBackgroundOcarina(objectStart int, local vecmath.Mat4) backgroundOcarina {
	return backgroundOcarina{
		objectStart: objectStart,
		local:       local,
		pos: vecmath.NewVec3(
			randRange(bgBoundMinX, bgBoundMaxX),
			randRange(bgBoundMinY, bgBoundMaxY),
			randRange(bgBoundMinZ, bgBoundMaxZ),
		),
		vel: vecmath.NewVec3(
			randSigned(bgSpeedMin, bgSpeedMax),
			randSigned(bgSpeedMin, bgSpeedMax),
			randSigned(bgSpeedMin, bgSpeedMax),
		),
		spin: vecmath.NewVec3(
			randSigned(bgSpinMin, bgSpinMax),
			randSigned(bgSpinMin, bgSpinMax),
			randSigned(bgSpinMin, bgSpinMax),
		),
	}
}

// bounce は、vをdt秒分だけposに加え、[min, max]の範囲を超えたら跳ね返す
// (速度の符号を反転し、範囲の内側に収める)ように更新する。
func bounce(pos, vel *float64, dt, min, max float64) {
	*pos += *vel * dt
	if *pos < min {
		*pos = min
		*vel = -*vel
	} else if *pos > max {
		*pos = max
		*vel = -*vel
	}
}

// update は、dt秒分だけ位置・回転角を進め(範囲外に出たら跳ね返り)、
// scene.Objects内の対応するパーツのTransformを書き換える。
func (b *backgroundOcarina) update(dt float64, scene *renderer.Scene, partCount int) {
	bounce(&b.pos.X, &b.vel.X, dt, bgBoundMinX, bgBoundMaxX)
	bounce(&b.pos.Y, &b.vel.Y, dt, bgBoundMinY, bgBoundMaxY)
	bounce(&b.pos.Z, &b.vel.Z, dt, bgBoundMinZ, bgBoundMaxZ)

	b.angle.X += b.spin.X * dt
	b.angle.Y += b.spin.Y * dt
	b.angle.Z += b.spin.Z * dt

	transform := vecmath.Translate(b.pos).
		Mul(vecmath.RotateX(b.angle.X)).
		Mul(vecmath.RotateY(b.angle.Y)).
		Mul(vecmath.RotateZ(b.angle.Z)).
		Mul(b.local)

	for i := 0; i < partCount; i++ {
		scene.Objects[b.objectStart+i].Transform = transform
	}
}

// タイトル画面用のエントリーポイント。ロゴ(HTML側のimg要素、web/index.html
// 参照)がcanvasの手前に重なるため、この3DシーンはHUDのような背景装飾
// として使う。手前でメインのオカリナがくるくる回り、奥では同じモデルの
// 小さなオカリナが大量に飛び回る。MIDIキーボードで「ド(C、オクターブ
// 不問)」を弾くと、ストーリー・操作方法の説明画面(story.html)へ
// 遷移する導線を用意する。
func main() {
	fmt.Println("Title screen initialized")

	if runtime.GOOS == "js" {
		bridge.Init()

		ctx, err := renderer.NewContext("game-canvas")
		if err != nil {
			fmt.Println("renderer: failed to initialize:", err)
		} else {
			width, height := ctx.CanvasSize()
			ctx.Viewport(width, height)
			ctx.EnableDepthTest()
			ctx.ClearColor(0.05, 0.05, 0.1, 1.0) // 仮の暗い背景

			scene, localTransform, ocarinaParts, err := renderer.BuildTitleScene(ctx)
			if err != nil {
				fmt.Println("renderer: failed to build title scene:", err)
			} else {
				// BuildTitleSceneが返した時点でのscene.Objectsは、メインの
				// オカリナのパーツのみ。以後に追加する背景のオカリナ群と
				// 区別するため、ここで個数を控えておく。
				mainOcarinaObjectCount := len(scene.Objects)
				partCount := len(ocarinaParts)

				backgrounds := make([]backgroundOcarina, 0, numBackgroundOcarinas)
				for i := 0; i < numBackgroundOcarinas; i++ {
					h := randRange(bgOcarinaMinHeight, bgOcarinaMaxHeight)
					objects, local := renderer.BuildTitleBackgroundOcarinaObjects(ocarinaParts, h)
					objectStart := len(scene.Objects)
					scene.Objects = append(scene.Objects, objects...)
					backgrounds = append(backgrounds, newBackgroundOcarina(objectStart, local))
				}

				angle := 0.0
				ctx.RunLoop(func(dt float64) {
					angle += titleSpinSpeed * dt
					transform := vecmath.RotateY(angle).Mul(localTransform)
					for i := 0; i < mainOcarinaObjectCount; i++ {
						scene.Objects[i].Transform = transform
					}

					for i := range backgrounds {
						backgrounds[i].update(dt, scene, partCount)
					}

					scene.Render(ctx)
				})
				fmt.Println("title: ocarina spinning, waiting for C to start")
			}

			game.SetTitleStartTrigger(func() {
				ctx.Navigate("story.html")
			})
		}

		select {}
	}
}
