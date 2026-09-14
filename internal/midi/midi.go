// Package midi は、MIDIキーボード(MPK Mini等)からの入力を受け取り、
// ノートイベントの解析・メロディパターンの判定を担当する。
//
// ブラウザ環境では Web MIDI API から渡されるイベントを
// JS-Go Bridge (internal/bridge) 経由で受け取る想定。
package midi

import "fmt"

// Event は1件のMIDIノートイベントを表す。
type Event struct {
	Note      int     // MIDIノート番号 (0-127)
	Velocity  int     // ベロシティ (0-127)
	IsNoteOn  bool    // true: Note On, false: Note Off
	Timestamp float64 // イベント発生時刻 (JS側のperformance.now()由来、ミリ秒)
}

// HandleEvent はブリッジ経由で受け取ったMIDIイベントを処理する。
// 現時点ではログ出力のみ。実際のノート解析・メロディ判定は別途実装する。
func HandleEvent(e Event) {
	fmt.Printf("[midi] note=%d velocity=%d isNoteOn=%v timestamp=%.2f\n",
		e.Note, e.Velocity, e.IsNoteOn, e.Timestamp)
}
