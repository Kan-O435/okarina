package game

import (
	"sync"
	"testing"
	"time"

	"github.com/Kan-O435/okarina/internal/music"
	"github.com/Kan-O435/okarina/internal/player"
	"github.com/Kan-O435/okarina/internal/renderer"
	"github.com/Kan-O435/okarina/internal/world"
)

var songOfTime = music.Melody{
	{Pitch: music.A}, {Pitch: music.D}, {Pitch: music.F},
	{Pitch: music.A}, {Pitch: music.D}, {Pitch: music.F},
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
