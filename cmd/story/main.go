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
// 直接遷移する前にこのページを挟む。3Dシーンは不要なため、背景色の
// クリアのみ行い、本文はHTML側(web/story.html)のテキストボックスで
// 表示する。このページでも「ド」を弾く(またはOキー)と神殿フィールドへ
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
			ctx.ClearColor(0.05, 0.05, 0.1, 1.0) // タイトル画面と同じ暗い背景
			ctx.Clear()
		}

		game.SetTitleStartTrigger(func() {
			ctx.Navigate("temple.html")
		})

		select {}
	}
}
