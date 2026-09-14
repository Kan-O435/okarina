package music

import (
	"fmt"
	"sync"
	"time"

	"github.com/Kan-O435/okarina/internal/midi"
)

// activeNote はNote ONを受け取ったがまだNote OFFが来ていない音符の情報。
type activeNote struct {
	pitch     Pitch
	velocity  int
	startTime time.Time
}

// Recorder はMIDIイベントを受け取り、演奏された音符を時系列で記録する。
// 一定時間(IdleTimeout)入力がない場合、そこまでの記録を1回の演奏(Melody)として確定する。
type Recorder struct {
	mu          sync.Mutex
	idleTimeout time.Duration
	active      map[int]activeNote // MIDIノート番号 -> 押下中の音符
	current     Melody
	onDone      func(Melody)
	timer       *time.Timer
}

// NewRecorder は、演奏終了とみなす無音時間(idleTimeout)と、
// 1回の演奏が確定した際に呼び出すコールバックを指定してRecorderを作成する。
func NewRecorder(idleTimeout time.Duration, onDone func(Melody)) *Recorder {
	return &Recorder{
		idleTimeout: idleTimeout,
		active:      make(map[int]activeNote),
		onDone:      onDone,
	}
}

// HandleEvent はMIDIイベントを1件処理する。
// Note ONで音符の記録を開始し、対応するNote OFFが来た時点でNoteとして確定する。
func (r *Recorder) HandleEvent(e midi.Event) {
	r.mu.Lock()

	now := time.Now()

	if e.IsNoteOn {
		if len(r.current) == 0 && len(r.active) == 0 {
			fmt.Println("[music] performance started")
		}
		r.active[e.Note] = activeNote{
			pitch:     PitchFromMIDINote(e.Note),
			velocity:  e.Velocity,
			startTime: now,
		}
	} else if an, ok := r.active[e.Note]; ok {
		r.current = append(r.current, Note{
			Pitch:     an.pitch,
			Velocity:  an.velocity,
			StartTime: an.startTime,
			EndTime:   now,
		})
		delete(r.active, e.Note)
	}

	r.resetIdleTimerLocked()
	r.mu.Unlock()
}

// resetIdleTimerLocked はidleTimeout後にfinishを実行するタイマーを再設定する。
// 呼び出し前にr.muがロックされている必要がある。
func (r *Recorder) resetIdleTimerLocked() {
	if r.timer != nil {
		r.timer.Stop()
	}
	r.timer = time.AfterFunc(r.idleTimeout, r.finish)
}

// finish は無音状態が続いた際に呼ばれ、記録済みの音符を1つのMelodyとして確定する。
func (r *Recorder) finish() {
	r.mu.Lock()
	melody := r.current
	r.current = nil
	r.mu.Unlock()

	if len(melody) == 0 {
		return
	}

	fmt.Println("[music] performance ended")
	if r.onDone != nil {
		r.onDone(melody)
	}
}
