// Package music は、認識対象のメロディ(合言葉曲)の定義と、
// 演奏内容とメロディとの照合ロジックを担当する。
package music

import "time"

// Pitch は音名を表す(オクターブは区別しない)。
type Pitch string

const (
	C  Pitch = "C"
	Cs Pitch = "C#"
	D  Pitch = "D"
	Ds Pitch = "D#"
	E  Pitch = "E"
	F  Pitch = "F"
	Fs Pitch = "F#"
	G  Pitch = "G"
	Gs Pitch = "G#"
	A  Pitch = "A"
	As Pitch = "A#"
	B  Pitch = "B"
)

// pitchNames はMIDIノート番号を12音階のインデックスで音名に変換するための表。
var pitchNames = [12]Pitch{C, Cs, D, Ds, E, F, Fs, G, Gs, A, As, B}

// PitchFromMIDINote はMIDIノート番号(0-127)を音名に変換する(オクターブは無視)。
func PitchFromMIDINote(note int) Pitch {
	return pitchNames[((note%12)+12)%12]
}

// Note はゲーム内部で扱う1つの音符を表す。
// MIDIのNote ON(開始)とNote OFF(終了)のペアから生成される。
type Note struct {
	Pitch     Pitch
	Velocity  int
	StartTime time.Time
	EndTime   time.Time
}

// Duration はNoteの継続時間(Note ONからNote OFFまでの長さ)を返す。
func (n Note) Duration() time.Duration {
	return n.EndTime.Sub(n.StartTime)
}

// Melody は演奏された(または登録された)一連の音符を表す。
type Melody []Note

// Pitches はMelodyに含まれる音名だけを演奏順に並べたものを返す。
func (m Melody) Pitches() []Pitch {
	pitches := make([]Pitch, len(m))
	for i, n := range m {
		pitches[i] = n.Pitch
	}
	return pitches
}
