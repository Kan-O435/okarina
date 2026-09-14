// Package game は、ゲーム全体のループ・状態管理を担当する。
package game

import (
	"fmt"
	"time"

	"github.com/Kan-O435/okarina/internal/midi"
	"github.com/Kan-O435/okarina/internal/music"
	"github.com/Kan-O435/okarina/internal/player"
	"github.com/Kan-O435/okarina/internal/renderer"
	"github.com/Kan-O435/okarina/internal/vecmath"
)

// melodyIdleTimeout は、この時間MIDI入力がなければ1回の演奏が
// 終わったとみなす無音時間。
const melodyIdleTimeout = 2 * time.Second

var recorder = music.NewRecorder(melodyIdleTimeout, onMelodyRecorded)

// ctx・scene・linkIndexは、InitRenderer()で描画準備が完了した後、
// UpdateFrame()から毎フレームLinkの位置を書き換えて再描画するために保持する。
var (
	ctx       *renderer.Context
	scene     *renderer.Scene
	linkIndex = -1
)

// Run はゲームのエントリーポイント。
// 現時点ではプロジェクトの雛形確認用の最小実装。
func Run() {
	fmt.Println("game.Run() called")
}

// InitRenderer はWebGLコンテキストを初期化し、フィールドのデモシーンを
// 構築して最初の描画を行う。以後はUpdateFrame()がプレイヤーの移動に
// 合わせて再描画する。
func InitRenderer(canvasID string) error {
	c, err := renderer.NewContext(canvasID)
	if err != nil {
		return fmt.Errorf("renderer context init failed: %w", err)
	}

	width, height := c.CanvasSize()
	c.Viewport(width, height)
	c.EnableDepthTest()
	c.ClearColor(0.53, 0.75, 0.9, 1.0) // 空っぽい水色

	s, li, err := renderer.BuildFieldDemoScene(c)
	if err != nil {
		return fmt.Errorf("failed to build demo scene: %w", err)
	}

	ctx = c
	scene = s
	linkIndex = li
	scene.Render(ctx)
	return nil
}

// OnPitchDetected はマイクから検出された最新のピッチ(Hz)をプレイヤーの
// 移動方向判定に渡す。ピッチが検出できなかった場合はfreqに0以下を渡す。
func OnPitchDetected(freq float64) {
	player.Player.OnPitch(freq)
}

// PlayerDirection は現在のプレイヤーの移動方向を文字列で返す
// ("forward" | "backward" | "idle")。UI表示など、JS側からの参照用。
func PlayerDirection() string {
	return player.Player.Direction.String()
}

// UpdateFrame は毎フレーム呼び出され、プレイヤーを現在の移動方向に
// 応じて進め、Linkの位置を更新して再描画する。dtは前フレームからの
// 経過時間(秒)。
func UpdateFrame(dt float64) {
	if scene == nil || linkIndex < 0 {
		return
	}

	deltaZ := player.Player.Update(dt)
	if deltaZ == 0 {
		return
	}

	move := vecmath.Translate(vecmath.NewVec3(0, 0, deltaZ))
	scene.Objects[linkIndex].Transform = move.Mul(scene.Objects[linkIndex].Transform)
	scene.Render(ctx)
}

// OnMIDIEvent はJS-Go Bridge経由で受け取ったMIDIイベントを
// 旋律記録エンジン(internal/music.Recorder)に渡す。
func OnMIDIEvent(e midi.Event) {
	recorder.HandleEvent(e)
}

// onMelodyRecorded は一連の演奏が確定した際に呼ばれ、
// 登録済みの旋律パターンと照合する。
func onMelodyRecorded(melody music.Melody) {
	fmt.Printf("[music] melody recorded: %v\n", melody.Pitches())

	if name := music.Recognize(melody, music.DefaultPatterns); name != "" {
		fmt.Printf("[music] recognized: %s\n", name)
	} else {
		fmt.Println("[music] recognized: (no match)")
	}
}
