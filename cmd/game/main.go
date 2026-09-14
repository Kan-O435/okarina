package main

import (
	"fmt"
	"runtime"

	"github.com/Kan-O435/okarina/internal/bridge"
	"github.com/Kan-O435/okarina/internal/game"
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
		select {}
	}
}
