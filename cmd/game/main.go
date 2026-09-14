package main

import (
	"fmt"
	"runtime"

	"github.com/Kan-O435/okarina/internal/game"
)

func main() {
	fmt.Println("Game initialized")
	game.Run()

	// WASM(ブラウザ)で動かす場合、main()がreturnするとプログラムが
	// 終了してしまい、以後 JS 側から Go の関数を呼べなくなる。
	// #3, #6 で JS-Go Bridge を作る際に困らないよう、
	// js/wasm ビルドの時だけプログラムを常駐させる。
	if runtime.GOOS == "js" {
		select {}
	}
}
