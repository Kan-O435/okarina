package main

import (
	"fmt"
	"runtime"

	"github.com/Kan-O435/okarina/internal/bridge"
	"github.com/Kan-O435/okarina/internal/game"
	"github.com/Kan-O435/okarina/internal/player"
	"github.com/Kan-O435/okarina/internal/renderer"
	"github.com/Kan-O435/okarina/internal/vecmath"
)

// 草原フィールド(扉を抜けた先)用のエントリーポイント。神殿フィールド
// (cmd/game)と同じplayerパッケージ・ブリッジ関数を再利用し、オタマトーンの
// ピッチ入力/矢印キーによるLinkの移動・向き変更をそのまま使えるようにする。
// 隠し扉のような草原フィールド固有のゲームロジックはまだ無い。
func main() {
	fmt.Println("Grassland field initialized")

	if runtime.GOOS == "js" {
		bridge.Init()

		ctx, err := renderer.NewContext("game-canvas")
		if err != nil {
			fmt.Println("renderer: failed to initialize:", err)
		} else {
			width, height := ctx.CanvasSize()
			ctx.Viewport(width, height)
			ctx.EnableDepthTest()
			ctx.EnableBlend()                    // 城の後ろの後光(透過テクスチャ)を正しく合成するため
			ctx.ClearColor(0.6, 0.48, 0.65, 1.0) // 薄紫の不穏な空

			scene, link, err := renderer.BuildGrasslandScene(ctx)
			if err != nil {
				fmt.Println("renderer: failed to build grassland scene:", err)
			} else {
				player.Player.Z = link.SpawnZ
				transitioned := false
				horseIndex := -1
				var horseLocalTransform vecmath.Mat4

				// 馬の歌が演奏されたら、Linkの隣に馬を呼び出す。まだ呼んで
				// いなければSceneにObjectを追加し、すでにいる場合はLinkの
				// 現在位置まで連れてくる(同じ馬を動かすだけで、何頭も
				// 増やさない)。
				game.SetHorseSummoner(func() {
					if horseIndex < 0 {
						horse, localTransform, err := renderer.BuildHorseObject(ctx, player.Player.Z)
						if err != nil {
							fmt.Println("renderer: failed to build horse object:", err)
							return
						}
						scene.Objects = append(scene.Objects, horse)
						horseIndex = len(scene.Objects) - 1
						horseLocalTransform = localTransform
					} else {
						scene.Objects[horseIndex].Transform = renderer.HorseTransform(player.Player.Z).Mul(horseLocalTransform)
					}
				})

				ctx.RunLoop(func(dt float64) {
					deltaZ := player.Player.Update(dt)
					if deltaZ != 0 {
						scene.Objects[link.Index].Transform = player.Player.Transform(link.LocalTransform)
					}
					scene.Render(ctx)

					// Linkが右奥の木のあたりまで進んだら、次のフィールド
					// (ガノン)へページ遷移する。
					if !transitioned && player.Player.Z <= renderer.GrasslandTreeTriggerZ {
						transitioned = true
						ctx.Navigate("ganon.html")
					}
				})
				fmt.Println("renderer: grassland scene rendered, game loop started")
			}
		}

		select {}
	}
}
