//go:build js

// Package bridge は、ブラウザのJavaScriptとGo WASM間の相互呼び出しを仲介する。
package bridge

import (
	"fmt"
	"syscall/js"

	"github.com/Kan-O435/okarina/internal/game"
	"github.com/Kan-O435/okarina/internal/midi"
)

// Init はGo側の関数をJavaScriptのグローバルオブジェクトに公開し、
// JSからGoの関数を呼び出せるようにする。
func Init() {
	js.Global().Set("goPing", js.FuncOf(goPing))
	js.Global().Set("goOnMIDIEvent", js.FuncOf(goOnMIDIEvent))
	js.Global().Set("goOnPitchDetected", js.FuncOf(goOnPitchDetected))
	js.Global().Set("goUpdateFrame", js.FuncOf(goUpdateFrame))
	js.Global().Set("goGetPlayerDirection", js.FuncOf(goGetPlayerDirection))
	CallConsoleLog("bridge: Go functions registered (goPing, goOnMIDIEvent, goOnPitchDetected, goUpdateFrame, goGetPlayerDirection)")
}

// CallConsoleLog はGoからJavaScriptのconsole.logを呼び出す(Go→JSの実演)。
func CallConsoleLog(args ...interface{}) {
	js.Global().Get("console").Call("log", args...)
}

// goPing はJavaScriptから呼び出されるGo関数(JS→Goの実演)。
func goPing(this js.Value, args []js.Value) interface{} {
	CallConsoleLog("bridge: goPing() called from JavaScript")
	return "pong from Go"
}

// goOnMIDIEvent はJavaScript(Web MIDI API)側からMIDIイベントを渡すための
// エントリーポイント。引数: note, velocity, isNoteOn, timestamp。
// 実際のMIDIデバイス検出・入力取得は別途実装し、ここではGoへイベントを
// 渡す構造のみを用意する。
func goOnMIDIEvent(this js.Value, args []js.Value) interface{} {
	if len(args) < 4 {
		CallConsoleLog("bridge: goOnMIDIEvent expects 4 args (note, velocity, isNoteOn, timestamp)")
		return nil
	}
	event := midi.Event{
		Note:      args[0].Int(),
		Velocity:  args[1].Int(),
		IsNoteOn:  args[2].Bool(),
		Timestamp: args[3].Float(),
	}
	midi.HandleEvent(event)
	game.OnMIDIEvent(event)
	CallConsoleLog(fmt.Sprintf("bridge: MIDI event forwarded to Go: %+v", event))
	return nil
}

// goOnPitchDetected はJavaScript(マイク入力のピッチ検出)側から、検出した
// 周波数(Hz)を渡すためのエントリーポイント。音量不足などでピッチが
// 検出できなかった場合は0以下を渡す。
func goOnPitchDetected(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		CallConsoleLog("bridge: goOnPitchDetected expects 1 arg (frequencyHz)")
		return nil
	}
	game.OnPitchDetected(args[0].Float())
	return nil
}

// goUpdateFrame はJavaScript側のrequestAnimationFrameループから毎フレーム
// 呼び出され、プレイヤーの移動と再描画を進める。引数は前フレームからの
// 経過時間(秒)。
func goUpdateFrame(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		CallConsoleLog("bridge: goUpdateFrame expects 1 arg (deltaTimeSeconds)")
		return nil
	}
	game.UpdateFrame(args[0].Float())
	return nil
}

// goGetPlayerDirection はプレイヤーの現在の移動方向をJS側に返す
// ("forward" | "backward" | "idle")。UI表示等に使う。
func goGetPlayerDirection(this js.Value, args []js.Value) interface{} {
	return game.PlayerDirection()
}
