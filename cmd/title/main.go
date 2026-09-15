package main

import (
	"fmt"
	"runtime"

	"github.com/Kan-O435/okarina/internal/bridge"
	"github.com/Kan-O435/okarina/internal/game"
	"github.com/Kan-O435/okarina/internal/renderer"
)

// タイトル画面用のエントリーポイント。見た目はまだ仮の最小構成で、
// MIDIキーボードで「ド(C、オクターブ不問)」を弾くと神殿フィールド
// (temple.html)へ遷移する導線だけを用意する。
func main() {
	fmt.Println("Title screen initialized")

	if runtime.GOOS == "js" {
		bridge.Init()

		ctx, err := renderer.NewContext("game-canvas")
		if err != nil {
			fmt.Println("renderer: failed to initialize:", err)
		} else {
			ctx.ClearColor(0.05, 0.05, 0.1, 1.0) // 仮の暗い背景
			ctx.Clear()

			game.SetTitleStartTrigger(func() {
				ctx.Navigate("temple.html")
			})
			fmt.Println("title: waiting for C to start")
		}

		select {}
	}
}
