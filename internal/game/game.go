// Package game は、ゲーム全体のループ・状態管理を担当する。
package game

import (
	"fmt"
	"time"

	"github.com/Kan-O435/okarina/internal/midi"
	"github.com/Kan-O435/okarina/internal/music"
)

// melodyIdleTimeout は、この時間MIDI入力がなければ1回の演奏が
// 終わったとみなす無音時間。
const melodyIdleTimeout = 2 * time.Second

var recorder = music.NewRecorder(melodyIdleTimeout, onMelodyRecorded)

// Run はゲームのエントリーポイント。
// 現時点ではプロジェクトの雛形確認用の最小実装。
func Run() {
	fmt.Println("game.Run() called")
}

// OnMIDIEvent はJS-Go Bridge経由で受け取ったMIDIイベントを
// 旋律記録エンジン(internal/music.Recorder)に渡す。
func OnMIDIEvent(e midi.Event) {
	recorder.HandleEvent(e)
}

// onMelodyRecorded は一連の演奏が確定した際に呼ばれ、
// 登録済みの旋律パターンと照合する。
func onMelodyRecorded(melody music.Melody) {
	fmt.Printf("[music] melody recorded: %v\n", melody.Pitches())

	if name := music.Recognize(melody, music.DefaultPatterns); name != "" {
		fmt.Printf("[music] recognized: %s\n", name)
	} else {
		fmt.Println("[music] recognized: (no match)")
	}
}
