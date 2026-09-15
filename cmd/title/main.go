package main

import (
	"fmt"
	"runtime"

	"github.com/Kan-O435/okarina/internal/bridge"
	"github.com/Kan-O435/okarina/internal/game"
	"github.com/Kan-O435/okarina/internal/renderer"
	"github.com/Kan-O435/okarina/internal/vecmath"
)

// titleSpinSpeed は、背景のオカリナが1秒間に回転する角度(ラジアン)。
// 2πよりやや小さくすることで、約7秒で1周する穏やかな速さにしている。
const titleSpinSpeed = 0.9

// タイトル画面用のエントリーポイント。ロゴ(HTML側のimg要素、web/index.html
// 参照)がcanvasの手前に重なるため、この3DシーンはHUDのような背景装飾
// (くるくる回るオカリナ)として使う。見た目はまだ仮の最小構成で、MIDI
// キーボードで「ド(C、オクターブ不問)」を弾くと神殿フィールド
// (temple.html)へ遷移する導線を用意する。
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

			scene, localTransform, err := renderer.BuildTitleScene(ctx)
			if err != nil {
				fmt.Println("renderer: failed to build title scene:", err)
			} else {
				angle := 0.0
				ctx.RunLoop(func(dt float64) {
					angle += titleSpinSpeed * dt
					transform := vecmath.RotateY(angle).Mul(localTransform)
					for i := range scene.Objects {
						scene.Objects[i].Transform = transform
					}
					scene.Render(ctx)
				})
				fmt.Println("title: ocarina spinning, waiting for C to start")
			}

			game.SetTitleStartTrigger(func() {
				ctx.Navigate("temple.html")
			})
		}

		select {}
	}
}
