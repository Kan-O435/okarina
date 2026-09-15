// Package player は、オタマトーンの音程変化に応じて動くプレイヤー(リンク)の
// 位置・移動方向を管理する。
//
// 低い音から高い音へ変化させると前進、高い音から低い音へ変化させると後退する。
// Hz(周波数)そのままでは低音域と高音域で1音の変化幅が大きく異なるため、
// 半音(セミトーン)単位に変換してから比較する。
package player

import (
	"math"

	"github.com/Kan-O435/okarina/internal/vecmath"
)

// moveSpeed はワールド単位/秒での移動速度。
const moveSpeed = 8.0

// semitoneDeadZone は「音程が変化した」と判定するための最小変化量(半音)。
// ピッチ検出のわずかなブレをノイズとして無視するための遊び。
const semitoneDeadZone = 0.5

// Direction はプレイヤーの現在の移動方向。
type Direction int

const (
	Idle Direction = iota
	Forward
	Backward
)

// String はDirectionをJS側に渡しやすい文字列表現に変換する。
func (d Direction) String() string {
	switch d {
	case Forward:
		return "forward"
	case Backward:
		return "backward"
	default:
		return "idle"
	}
}

// State はプレイヤー(リンク)の現在位置・移動方向・向きを保持する。
type State struct {
	Z         float64
	Direction Direction
	Yaw       float64 // Y軸周りの向き(ラジアン)。進行方向に合わせて変わる。

	lastSemitone float64
	hasPitch     bool
}

// Player はゲーム全体で共有するプレイヤー状態。
var Player = &State{}

// frequencyToSemitone は周波数(Hz)を、A4(440Hz)を基準にした連続的な
// 半音値に変換する(オクターブをまたいでも大小関係が保たれる)。
func frequencyToSemitone(freq float64) float64 {
	return 12 * math.Log2(freq/440)
}

// SetDirection はプレイヤーの移動方向を直接指定する。オタマトーンの
// ピッチ入力を介さないデバッグ操作(矢印キー等)から使う。
func (s *State) SetDirection(d Direction) {
	s.Direction = d
}

// OnPitch はマイクから検出された最新のピッチ(Hz)を受け取り、直前のピッチ
// との比較から移動方向を更新する。freqが0以下の場合は「音が検出できな
// かった(無音・音量不足)」として移動を止め、次に音が検出された時点の
// ピッチを新たな基準にする。
func (s *State) OnPitch(freq float64) {
	if freq <= 0 {
		s.Direction = Idle
		s.hasPitch = false
		return
	}

	semitone := frequencyToSemitone(freq)
	if !s.hasPitch {
		s.lastSemitone = semitone
		s.hasPitch = true
		s.Direction = Idle
		return
	}

	delta := semitone - s.lastSemitone
	switch {
	case delta > semitoneDeadZone:
		s.Direction = Forward
	case delta < -semitoneDeadZone:
		s.Direction = Backward
	default:
		s.Direction = Idle
	}
	s.lastSemitone = semitone
}

// Update は経過時間dt(秒)に応じてプレイヤーの位置・向きを進め、この
// フレームで移動したZ方向の量(ワールド単位)を返す。Idle中は位置も
// 向きも変えず、直前に移動していた方向を向いたままにする。
func (s *State) Update(dt float64) float64 {
	var deltaZ float64
	switch s.Direction {
	case Forward:
		s.Yaw = math.Pi
		deltaZ = -moveSpeed * dt
	case Backward:
		s.Yaw = 0
		deltaZ = moveSpeed * dt
	default:
		return 0
	}
	s.Z += deltaZ
	return deltaZ
}

// Transform は、localTransform(Linkモデル自身の原点補正・スケール)に、
// 現在のワールド座標(Z)・向き(Yaw)を適用した最終的な配置行列を返す。
// フィールド(神殿・草原等)を問わず、Linkの見た目を毎フレーム組み立て直す
// 際に共通で使う。
func (s *State) Transform(localTransform vecmath.Mat4) vecmath.Mat4 {
	worldPos := vecmath.Translate(vecmath.NewVec3(0, 0, s.Z))
	facing := vecmath.RotateY(s.Yaw)
	return worldPos.Mul(facing).Mul(localTransform)
}
