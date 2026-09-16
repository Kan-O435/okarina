package main

import (
	"fmt"
	"runtime"

	"github.com/Kan-O435/okarina/internal/bridge"
	"github.com/Kan-O435/okarina/internal/game"
	"github.com/Kan-O435/okarina/internal/renderer"
)

// ストーリー・操作方法説明画面用のエントリーポイント。タイトル画面
// (cmd/title)で「ド(C)」が弾かれた後、神殿フィールド(temple.html)へ
// 直接遷移する前にこのページを挟む。本文はHTML側(web/story.html)の
// テキストボックスで表示するが、その背後にganon.goと同じガノン
// (renderer.BuildStoryScene参照)を浮かべ、これから挑む相手の存在感を
// 出している。このページでも「ド」を弾く(またはOキー)と神殿フィールドへ
// 進む。タイトル画面と同じgame.SetTitleStartTrigger/OnMIDIEventの仕組み
// をそのまま再利用する(ページごとに別プロセスとして起動するため、
// トリガーの登録先が競合することはない)。
func main() {
	fmt.Println("Story screen initialized")

	if runtime.GOOS == "js" {
		bridge.Init()

		ctx, err := renderer.NewContext("game-canvas")
		if err != nil {
			fmt.Println("renderer: failed to initialize:", err)
		} else {
			width, height := ctx.CanvasSize()
			ctx.Viewport(width, height)
			ctx.EnableDepthTest()
			ctx.EnableBlend()                    // 背景の炎(半透明テクスチャ)を正しく合成するため
			ctx.ClearColor(0.05, 0.05, 0.1, 1.0) // タイトル画面と同じ暗い背景

			scene, flameAnim, sceneErr := renderer.BuildStoryScene(ctx)
			background, bgErr := renderer.BuildStoryBackground(ctx, width, height)
			if sceneErr != nil {
				fmt.Println("renderer: failed to build story scene:", sceneErr)
			} else if bgErr != nil {
				fmt.Println("renderer: failed to build story background:", bgErr)
			} else {
				ctx.RunLoop(func(dt float64) {
					background.Update(dt)
					flameAnim.Update(dt)
					ctx.Clear()
					background.Render(ctx, width, height)
					scene.RenderWithoutClear(ctx)
				})
			}
		}

		game.SetTitleStartTrigger(func() {
			ctx.Navigate("temple.html")
		})

		select {}
	}
}
