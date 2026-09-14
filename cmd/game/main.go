package main

import (
	"fmt"
	"runtime"

	"github.com/Kan-O435/okarina/internal/bridge"
	"github.com/Kan-O435/okarina/internal/game"
	"github.com/Kan-O435/okarina/internal/renderer"
)

func main() {
	fmt.Println("Game initialized")
	game.Run()

	// WASM(ブラウザ)で動かす場合、main()がreturnするとプログラムが
	// 終了してしまい、以後 JS 側から Go の関数を呼べなくなる。
	// JS-Go Bridge を有効にするため、js/wasm ビルドの時だけ
	// 関数を登録してからプログラムを常駐させる。
	if runtime.GOOS == "js" {
		bridge.Init()

		ctx, err := renderer.NewContext("game-canvas")
		if err != nil {
			fmt.Println("renderer: failed to initialize:", err)
		} else {
			width, height := ctx.CanvasSize()
			ctx.Viewport(width, height)
			ctx.EnableDepthTest()
			ctx.EnableBlend()                    // 扉画像のような透過テクスチャを正しく合成するため
			ctx.ClearColor(0.53, 0.75, 0.9, 1.0) // 空っぽい水色

			scene, link, doorIndex, err := renderer.BuildFieldDemoScene(ctx)
			if err != nil {
				fmt.Println("renderer: failed to build demo scene:", err)
			} else {
				g := game.New(scene, link, doorIndex, func() {
					ctx.Navigate("grassland.html")
				})
				game.SetInstance(g)

				ctx.RunLoop(func(dt float64) {
					g.Update(dt)
					scene.Render(ctx)
				})
				fmt.Println("renderer: demo scene rendered, game loop started")
			}
		}

		select {}
	}
}
