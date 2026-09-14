package main

import (
	"fmt"
	"runtime"

	"github.com/Kan-O435/okarina/internal/renderer"
)

// 草原フィールド(扉を抜けた先)用のエントリーポイント。今はまだ地面だけの
// 最小構成で、神殿フィールド(cmd/game)と同じくここから機能を足していく。
func main() {
	fmt.Println("Grassland field initialized")

	if runtime.GOOS == "js" {
		ctx, err := renderer.NewContext("game-canvas")
		if err != nil {
			fmt.Println("renderer: failed to initialize:", err)
		} else {
			width, height := ctx.CanvasSize()
			ctx.Viewport(width, height)
			ctx.EnableDepthTest()
			ctx.ClearColor(0.53, 0.75, 0.9, 1.0) // 空っぽい水色

			scene, err := renderer.BuildGrasslandScene(ctx)
			if err != nil {
				fmt.Println("renderer: failed to build grassland scene:", err)
			} else {
				scene.Render(ctx)
				fmt.Println("renderer: grassland scene rendered")
			}
		}

		select {}
	}
}
