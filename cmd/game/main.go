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

		// WebGL1コンテキストが取得できるかの疎通確認。
		// シェーダー・メッシュを使った実際の描画は今後実装する。
		ctx, err := renderer.NewContext("game-canvas")
		if err != nil {
			fmt.Println("renderer: failed to initialize:", err)
		} else {
			ctx.ClearColor(0.1, 0.15, 0.25, 1.0)
			ctx.Clear()
			fmt.Println("renderer: WebGL1 context initialized")
		}

		select {}
	}
}
