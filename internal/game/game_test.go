package game

import (
	"sync"
	"testing"
	"time"

	"github.com/Kan-O435/okarina/internal/midi"
	"github.com/Kan-O435/okarina/internal/music"
	"github.com/Kan-O435/okarina/internal/player"
	"github.com/Kan-O435/okarina/internal/renderer"
	"github.com/Kan-O435/okarina/internal/world"
)

var songOfTime = music.Melody{
	{Pitch: music.A}, {Pitch: music.D}, {Pitch: music.F},
	{Pitch: music.A}, {Pitch: music.D}, {Pitch: music.F},
}

var horseSong = music.Melody{
	{Pitch: music.D}, {Pitch: music.B}, {Pitch: music.A},
	{Pitch: music.D}, {Pitch: music.B}, {Pitch: music.A},
}

func TestOnMelodyRecorded_SongOfTimeCorrect(t *testing.T) {
	songOfTimePlayed = false

	onMelodyRecorded(songOfTime)

	if !SongOfTimePlayed() {
		t.Fatal("expected SongOfTimePlayed() to be true after playing ラレファラレファ correctly")
	}
}

func TestOnMelodyRecorded_UnrecognizedMelodyDoesNotSetFlag(t *testing.T) {
	songOfTimePlayed = false
	melody := music.Melody{{Pitch: music.C}, {Pitch: music.C}}

	onMelodyRecorded(melody)

	if SongOfTimePlayed() {
		t.Fatal("expected SongOfTimePlayed() to remain false for an unrecognized melody")
	}
}

func TestOnMelodyRecorded_PartialSongDoesNotSetFlag(t *testing.T) {
	songOfTimePlayed = false
	// 前半(ラ→レ→ファ)だけでは不十分。
	melody := music.Melody{{Pitch: music.A}, {Pitch: music.D}, {Pitch: music.F}}

	onMelodyRecorded(melody)

	if SongOfTimePlayed() {
		t.Fatal("expected SongOfTimePlayed() to remain false for a partial performance")
	}
}

func TestOnMelodyRecorded_SongOfTimeOpensDoorAfterDelay(t *testing.T) {
	songOfTimePlayed = false
	g := &Game{
		scene:     &renderer.Scene{Objects: []renderer.Object{{}}},
		doorIndex: 0,
	}
	SetInstance(g)
	defer SetInstance(nil)

	// 扉が開くまでの「ため」の時間はテストでは短くしておく。
	origDelay := doorOpenDelay
	doorOpenDelay = time.Millisecond
	defer func() { doorOpenDelay = origDelay }()

	onMelodyRecorded(songOfTime)

	if g.door.State != world.DoorClosed {
		t.Fatalf("expected door to remain closed immediately after recognition, got state=%v", g.door.State)
	}

	time.Sleep(20 * time.Millisecond)

	if g.door.State != world.DoorOpening {
		t.Fatalf("expected door to start opening after the delay, got state=%v", g.door.State)
	}
}

func TestIsNearDoor(t *testing.T) {
	g := &Game{}
	SetInstance(g)
	defer SetInstance(nil)

	player.Player.Z = renderer.DoorCenterZ
	if !IsNearDoor() {
		t.Error("expected IsNearDoor() to be true right at the door")
	}

	player.Player.Z = renderer.DoorCenterZ + nearDoorRangeZ
	if !IsNearDoor() {
		t.Error("expected IsNearDoor() to be true at the edge of the range")
	}

	player.Player.Z = renderer.DoorCenterZ + nearDoorRangeZ + 1
	if IsNearDoor() {
		t.Error("expected IsNearDoor() to be false just outside the range")
	}
}

func TestIsNearDoor_NoInstance(t *testing.T) {
	SetInstance(nil)
	if IsNearDoor() {
		t.Error("expected IsNearDoor() to be false when no Game instance is registered")
	}
}

// setSleepHookForTest はtime.Sleepの差し替え用フックを、他のフックと同じ
// mutex経由で差し替える(playSongOfTimeContinuationとのデータ競合を防ぐ)。
func setSleepHookForTest(f func(time.Duration)) (restore func()) {
	audioHooksMu.Lock()
	orig := sleepHook
	sleepHook = f
	audioHooksMu.Unlock()
	return func() {
		audioHooksMu.Lock()
		sleepHook = orig
		audioHooksMu.Unlock()
	}
}

func TestPlaySongOfTimeAudio_PlaysConfirmationThenContinuation(t *testing.T) {
	defer setSleepHookForTest(func(time.Duration) {})()

	var played, stopped []int
	SetPlayNoteFunc(func(note, velocity int) { played = append(played, note) })
	SetStopNoteFunc(func(note int) { stopped = append(stopped, note) })
	defer func() { SetPlayNoteFunc(nil); SetStopNoteFunc(nil) }()

	playSongOfTimeAudio()

	want := append(append([]music.ContinuationNote{}, music.SongOfTimeConfirmation...), music.SongOfTimeOpening...)
	want = append(want, music.SongOfTimeContinuation...)
	if len(played) != len(want) {
		t.Fatalf("played %d notes, want %d", len(played), len(want))
	}
	for i, n := range want {
		if played[i] != n.MIDINote {
			t.Errorf("played[%d] = %d, want %d", i, played[i], n.MIDINote)
		}
		if stopped[i] != n.MIDINote {
			t.Errorf("stopped[%d] = %d, want %d", i, stopped[i], n.MIDINote)
		}
	}
}

func TestOnMelodyRecorded_SongOfTimeTriggersConfirmationAndContinuation(t *testing.T) {
	songOfTimePlayed = false
	defer setSleepHookForTest(func(time.Duration) {})()

	wantTotal := len(music.SongOfTimeConfirmation) + len(music.SongOfTimeOpening) + len(music.SongOfTimeContinuation)

	done := make(chan struct{})
	var mu sync.Mutex
	var played []int
	SetPlayNoteFunc(func(note, velocity int) {
		mu.Lock()
		played = append(played, note)
		n := len(played)
		mu.Unlock()
		if n == wantTotal {
			close(done)
		}
	})
	SetStopNoteFunc(func(note int) {})
	defer func() { SetPlayNoteFunc(nil); SetStopNoteFunc(nil) }()

	onMelodyRecorded(songOfTime)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("confirmation/continuation was not played within timeout")
	}
}

func TestOnMelodyRecorded_HorseSongCorrect(t *testing.T) {
	horseSongPlayed = false

	onMelodyRecorded(horseSong)

	if !HorseSongPlayed() {
		t.Fatal("expected HorseSongPlayed() to be true after playing レシラレシラ correctly")
	}
}

func TestOnMelodyRecorded_HorseSongSummonsHorse(t *testing.T) {
	horseSongPlayed = false
	summoned := false
	SetHorseSummoner(func() { summoned = true })
	defer SetHorseSummoner(nil)

	onMelodyRecorded(horseSong)

	if !summoned {
		t.Fatal("expected the registered horse summoner to be called")
	}
}

func TestOnMelodyRecorded_HorseSongTriggersConfirmationAndContinuation(t *testing.T) {
	horseSongPlayed = false
	defer setSleepHookForTest(func(time.Duration) {})()

	wantTotal := len(music.SongOfTimeConfirmation) + len(music.HorseSongContinuation)

	done := make(chan struct{})
	var mu sync.Mutex
	var played []int
	SetPlayNoteFunc(func(note, velocity int) {
		mu.Lock()
		played = append(played, note)
		n := len(played)
		mu.Unlock()
		if n == wantTotal {
			close(done)
		}
	})
	SetStopNoteFunc(func(note int) {})
	defer func() { SetPlayNoteFunc(nil); SetStopNoteFunc(nil) }()

	onMelodyRecorded(horseSong)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("confirmation/continuation was not played within timeout")
	}
}

// resetJumpGesture はジャンプジェスチャーの検出状態をテスト用に初期化する。
func resetJumpGesture() {
	jumpGestureLow = false
	jumpGestureCount = 0
}

func TestOnPitchDetected_TwoLowNotesTriggerJump(t *testing.T) {
	resetJumpGesture()
	triggered := 0
	SetJumpTrigger(func() { triggered++ })
	defer SetJumpTrigger(nil)

	OnPitchDetected(100) // 1回目の低い音(立ち上がり)
	if triggered != 0 {
		t.Fatalf("expected no trigger after only one low note, got %d", triggered)
	}

	OnPitchDetected(0)   // 無音(1回目の低い音が終わる)
	OnPitchDetected(120) // 2回目の低い音(立ち上がり) → 発火するはず
	if triggered != 1 {
		t.Fatalf("expected exactly one trigger after two low notes, got %d", triggered)
	}
}

func TestOnPitchDetected_SustainedLowNoteCountsOnce(t *testing.T) {
	resetJumpGesture()
	triggered := 0
	SetJumpTrigger(func() { triggered++ })
	defer SetJumpTrigger(nil)

	// 同じ低い音を無音を挟まず連続で読み取っても、1回とカウントする。
	for i := 0; i < 5; i++ {
		OnPitchDetected(100)
	}
	if triggered != 0 {
		t.Fatalf("expected sustained low pitch readings to count as a single onset, got %d triggers", triggered)
	}
}

func TestOnPitchDetected_HighNoteResetsJumpGesture(t *testing.T) {
	resetJumpGesture()
	triggered := 0
	SetJumpTrigger(func() { triggered++ })
	defer SetJumpTrigger(nil)

	OnPitchDetected(100) // 1回目の低い音
	OnPitchDetected(0)
	OnPitchDetected(440) // 明確に高い音 → カウントリセット
	OnPitchDetected(0)
	OnPitchDetected(120) // これは(リセット後の)1回目扱いのはず
	if triggered != 0 {
		t.Fatalf("expected the gesture count to reset after a clearly high note, got %d triggers", triggered)
	}
}

func TestOnMIDIEvent_NoteCTriggersTitleStart(t *testing.T) {
	triggered := 0
	SetTitleStartTrigger(func() { triggered++ })
	defer SetTitleStartTrigger(nil)

	OnMIDIEvent(midi.Event{Note: 60, Velocity: 100, IsNoteOn: true}) // C4
	if triggered != 1 {
		t.Fatalf("expected exactly one trigger for a C Note On, got %d", triggered)
	}

	OnMIDIEvent(midi.Event{Note: 72, Velocity: 100, IsNoteOn: true}) // C5(オクターブ違い)
	if triggered != 2 {
		t.Fatalf("expected the trigger to fire regardless of octave, got %d", triggered)
	}
}

func TestOnMIDIEvent_NonCNoteDoesNotTriggerTitleStart(t *testing.T) {
	triggered := 0
	SetTitleStartTrigger(func() { triggered++ })
	defer SetTitleStartTrigger(nil)

	OnMIDIEvent(midi.Event{Note: 62, Velocity: 100, IsNoteOn: true}) // D4
	if triggered != 0 {
		t.Fatalf("expected no trigger for a non-C note, got %d", triggered)
	}
}

func TestOnMIDIEvent_NoteOffDoesNotTriggerTitleStart(t *testing.T) {
	triggered := 0
	SetTitleStartTrigger(func() { triggered++ })
	defer SetTitleStartTrigger(nil)

	OnMIDIEvent(midi.Event{Note: 60, Velocity: 0, IsNoteOn: false}) // C4 Note Off
	if triggered != 0 {
		t.Fatalf("expected no trigger for a Note Off, got %d", triggered)
	}
}
