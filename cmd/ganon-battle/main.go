package main

import (
	"fmt"
	"runtime"

	"github.com/Kan-O435/okarina/internal/bridge"
	"github.com/Kan-O435/okarina/internal/player"
	"github.com/Kan-O435/okarina/internal/renderer"
)

// ガノンフィールド「戦場跡」案用のエントリーポイント。玉座の間案
// (cmd/ganon-hall)とは別ページ・別Linkに分けており、このページでは
// 戦場跡の背景だけを固定で表示する。
//
// 草原フィールド(cmd/grassland)と同じplayerパッケージ・ブリッジ関数を
// 再利用し、オタマトーンのピッチ入力/矢印キーによるLinkの移動・向き変更を
// そのまま使えるようにする。ボス戦固有のゲームロジックはまだ無い。
func main() {
	fmt.Println("Ganon field (battle) initialized")

	if runtime.GOOS == "js" {
		bridge.Init()

		ctx, err := renderer.NewContext("game-canvas")
		if err != nil {
			fmt.Println("renderer: failed to initialize:", err)
			select {}
		}

		width, height := ctx.CanvasSize()
		ctx.Viewport(width, height)
		ctx.EnableDepthTest()
		ctx.ClearColor(0.15, 0.05, 0.05, 1.0) // 暗く不穏な赤黒い空気

		scene, link, err := renderer.BuildGanonScene(ctx, renderer.GanonBackgroundBattle)
		if err != nil {
			fmt.Println("renderer: failed to build ganon scene:", err)
		} else {
			player.Player.Z = link.SpawnZ

			ctx.RunLoop(func(dt float64) {
				deltaZ := player.Player.Update(dt)
				if deltaZ != 0 {
					scene.Objects[link.Index].Transform = player.Player.Transform(link.LocalTransform)
				}
				scene.Render(ctx)
			})
			fmt.Println("renderer: ganon (battle) scene rendered, game loop started")
		}

		select {}
	}
}
