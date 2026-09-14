package music

import "time"

// Pattern はゲーム内で意味を持つ、登録済みの旋律を表す。
type Pattern struct {
	Name    string
	Pitches []Pitch
}

// SongOfTimeName は「時の歌」のPattern名。プレイヤーがこの合図
// (ラ→レ→ファ→ラ→レ→ファ)を演奏すると隠し扉が開き、続けて
// SongOfTimeContinuation(曲の後半部分)が自動再生される。
const SongOfTimeName = "SONG_OF_TIME"

// DefaultPatterns はゲームで使用する魔法の旋律の定義一覧(MVP)。
var DefaultPatterns = []Pattern{
	{Name: "SUN_MELODY", Pitches: []Pitch{C, D, E, G}},
	{Name: "RAIN_MELODY", Pitches: []Pitch{G, E, D, C}},
	{Name: "DOOR_MELODY", Pitches: []Pitch{C, G, C}},
	// 時の歌(前半・演奏で認識させる合図): ラ→レ→ファ→ラ→レ→ファ
	{Name: SongOfTimeName, Pitches: []Pitch{A, D, F, A, D, F}},
}

// ContinuationNote は、自動再生する曲の1音を表す。演奏判定に使うPitch
// (オクターブ非区別)とは異なり、実際に鳴らすための具体的なMIDIノート
// 番号と、鳴らす長さを持つ。
type ContinuationNote struct {
	MIDINote int
	Duration time.Duration
}

const (
	continuationNormalDur = 350 * time.Millisecond
	continuationLongDur   = 700 * time.Millisecond // 「ー」で伸ばす音
)

// SongOfTimeContinuation は、プレイヤーがSongOfTimeName(前半6音)を正しく
// 演奏した後、本家のゼルダのように続けて自動再生される「時の歌」の後半部分。
// ラ₄→ド₅→シ₄ー→ソ₄ー→ファ₄→ソ₄→ラ₄ー｜レ₅→ド₅→ミ₅→レ₅ー
var SongOfTimeContinuation = []ContinuationNote{
	{MIDINote: 69, Duration: continuationNormalDur}, // ラ₄
	{MIDINote: 72, Duration: continuationNormalDur}, // ド₅
	{MIDINote: 71, Duration: continuationLongDur},   // シ₄ー
	{MIDINote: 67, Duration: continuationLongDur},   // ソ₄ー
	{MIDINote: 65, Duration: continuationNormalDur}, // ファ₄
	{MIDINote: 67, Duration: continuationNormalDur}, // ソ₄
	{MIDINote: 69, Duration: continuationLongDur},   // ラ₄ー
	{MIDINote: 74, Duration: continuationNormalDur}, // レ₅
	{MIDINote: 72, Duration: continuationNormalDur}, // ド₅
	{MIDINote: 76, Duration: continuationNormalDur}, // ミ₅
	{MIDINote: 74, Duration: continuationLongDur},   // レ₅ー
}

// Recognize は演奏されたMelodyが登録済みPatternのいずれかと完全一致するか判定する。
// 一致すればそのPattern名を返す。一致しなければ空文字を返す。
// MVPでは完全一致のみをサポートする。
func Recognize(melody Melody, patterns []Pattern) string {
	played := melody.Pitches()
	for _, p := range patterns {
		if equalPitches(played, p.Pitches) {
			return p.Name
		}
	}
	return ""
}

func equalPitches(a, b []Pitch) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
