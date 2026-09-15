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

// jumpDuration/jumpHeight は、StartJump()で始まるジャンプの弧の長さと高さ。
// アニメーション・スキニングが未実装のため、Transform()に加える見た目上の
// Y方向オフセットとして表現する(実際のメッシュは変形しない)。
const (
	jumpDuration = 0.6 // 秒
	jumpHeight   = 1.5 // ワールド単位(最高点)
)

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

	// SpeedMultiplier は、moveSpeedに掛け合わせる倍率。デフォルトは1.0
	// (等速)で、馬に乗っている間だけ大きくする、といった用途に使う
	// (0の場合も1.0として扱う。ゼロ値のStateをそのまま使えるようにするため)。
	SpeedMultiplier float64

	lastSemitone float64
	hasPitch     bool

	jumping     bool
	jumpElapsed float64

	// lastMoveDirection は、直前にDirectionがForward/Backwardになった
	// ときの値を保持する。ジャンプ中にDirectionがIdleになっても
	// (ジャンプのジェスチャー自体は前後移動を止めるため)、ジャンプ中は
	// この向きへ進み続けるために使う(Update参照)。
	lastMoveDirection Direction
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

// effectiveSpeed はmoveSpeedにSpeedMultiplierを適用した実際の移動速度を
// 返す。SpeedMultiplierが未設定(ゼロ値)の場合は1.0(等速)として扱う。
func (s *State) effectiveSpeed() float64 {
	if s.SpeedMultiplier == 0 {
		return moveSpeed
	}
	return moveSpeed * s.SpeedMultiplier
}

// StartJump はジャンプの弧を開始する。既にジャンプ中の場合は何もしない
// (二重にジャンプが始まらないようにするため)。
func (s *State) StartJump() {
	if s.jumping {
		return
	}
	s.jumping = true
	s.jumpElapsed = 0
}

// IsJumping は、現在ジャンプの弧の途中(空中)かどうかを返す。障害物の
// 当たり判定などで、「ジャンプ中は通り抜けられる」といった判定に使う。
func (s *State) IsJumping() bool {
	return s.jumping
}

// jumpOffsetY は、ジャンプ中のY方向オフセット(サインカーブで放物線状の
// 弧を描く)を返す。ジャンプ中でなければ0。
func (s *State) jumpOffsetY() float64 {
	if !s.jumping {
		return 0
	}
	t := s.jumpElapsed / jumpDuration
	return jumpHeight * math.Sin(math.Pi*t)
}

// Update は経過時間dt(秒)に応じてプレイヤーの位置・向き・ジャンプの進み
// 具合を進め、このフレームで移動したZ方向の量(ワールド単位)を返す。
// Idle中は位置も向きも変えず、直前に移動していた方向を向いたままにする
// (ジャンプ中かどうかに関わらず、これは変わらない)。
//
// ただし、ジャンプ中(IsJumping)にDirectionがIdleの場合は例外で、
// 直前に前後移動していた方向(lastMoveDirection)へ進み続ける。ジャンプの
// ジェスチャー自体(低い音を鳴らす等)がDirectionをIdleにしてしまうため、
// これが無いとジャンプ中に足が止まって見えてしまう。
func (s *State) Update(dt float64) float64 {
	if s.jumping {
		s.jumpElapsed += dt
		if s.jumpElapsed >= jumpDuration {
			s.jumping = false
			s.jumpElapsed = 0
		}
	}

	if s.Direction == Forward || s.Direction == Backward {
		s.lastMoveDirection = s.Direction
	}

	direction := s.Direction
	if s.jumping && direction == Idle {
		direction = s.lastMoveDirection
	}

	var deltaZ float64
	speed := s.effectiveSpeed()
	switch direction {
	case Forward:
		s.Yaw = math.Pi
		deltaZ = -speed * dt
	case Backward:
		s.Yaw = 0
		deltaZ = speed * dt
	default:
		return 0
	}
	s.Z += deltaZ
	return deltaZ
}

// Transform は、localTransform(Linkモデル自身の原点補正・スケール)に、
// 現在のワールド座標(Z)・向き(Yaw)・ジャンプ中のY方向オフセットを適用した
// 最終的な配置行列を返す。フィールド(神殿・草原等)を問わず、Linkの見た目を
// 毎フレーム組み立て直す際に共通で使う。
func (s *State) Transform(localTransform vecmath.Mat4) vecmath.Mat4 {
	worldPos := vecmath.Translate(vecmath.NewVec3(0, s.jumpOffsetY(), s.Z))
	facing := vecmath.RotateY(s.Yaw)
	return worldPos.Mul(facing).Mul(localTransform)
}
