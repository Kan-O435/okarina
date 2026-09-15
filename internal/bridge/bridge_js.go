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
	js.Global().Set("goOpenDoor", js.FuncOf(goOpenDoor))
	js.Global().Set("goOnPitchDetected", js.FuncOf(goOnPitchDetected))
	js.Global().Set("goGetPlayerDirection", js.FuncOf(goGetPlayerDirection))
	js.Global().Set("goSetDebugDirection", js.FuncOf(goSetDebugDirection))
	js.Global().Set("goDefeatGanonFirstForm", js.FuncOf(goDefeatGanonFirstForm))
	js.Global().Set("goPreviewGanonHallCollapse", js.FuncOf(goPreviewGanonHallCollapse))
	js.Global().Set("goSummonHorse", js.FuncOf(goSummonHorse))
	CallConsoleLog("bridge: Go functions registered (goPing, goOnMIDIEvent, goOpenDoor, goOnPitchDetected, goGetPlayerDirection, goSetDebugDirection, goDefeatGanonFirstForm, goPreviewGanonHallCollapse, goSummonHorse)")

	// JS側(web/audio.jsのplayNote/stopNote/playConfirmationFanfare、
	// web/grassland.jsのplayHorseJumpSound)の実装をgameパッケージに
	// 差し込む。これにより、Goから「時の歌」の続きや効果音などを
	// 自動再生できる。
	game.SetPlayNoteFunc(callPlayNote)
	game.SetStopNoteFunc(callStopNote)
	game.SetPlayConfirmationFanfareFunc(callPlayConfirmationFanfare)
	game.SetPlayHorseJumpSoundFunc(callPlayHorseJumpSound)
}

// callPlayNote はGoからJavaScript側のplayNote(note, velocity)を呼び出す。
func callPlayNote(note, velocity int) {
	js.Global().Call("playNote", note, velocity)
}

// callStopNote はGoからJavaScript側のstopNote(note)を呼び出す。
func callStopNote(note int) {
	js.Global().Call("stopNote", note)
}

// callPlayConfirmationFanfare はGoからJavaScript側の
// playConfirmationFanfare()を呼び出し、時の歌・馬の歌の確認音
// (「テレレレレ」、mp3の効果音)を再生する。
func callPlayConfirmationFanfare() {
	js.Global().Call("playConfirmationFanfare")
}

// callPlayHorseJumpSound はGoからJavaScript側のplayHorseJumpSound()を
// 呼び出し、馬がジャンプした際のいななき効果音を再生する
// (web/grassland.js、草原フィールドのみで定義される)。
func callPlayHorseJumpSound() {
	js.Global().Call("playHorseJumpSound")
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

// goOpenDoor はJavaScript側から呼び出され、隠し扉を開く。
// 現時点では動作確認用のボタンから直接呼ぶ想定。将来的にはMIDIのメロディ
// 認識が成功した際にGo側(game.OpenDoor())から呼ばれる形に置き換える。
func goOpenDoor(this js.Value, args []js.Value) interface{} {
	game.OpenDoor()
	CallConsoleLog("bridge: goOpenDoor() called, opening secret door")
	return nil
}

// goOnPitchDetected はJavaScript(マイク入力のピッチ検出)側から、検出した
// 周波数(Hz)を渡すためのエントリーポイント。音量不足などでピッチが
// 検出できなかった場合は0以下を渡す。実際のプレイヤー移動は、Go側で
// 常時回っているゲームループ(Context.RunLoop、cmd/game/main.go参照)が
// 毎フレームplayer.Playerの状態を読んで進める。
func goOnPitchDetected(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		CallConsoleLog("bridge: goOnPitchDetected expects 1 arg (frequencyHz)")
		return nil
	}
	game.OnPitchDetected(args[0].Float())
	return nil
}

// goGetPlayerDirection はプレイヤーの現在の移動方向をJS側に返す
// ("forward" | "backward" | "idle")。UI表示等に使う。
func goGetPlayerDirection(this js.Value, args []js.Value) interface{} {
	return game.PlayerDirection()
}

// goSetDebugDirection はJavaScript側(デバッグ用の矢印キー操作)から、
// オタマトーンのピッチ入力を介さずにプレイヤーの移動方向を直接指定する。
// 引数は"forward" | "backward" | "idle"のいずれか。
func goSetDebugDirection(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		CallConsoleLog("bridge: goSetDebugDirection expects 1 arg (direction)")
		return nil
	}
	game.SetDebugDirection(args[0].String())
	return nil
}

// goDefeatGanonFirstForm はJavaScript側(玉座の間のデバッグボタン)から
// 呼び出され、Ganonの第一形態を倒した演出(玉座の間の崩落→戦場跡
// フィールドへのページ遷移)を開始する。「特定の演奏で倒す」処理はまだ
// 無いため、現時点では動作確認用のボタンから直接呼ぶ想定
// (goOpenDoorと同様の仮実装パターン)。
func goDefeatGanonFirstForm(this js.Value, args []js.Value) interface{} {
	game.TriggerGanonHallCollapse()
	CallConsoleLog("bridge: goDefeatGanonFirstForm() called, starting hall collapse")
	return nil
}

// goPreviewGanonHallCollapse はJavaScript側(玉座の間のデバッグ「テスト再生」
// ボタン)から呼び出され、崩落演出を戦場跡フィールドへのページ遷移なしで
// その場で再生する(何度でも試せる、動作確認用のプレビュー)。
func goPreviewGanonHallCollapse(this js.Value, args []js.Value) interface{} {
	game.TriggerGanonHallCollapsePreview()
	CallConsoleLog("bridge: goPreviewGanonHallCollapse() called, previewing hall collapse")
	return nil
}

// goSummonHorse はJavaScript側(草原フィールドのデバッグ「馬に乗る」
// ボタン)から呼び出され、馬の歌を演奏したときと同じ処理(馬を呼び出す)を
// 演奏なしで直接実行する。
func goSummonHorse(this js.Value, args []js.Value) interface{} {
	game.TriggerHorseSummon()
	CallConsoleLog("bridge: goSummonHorse() called, summoning horse")
	return nil
}
