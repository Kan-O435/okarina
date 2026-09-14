//go:build !js

// Package bridge は、ブラウザのJavaScriptとGo WASM間の相互呼び出しを仲介する。
// ネイティブ(非WASM)ビルドでは syscall/js が使えないため、
// go build ./... やネイティブ実行が通るように何もしない実装を提供する。
package bridge

// Init はネイティブビルドでは何もしない。
func Init() {}

// CallConsoleLog はネイティブビルドでは何もしない。
func CallConsoleLog(args ...interface{}) {}
