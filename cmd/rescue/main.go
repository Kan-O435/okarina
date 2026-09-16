package main

import (
	"fmt"
	"runtime"

	"github.com/Kan-O435/okarina/internal/bridge"
	"github.com/Kan-O435/okarina/internal/player"
	"github.com/Kan-O435/okarina/internal/renderer"
)

// 「ゼルダ姫を迎えに行く」フィールド用のエントリーポイント。戦場跡
// (cmd/ganon-battle)でGanon最終形態を倒した後、エンディング(cmd/ending)
// へ直接遷移するのではなくこのページを挟む。神殿・草原フィールドと同じ
// player/bridgeを再利用し、オタマトーンのピッチ入力/矢印キーでLinkを
// 前後移動できるようにする(renderer.BuildRescueScene参照)。Linkが
// ゼルダ姫にRescueReachRangeZまで近づいたら、エンディングへ遷移する。
func main() {
	fmt.Println("Rescue field initialized")

	if runtime.GOOS == "js" {
		bridge.Init()

		ctx, err := renderer.NewContext("game-canvas")
		if err != nil {
			fmt.Println("renderer: failed to initialize:", err)
		} else {
			width, height := ctx.CanvasSize()
			ctx.Viewport(width, height)
			ctx.EnableDepthTest()
			ctx.EnableBlend()                     // 雲(エンディングと共通)を正しく合成するため
			ctx.ClearColor(0.55, 0.75, 0.95, 1.0) // 空のような水色(エンディングと同じ)

			scene, link, err := renderer.BuildRescueScene(ctx)
			if err != nil {
				fmt.Println("renderer: failed to build rescue scene:", err)
			} else {
				player.Player.SpawnAt(link.SpawnZ)
				renderer.SetLinkTransform(scene, link, player.Player.Transform(link.LocalTransform))

				reachZ := renderer.RescueZeldaZ + renderer.RescueReachRangeZ
				navigated := false

				ctx.RunLoop(func(dt float64) {
					player.Player.Update(dt)
					renderer.SetLinkTransform(scene, link, player.Player.Transform(link.LocalTransform))

					if !navigated && player.Player.Z <= reachZ {
						navigated = true
						ctx.Navigate("ending.html")
					}

					scene.Render(ctx)
				})
				fmt.Println("renderer: rescue scene rendered, game loop started")
			}
		}

		select {}
	}
}
