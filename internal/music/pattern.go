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

// HorseSongName は「馬の歌」のPattern名。草原フィールドでプレイヤーが
// この合図(レ→シ→ラ→レ→シ→ラ)を演奏すると馬が呼び出される。
const HorseSongName = "HORSE_SONG"

// DefaultPatterns はゲームで使用する魔法の旋律の定義一覧(MVP)。
var DefaultPatterns = []Pattern{
	{Name: "SUN_MELODY", Pitches: []Pitch{C, D, E, G}},
	{Name: "RAIN_MELODY", Pitches: []Pitch{G, E, D, C}},
	{Name: "DOOR_MELODY", Pitches: []Pitch{C, G, C}},
	// 時の歌(前半・演奏で認識させる合図): ラ→レ→ファ→ラ→レ→ファ
	{Name: SongOfTimeName, Pitches: []Pitch{A, D, F, A, D, F}},
	// 馬の歌(演奏で認識させる合図): レ→シ→ラ→レ→シ→ラ
	{Name: HorseSongName, Pitches: []Pitch{D, B, A, D, B, A}},
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
	continuationLongDur   = 700 * time.Millisecond    // 「ー」1個で伸ばす音(2拍分)
	continuationLongDur2  = 3 * continuationNormalDur // 「ーー」2個で伸ばす音(3拍分)
	continuationLongDur3  = 4 * continuationNormalDur // 「ーーー」3個で伸ばす音(4拍分)
	confirmationNoteDur   = 120 * time.Millisecond    // 「テレレレレ」の確認音は短く刻む

	// SongOfTimePauseDur は、確認音(SongOfTimeConfirmation)が鳴り終わって
	// から曲の続き(SongOfTimeContinuation等)が始まるまでの間(1拍分)。
	SongOfTimePauseDur = continuationNormalDur
)

// SongOfTimeConfirmation は、プレイヤーがSongOfTimeName(前半6音)を正しく
// 演奏した直後に鳴る、本家のゼルダのような短い確認音(「テレレレレ」)。
// オクターブ指定が無いため、続きのメロディ(SongOfTimeContinuation)より
// 1オクターブ高い6オクターブで鳴らしている。
// ソ₆→ファ#₆→レ#₆→ラ₆→ソ#₆→ミ₆→ソ#₆→ド₆
var SongOfTimeConfirmation = []ContinuationNote{
	{MIDINote: 91, Duration: confirmationNoteDur}, // ソ₆
	{MIDINote: 90, Duration: confirmationNoteDur}, // ファ#₆
	{MIDINote: 87, Duration: confirmationNoteDur}, // レ#₆
	{MIDINote: 93, Duration: confirmationNoteDur}, // ラ₆
	{MIDINote: 92, Duration: confirmationNoteDur}, // ソ#₆
	{MIDINote: 88, Duration: confirmationNoteDur}, // ミ₆
	{MIDINote: 92, Duration: confirmationNoteDur}, // ソ#₆
	{MIDINote: 84, Duration: confirmationNoteDur}, // ド₆
}

// SongOfTimeOpening は、確認音(SongOfTimeConfirmation)の後に自動再生する
// 「時の歌」の冒頭部分(=SongOfTimeNameのPitchesと同じ音: ラ→レ→ファ→ラ→
// レ→ファ)を、実際に鳴らすためのMIDIノート番号・長さ付きで表したもの。
// 本家のゼルダは確認音の後に曲を最初から(プレイヤーが演奏した部分も
// 含めて)通して流すため、SongOfTimeContinuationの前にこれを鳴らす。
// ラ₅ー→レ₅ー→ファ₅ー→ラ₅ー→レ₅ー→ファ₅ー
var SongOfTimeOpening = []ContinuationNote{
	{MIDINote: 81, Duration: continuationLongDur}, // ラ₅ー
	{MIDINote: 74, Duration: continuationLongDur}, // レ₅ー
	{MIDINote: 77, Duration: continuationLongDur}, // ファ₅ー
	{MIDINote: 81, Duration: continuationLongDur}, // ラ₅ー
	{MIDINote: 74, Duration: continuationLongDur}, // レ₅ー
	{MIDINote: 77, Duration: continuationLongDur}, // ファ₅ー
}

// SongOfTimeContinuation は、SongOfTimeOpeningに続けて自動再生される
// 「時の歌」の後半部分。
// ラ₅→ド₆→シ₅ー→ソ₅ー→ファ₅→ソ₅→ラ₅ー｜レ₆→ド₆→ミ₆→レ₆ー
var SongOfTimeContinuation = []ContinuationNote{
	{MIDINote: 81, Duration: continuationNormalDur}, // ラ₅
	{MIDINote: 84, Duration: continuationNormalDur}, // ド₆
	{MIDINote: 83, Duration: continuationLongDur},   // シ₅ー
	{MIDINote: 79, Duration: continuationLongDur},   // ソ₅ー
	{MIDINote: 77, Duration: continuationNormalDur}, // ファ₅
	{MIDINote: 79, Duration: continuationNormalDur}, // ソ₅
	{MIDINote: 81, Duration: continuationLongDur},   // ラ₅ー
	{MIDINote: 86, Duration: continuationNormalDur}, // レ₆
	{MIDINote: 84, Duration: continuationNormalDur}, // ド₆
	{MIDINote: 88, Duration: continuationNormalDur}, // ミ₆
	{MIDINote: 86, Duration: continuationLongDur},   // レ₆ー
}

// HorseSongContinuation は、プレイヤーがHorseSongName(レ→シ→ラ→レ→シ→ラ)を
// 正しく演奏した後、確認音(SongOfTimeConfirmationを共用)に続けて
// 自動再生される「馬の歌」の続き。
// レ→シ→ラーーー→レ→シ→ラーー→レ→シ→ラー→レ→ラーー
var HorseSongContinuation = []ContinuationNote{
	{MIDINote: 74, Duration: continuationNormalDur}, // レ₅
	{MIDINote: 71, Duration: continuationNormalDur}, // シ₄
	{MIDINote: 69, Duration: continuationLongDur3},  // ラ₄ーーー
	{MIDINote: 74, Duration: continuationNormalDur}, // レ₅
	{MIDINote: 71, Duration: continuationNormalDur}, // シ₄
	{MIDINote: 69, Duration: continuationLongDur2},  // ラ₄ーー
	{MIDINote: 74, Duration: continuationNormalDur}, // レ₅
	{MIDINote: 71, Duration: continuationNormalDur}, // シ₄
	{MIDINote: 69, Duration: continuationLongDur},   // ラ₄ー
	{MIDINote: 74, Duration: continuationNormalDur}, // レ₅
	{MIDINote: 69, Duration: continuationLongDur2},  // ラ₄ーー
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
