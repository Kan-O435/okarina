package music

import "testing"

func TestPitchFromMIDINote(t *testing.T) {
	cases := map[int]Pitch{
		60: C,
		62: D,
		64: E,
		67: G,
		72: C, // オクターブ違いでも同じ音名
	}
	for note, want := range cases {
		if got := PitchFromMIDINote(note); got != want {
			t.Errorf("PitchFromMIDINote(%d) = %s, want %s", note, got, want)
		}
	}
}

func TestRecognize(t *testing.T) {
	melody := Melody{
		{Pitch: C}, {Pitch: D}, {Pitch: E}, {Pitch: G},
	}
	if name := Recognize(melody, DefaultPatterns); name != "SUN_MELODY" {
		t.Errorf("Recognize() = %q, want SUN_MELODY", name)
	}

	unknown := Melody{{Pitch: C}, {Pitch: C}, {Pitch: C}}
	if name := Recognize(unknown, DefaultPatterns); name != "" {
		t.Errorf("Recognize() = %q, want no match", name)
	}
}

func TestRecognize_SongOfTime(t *testing.T) {
	// 時の歌: ラ→レ→ファ→ラ→レ→ファ
	melody := Melody{
		{Pitch: A}, {Pitch: D}, {Pitch: F},
		{Pitch: A}, {Pitch: D}, {Pitch: F},
	}
	if name := Recognize(melody, DefaultPatterns); name != SongOfTimeName {
		t.Errorf("Recognize() = %q, want %s", name, SongOfTimeName)
	}

	wrongOrder := Melody{
		{Pitch: D}, {Pitch: A}, {Pitch: F},
		{Pitch: A}, {Pitch: D}, {Pitch: F},
	}
	if name := Recognize(wrongOrder, DefaultPatterns); name != "" {
		t.Errorf("Recognize() = %q, want no match for wrong order", name)
	}
}
