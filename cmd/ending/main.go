package main

import (
	"fmt"
	"math"
	"runtime"

	"github.com/Kan-O435/okarina/internal/renderer"
	"github.com/Kan-O435/okarina/internal/vecmath"
)

// endingSpinSpeed は、Link・ゼルダ姫の「塊」が1秒間に回転する角度
// (ラジアン)。タイトル画面のオカリナ(titleSpinSpeed)よりゆっくり、
// 静かに回るくらいの速さにしている。
const endingSpinSpeed = 0.35

// エンディング画面用のエントリーポイント。ガノン最終形態の撃破演出
// (cmd/ganon-battle)から遷移してくる。花畑の上でLinkとゼルダ姫が
// 向かい合って立ち、その2人をまとめて1つの塊とみなしてゆっくり回転させる
// (renderer.BuildEndingScene参照)。MIDI/マイク入力等のゲーム操作は
// 無いページのため、bridge.Init()は呼ばない。
func main() {
	fmt.Println("Ending screen initialized")

	if runtime.GOOS == "js" {
		ctx, err := renderer.NewContext("game-canvas")
		if err != nil {
			fmt.Println("renderer: failed to initialize:", err)
		} else {
			width, height := ctx.CanvasSize()
			ctx.Viewport(width, height)
			ctx.EnableDepthTest()
			ctx.EnableBlend()                     // 花びら(透過テクスチャ)を正しく合成するため
			ctx.ClearColor(0.55, 0.75, 0.95, 1.0) // 空のような水色

			ending, err := renderer.BuildEndingScene(ctx)
			if err != nil {
				fmt.Println("renderer: failed to build ending scene:", err)
			} else {
				angle := 0.0

				// 花びら(花吹雪)の実行時の状態(現在のY座標・累積回転角)。
				// 静的なパラメータ(揺れ方・落下速度等)はending.Petals側に
				// 持たせてあるので、ここでは動く値だけを追いかける。
				petalY := make([]float64, len(ending.Petals))
				petalAngle := make([]float64, len(ending.Petals))
				for i, p := range ending.Petals {
					petalY[i] = p.StartY
				}
				elapsed := 0.0

				ctx.RunLoop(func(dt float64) {
					angle += endingSpinSpeed * dt
					rotate := vecmath.RotateY(angle)
					for i, local := range ending.PairLocalTransforms {
						ending.Scene.Objects[ending.PairObjectStart+i].Transform = rotate.Mul(local)
					}

					elapsed += dt
					for i, p := range ending.Petals {
						petalY[i] -= p.FallSpeed * dt
						if petalY[i] < 0 {
							// 地面まで落ちたら、また上空から降らせ直す
							// (途切れず舞い続ける花吹雪にするため)。
							petalY[i] = renderer.EndingPetalMaxHeight
						}
						petalAngle[i] += p.SpinSpeed * dt

						x := p.X + math.Sin(elapsed*p.SwayFrequency+p.SwayPhase)*p.SwayAmplitude
						transform := vecmath.Translate(vecmath.NewVec3(x, petalY[i], p.Z)).Mul(vecmath.RotateZ(petalAngle[i]))
						ending.Scene.Objects[p.ObjectIndex].Transform = transform
					}

					ending.Scene.Render(ctx)
				})
				fmt.Println("renderer: ending scene rendered, game loop started")
			}
		}

		select {}
	}
}
