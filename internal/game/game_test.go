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

var ganonHallMelody = music.Melody{
	{Pitch: music.D}, {Pitch: music.A}, {Pitch: music.D},
	{Pitch: music.A}, {Pitch: music.B}, {Pitch: music.D},
}

var ganonBattleMelody = music.Melody{
	{Pitch: music.D}, {Pitch: music.F}, {Pitch: music.D},
	{Pitch: music.D}, {Pitch: music.F}, {Pitch: music.D},
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

func TestPlayConfirmationFanfare_CallsHookAndSleepsForItsDuration(t *testing.T) {
	called := 0
	var sleptFor time.Duration
	playConfirmationFanfare(func() { called++ }, func(d time.Duration) { sleptFor = d })

	if called != 1 {
		t.Fatalf("expected the fanfare hook to be called once, got %d", called)
	}
	if sleptFor != music.ConfirmationFanfareDuration {
		t.Errorf("slept for %v, want %v (music.ConfirmationFanfareDuration)", sleptFor, music.ConfirmationFanfareDuration)
	}
}

func TestPlayConfirmationFanfare_NilHookDoesNothing(t *testing.T) {
	slept := false
	playConfirmationFanfare(nil, func(time.Duration) { slept = true })

	if slept {
		t.Fatal("expected no sleep when the fanfare hook is not registered")
	}
}

func TestPlaySongOfTimeAudio_PlaysConfirmationFanfareThenContinuation(t *testing.T) {
	defer setSleepHookForTest(func(time.Duration) {})()

	fanfareCalls := 0
	SetPlayConfirmationFanfareFunc(func() { fanfareCalls++ })
	var played, stopped []int
	SetPlayNoteFunc(func(note, velocity int) { played = append(played, note) })
	SetStopNoteFunc(func(note int) { stopped = append(stopped, note) })
	defer func() { SetPlayConfirmationFanfareFunc(nil); SetPlayNoteFunc(nil); SetStopNoteFunc(nil) }()

	playSongOfTimeAudio()

	if fanfareCalls != 1 {
		t.Fatalf("expected the confirmation fanfare (mp3) to be played once, got %d calls", fanfareCalls)
	}

	want := append(append([]music.ContinuationNote{}, music.SongOfTimeOpening...), music.SongOfTimeContinuation...)
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

	wantTotal := len(music.SongOfTimeOpening) + len(music.SongOfTimeContinuation)

	done := make(chan struct{})
	var mu sync.Mutex
	var played []int
	SetPlayConfirmationFanfareFunc(func() {})
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
	defer func() { SetPlayConfirmationFanfareFunc(nil); SetPlayNoteFunc(nil); SetStopNoteFunc(nil) }()

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

func TestOnMelodyRecorded_GanonHallMelodyCorrect(t *testing.T) {
	ganonHallMelodyPlayed = false
	SetGanonHallCollapseTrigger(func() {})
	defer SetGanonHallCollapseTrigger(nil)

	onMelodyRecorded(ganonHallMelody)

	if !GanonHallMelodyPlayed() {
		t.Fatal("expected GanonHallMelodyPlayed() to be true after playing レラレラシレ correctly")
	}
}

func TestOnMelodyRecorded_GanonHallMelodyPlaysConfirmationThenTriggersCollapse(t *testing.T) {
	ganonHallMelodyPlayed = false
	defer setSleepHookForTest(func(time.Duration) {})()

	fanfareCalls := 0
	done := make(chan struct{})
	SetPlayConfirmationFanfareFunc(func() { fanfareCalls++ })
	SetGanonHallCollapseTrigger(func() { close(done) })
	defer func() {
		SetPlayConfirmationFanfareFunc(nil)
		SetGanonHallCollapseTrigger(nil)
	}()

	onMelodyRecorded(ganonHallMelody)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ganon hall collapse trigger was not called within timeout")
	}

	if fanfareCalls != 1 {
		t.Fatalf("expected the confirmation fanfare to be played once before the collapse, got %d calls", fanfareCalls)
	}
}

func TestOnMelodyRecorded_GanonHallMelodyWithNoTriggerRegisteredDoesNothing(t *testing.T) {
	ganonHallMelodyPlayed = false
	SetGanonHallCollapseTrigger(nil)

	onMelodyRecorded(ganonHallMelody) // パニックしないことを確認する

	if !GanonHallMelodyPlayed() {
		t.Fatal("expected GanonHallMelodyPlayed() to still be set even without a registered trigger")
	}
}

func TestOnMelodyRecorded_GanonBattleMelodyCorrect(t *testing.T) {
	ganonBattleMelodyPlayed = false
	SetGanonBattleDefeatTrigger(func() {})
	defer SetGanonBattleDefeatTrigger(nil)

	onMelodyRecorded(ganonBattleMelody)

	if !GanonBattleMelodyPlayed() {
		t.Fatal("expected GanonBattleMelodyPlayed() to be true after playing レファレレファレ correctly")
	}
}

func TestOnMelodyRecorded_GanonBattleMelodyPlaysConfirmationThenSongThenTriggersDefeat(t *testing.T) {
	ganonBattleMelodyPlayed = false
	defer setSleepHookForTest(func(time.Duration) {})()

	fanfareCalls, songCalls := 0, 0
	done := make(chan struct{})
	SetPlayConfirmationFanfareFunc(func() { fanfareCalls++ })
	SetPlayGanonBattleSongFunc(func() { songCalls++ })
	SetGanonBattleDefeatTrigger(func() { close(done) })
	defer func() {
		SetPlayConfirmationFanfareFunc(nil)
		SetPlayGanonBattleSongFunc(nil)
		SetGanonBattleDefeatTrigger(nil)
	}()

	onMelodyRecorded(ganonBattleMelody)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ganon battle defeat trigger was not called within timeout")
	}

	if fanfareCalls != 1 {
		t.Fatalf("expected the confirmation fanfare to be played once, got %d calls", fanfareCalls)
	}
	if songCalls != 1 {
		t.Fatalf("expected the song of storms BGM to be played once, got %d calls", songCalls)
	}
}

func TestOnMelodyRecorded_GanonBattleMelodyWithNoTriggerRegisteredDoesNothing(t *testing.T) {
	ganonBattleMelodyPlayed = false
	defer setSleepHookForTest(func(time.Duration) {})()
	SetGanonBattleDefeatTrigger(nil)
	SetPlayGanonBattleSongFunc(nil)

	onMelodyRecorded(ganonBattleMelody) // パニックしないことを確認する

	if !GanonBattleMelodyPlayed() {
		t.Fatal("expected GanonBattleMelodyPlayed() to still be set even without a registered trigger")
	}
}

func TestPlayHorseJumpSound_CallsRegisteredHook(t *testing.T) {
	calls := 0
	SetPlayHorseJumpSoundFunc(func() { calls++ })
	defer SetPlayHorseJumpSoundFunc(nil)

	PlayHorseJumpSound()

	if calls != 1 {
		t.Fatalf("expected the registered hook to be called once, got %d", calls)
	}
}

func TestPlayHorseJumpSound_NilHookDoesNothing(t *testing.T) {
	SetPlayHorseJumpSoundFunc(nil)

	PlayHorseJumpSound() // パニックしないことを確認する
}

func TestPlayGanonHallCollapseSound_CallsRegisteredHook(t *testing.T) {
	calls := 0
	SetPlayGanonHallCollapseSoundFunc(func() { calls++ })
	defer SetPlayGanonHallCollapseSoundFunc(nil)

	PlayGanonHallCollapseSound()

	if calls != 1 {
		t.Fatalf("expected the registered hook to be called once, got %d", calls)
	}
}

func TestPlayGanonHallCollapseSound_NilHookDoesNothing(t *testing.T) {
	SetPlayGanonHallCollapseSoundFunc(nil)

	PlayGanonHallCollapseSound() // パニックしないことを確認する
}

func TestOnMelodyRecorded_HorseSongTriggersConfirmationAndContinuation(t *testing.T) {
	horseSongPlayed = false
	defer setSleepHookForTest(func(time.Duration) {})()

	wantTotal := len(music.HorseSongContinuation)

	done := make(chan struct{})
	var mu sync.Mutex
	var played []int
	SetPlayConfirmationFanfareFunc(func() {})
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
	defer func() { SetPlayConfirmationFanfareFunc(nil); SetPlayNoteFunc(nil); SetStopNoteFunc(nil) }()

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
}

func TestOnPitchDetected_LowNoteTriggersJump(t *testing.T) {
	resetJumpGesture()
	triggered := 0
	SetJumpTrigger(func() { triggered++ })
	defer SetJumpTrigger(nil)

	OnPitchDetected(100) // 低い音の立ち上がり → 即座に発火するはず
	if triggered != 1 {
		t.Fatalf("expected exactly one trigger after a single low note, got %d", triggered)
	}
}

func TestOnPitchDetected_SustainedLowNoteTriggersOnce(t *testing.T) {
	resetJumpGesture()
	triggered := 0
	SetJumpTrigger(func() { triggered++ })
	defer SetJumpTrigger(nil)

	// 同じ低い音を無音を挟まず連続で読み取っても、立ち上がりの1回だけ発火する。
	for i := 0; i < 5; i++ {
		OnPitchDetected(100)
	}
	if triggered != 1 {
		t.Fatalf("expected sustained low pitch readings to trigger only once, got %d triggers", triggered)
	}
}

func TestOnPitchDetected_LowNoteAfterSilenceTriggersAgain(t *testing.T) {
	resetJumpGesture()
	triggered := 0
	SetJumpTrigger(func() { triggered++ })
	defer SetJumpTrigger(nil)

	OnPitchDetected(100) // 1回目の低い音 → 発火
	OnPitchDetected(0)   // 無音
	OnPitchDetected(120) // 2回目の低い音(新たな立ち上がり) → 再び発火するはず
	if triggered != 2 {
		t.Fatalf("expected a new low note after silence to trigger again, got %d triggers", triggered)
	}
}

func TestOnPitchDetected_HighNoteDoesNotTriggerJump(t *testing.T) {
	resetJumpGesture()
	triggered := 0
	SetJumpTrigger(func() { triggered++ })
	defer SetJumpTrigger(nil)

	OnPitchDetected(440) // 明確に高い音 → 発火しない
	if triggered != 0 {
		t.Fatalf("expected a high note not to trigger a jump, got %d triggers", triggered)
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

func TestDebugTriggerTitleStart_CallsRegisteredTrigger(t *testing.T) {
	triggered := 0
	SetTitleStartTrigger(func() { triggered++ })
	defer SetTitleStartTrigger(nil)

	DebugTriggerTitleStart()

	if triggered != 1 {
		t.Fatalf("expected the registered title start trigger to be called once, got %d", triggered)
	}
}

func TestDebugTriggerTitleStart_NilTriggerDoesNothing(t *testing.T) {
	SetTitleStartTrigger(nil)

	DebugTriggerTitleStart() // パニックしないことを確認する
}

func TestDebugTriggerJump_CallsRegisteredTrigger(t *testing.T) {
	triggered := 0
	SetJumpTrigger(func() { triggered++ })
	defer SetJumpTrigger(nil)

	DebugTriggerJump()

	if triggered != 1 {
		t.Fatalf("expected the registered jump trigger to be called once, got %d", triggered)
	}
}

func TestDebugTriggerJump_NilTriggerDoesNothing(t *testing.T) {
	SetJumpTrigger(nil)

	DebugTriggerJump() // パニックしないことを確認する
}

func TestDebugTriggerGanonHallMelody_PlaysConfirmationThenTriggersCollapse(t *testing.T) {
	ganonHallMelodyPlayed = false
	defer setSleepHookForTest(func(time.Duration) {})()

	fanfareCalls := 0
	done := make(chan struct{})
	SetPlayConfirmationFanfareFunc(func() { fanfareCalls++ })
	SetGanonHallCollapseTrigger(func() { close(done) })
	defer func() {
		SetPlayConfirmationFanfareFunc(nil)
		SetGanonHallCollapseTrigger(nil)
	}()

	DebugTriggerGanonHallMelody()

	if !GanonHallMelodyPlayed() {
		t.Fatal("expected GanonHallMelodyPlayed() to be true immediately")
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ganon hall collapse trigger was not called within timeout")
	}

	if fanfareCalls != 1 {
		t.Fatalf("expected the confirmation fanfare to be played once before the collapse, got %d calls", fanfareCalls)
	}
}
