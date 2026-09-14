package game

import (
	"testing"

	"github.com/Kan-O435/okarina/internal/music"
)

func TestOnMelodyRecorded_OcarinaSongCorrect(t *testing.T) {
	ocarinaSongPlayed = false
	melody := music.Melody{
		{Pitch: music.A}, {Pitch: music.D}, {Pitch: music.F},
		{Pitch: music.A}, {Pitch: music.D}, {Pitch: music.F},
	}

	onMelodyRecorded(melody)

	if !OcarinaSongPlayed() {
		t.Fatal("expected OcarinaSongPlayed() to be true after playing ラレファラレファ correctly")
	}
}

func TestOnMelodyRecorded_UnrecognizedMelodyDoesNotSetFlag(t *testing.T) {
	ocarinaSongPlayed = false
	melody := music.Melody{{Pitch: music.C}, {Pitch: music.C}}

	onMelodyRecorded(melody)

	if OcarinaSongPlayed() {
		t.Fatal("expected OcarinaSongPlayed() to remain false for an unrecognized melody")
	}
}

func TestOnMelodyRecorded_PartialSongDoesNotSetFlag(t *testing.T) {
	ocarinaSongPlayed = false
	// 前半(ラ→レ→ファ)だけでは不十分。
	melody := music.Melody{{Pitch: music.A}, {Pitch: music.D}, {Pitch: music.F}}

	onMelodyRecorded(melody)

	if OcarinaSongPlayed() {
		t.Fatal("expected OcarinaSongPlayed() to remain false for a partial performance")
	}
}
