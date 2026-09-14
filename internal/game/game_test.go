package game

import (
	"testing"

	"github.com/Kan-O435/okarina/internal/music"
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

func TestOnMelodyRecorded_SongOfTimeOpensDoor(t *testing.T) {
	songOfTimePlayed = false
	g := &Game{
		scene:     &renderer.Scene{Objects: []renderer.Object{{}}},
		doorIndex: 0,
	}
	SetInstance(g)
	defer SetInstance(nil)

	onMelodyRecorded(songOfTime)

	if g.door.State != world.DoorOpening {
		t.Fatalf("expected door to start opening, got state=%v", g.door.State)
	}
}
