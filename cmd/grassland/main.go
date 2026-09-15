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

				// 馬の歌が演奏されたら、Linkと同じ場所に馬を呼び出し、以後
				// Linkが馬に乗って移動しているように見せる。すでに呼んで
				// いる場合は何もしない(馬を増やさない)。
				game.SetHorseSummoner(func() {
					if horseIndex >= 0 {
						return
					}
					horse, localTransform, err := renderer.BuildHorseObject(ctx)
					if err != nil {
						fmt.Println("renderer: failed to build horse object:", err)
						return
					}
					horseLocalTransform = localTransform
					horse.Transform = player.Player.Transform(horseLocalTransform)
					scene.Objects = append(scene.Objects, horse)
					horseIndex = len(scene.Objects) - 1
				})

				mountOffset := vecmath.Translate(vecmath.NewVec3(0, renderer.LinkMountHeight, 0))

				ctx.RunLoop(func(dt float64) {
					deltaZ := player.Player.Update(dt)
					if deltaZ != 0 {
						linkTransform := player.Player.Transform(link.LocalTransform)
						if horseIndex >= 0 {
							// 馬に乗っている間は、Linkを持ち上げて背に
							// 座っているように見せる。位置・向きは馬と共通
							// (player.Player)なので、進行方向にあわせて
							// 一緒に向きが変わる。
							linkTransform = mountOffset.Mul(linkTransform)
							scene.Objects[horseIndex].Transform = player.Player.Transform(horseLocalTransform)
						}
						scene.Objects[link.Index].Transform = linkTransform
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
