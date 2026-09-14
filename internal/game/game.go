// Package game は、ゲーム全体のループ・状態管理を担当する。
package game

import (
	"fmt"
	"sync"
	"time"

	"github.com/Kan-O435/okarina/internal/midi"
	"github.com/Kan-O435/okarina/internal/music"
	"github.com/Kan-O435/okarina/internal/player"
	"github.com/Kan-O435/okarina/internal/renderer"
	"github.com/Kan-O435/okarina/internal/vecmath"
	"github.com/Kan-O435/okarina/internal/world"
)

// melodyIdleTimeout は、この時間MIDI入力がなければ1回の演奏が
// 終わったとみなす無音時間。
const melodyIdleTimeout = 2 * time.Second

var recorder = music.NewRecorder(melodyIdleTimeout, onMelodyRecorded)

// Run はゲームのエントリーポイント。
// 現時点ではプロジェクトの雛形確認用の最小実装。
func Run() {
	fmt.Println("game.Run() called")
}

// Game は、Sceneと、毎フレーム更新が必要な状態(プレイヤー=Linkの位置・
// 向き・隠し扉の開閉)をまとめて保持し、Update()でSceneのTransformへ反映する。
type Game struct {
	scene     *renderer.Scene
	link      renderer.LinkPlacement
	doorIndex int
	door      world.Door
}

// New はSceneと、Link/扉Objectの配置情報(BuildFieldDemoSceneが返す)から
// Gameを組み立てる。プレイヤーの初期位置をLinkのスポーン地点に合わせる。
func New(scene *renderer.Scene, link renderer.LinkPlacement, doorIndex int) *Game {
	player.Player.Z = link.SpawnZ
	return &Game{scene: scene, link: link, doorIndex: doorIndex}
}

// OpenDoor は隠し扉を開き始める。何らかの「特定の動作」(将来的にはMIDIの
// メロディ認識、今は動作確認用のボタン)から呼ばれる想定。
func (g *Game) OpenDoor() {
	g.door.Open()
}

// Update はdeltaTime(秒)だけゲーム状態を進め、扉・プレイヤー(Link)の
// 見た目(Transform)に反映する。プレイヤーは前進/後退に応じて位置だけで
// なく向き(Yaw)も変わるため、毎フレームTransformを一から組み立て直す。
func (g *Game) Update(dt float64) {
	g.door.Update(dt)
	g.scene.Objects[g.doorIndex].Transform = renderer.DoorTransform(g.door.Progress)

	deltaZ := player.Player.Update(dt)
	if deltaZ != 0 {
		worldPos := vecmath.Translate(vecmath.NewVec3(0, 0, player.Player.Z))
		facing := vecmath.RotateY(player.Player.Yaw)
		g.scene.Objects[g.link.Index].Transform = worldPos.Mul(facing).Mul(g.link.LocalTransform)
	}
}

// instance は、ブラウザ側(bridge)からの呼び出しを受けるための
// パッケージレベルのシングルトン。
var instance *Game

// SetInstance はブリッジ経由の呼び出し先となるGameインスタンスを登録する。
func SetInstance(g *Game) {
	instance = g
}

// OpenDoor はブリッジ(JavaScript側)から呼ばれ、登録済みのGameインスタンスの
// 扉を開く。インスタンスが未登録の場合は何もしない。
func OpenDoor() {
	if instance != nil {
		instance.OpenDoor()
	}
}

// OnPitchDetected はマイクから検出された最新のピッチ(Hz)をプレイヤーの
// 移動方向判定に渡す。ピッチが検出できなかった場合はfreqに0以下を渡す。
// 実際のプレイヤー移動は、Go側で常時回っているゲームループ
// (Context.RunLoop、cmd/game/main.go参照)がGame.Update()経由で進める。
func OnPitchDetected(freq float64) {
	player.Player.OnPitch(freq)
}

// PlayerDirection は現在のプレイヤーの移動方向を文字列で返す
// ("forward" | "backward" | "idle")。UI表示など、JS側からの参照用。
func PlayerDirection() string {
	return player.Player.Direction.String()
}

// SetDebugDirection は、オタマトーンのピッチ入力を介さずにプレイヤーの
// 移動方向を直接指定するデバッグ用エントリーポイント(矢印キー操作など)。
// directionは"forward" | "backward" | "idle"のいずれか。
func SetDebugDirection(direction string) {
	var d player.Direction
	switch direction {
	case "forward":
		d = player.Forward
	case "backward":
		d = player.Backward
	default:
		d = player.Idle
	}
	player.Player.SetDirection(d)
}

// OnMIDIEvent はJS-Go Bridge経由で受け取ったMIDIイベントを
// 旋律記録エンジン(internal/music.Recorder)に渡す。
func OnMIDIEvent(e midi.Event) {
	recorder.HandleEvent(e)
}

// songOfTimePlayed は、時の歌(music.SongOfTimeName)が正しく演奏された
// ことを示す。
var songOfTimePlayed bool

// SongOfTimePlayed は、時の歌が正しく演奏されたかどうかを返す。
func SongOfTimePlayed() bool {
	return songOfTimePlayed
}

// audioHooksMu は、下のplayNoteHook・stopNoteHook・sleepHookへの読み書きを
// 保護する。onMelodyRecordedはgoroutineを起動して非同期に曲の続きを再生する
// ため、bridge.Init()での差し込みやテストでの差し替えと同時に読まれても
// 安全なようにしている。
var audioHooksMu sync.Mutex

// playNoteHook・stopNoteHook は、Goから直接ブラウザの音声再生(Web Audio
// API、web/audio.jsのplayNote/stopNote)を呼び出すためのフック。
// bridge.Init()がJS側の実装を差し込む。ネイティブビルドやJS未初期化時は
// nilのまま。sleepHookはtime.Sleepの差し替え用(テストで待ち時間を省略する)。
var (
	playNoteHook func(note, velocity int)
	stopNoteHook func(note int)
	sleepHook    = time.Sleep
)

// SetPlayNoteFunc は、Goから曲を自動再生する際に使う「1音鳴らす」実装を
// 登録する(bridge.Init()から呼ばれる)。
func SetPlayNoteFunc(f func(note, velocity int)) {
	audioHooksMu.Lock()
	playNoteHook = f
	audioHooksMu.Unlock()
}

// SetStopNoteFunc は、Goから曲を自動再生する際に使う「1音止める」実装を
// 登録する(bridge.Init()から呼ばれる)。
func SetStopNoteFunc(f func(note int)) {
	audioHooksMu.Lock()
	stopNoteHook = f
	audioHooksMu.Unlock()
}

// onMelodyRecorded は一連の演奏が確定した際に呼ばれ、
// 登録済みの旋律パターンと照合する。時の歌が演奏された場合は隠し扉を開き、
// 本家のゼルダのように確認音(SongOfTimeConfirmation)→曲の続き
// (SongOfTimeContinuation)の順で自動再生する。
func onMelodyRecorded(melody music.Melody) {
	fmt.Printf("[music] melody recorded: %v\n", melody.Pitches())

	name := music.Recognize(melody, music.DefaultPatterns)
	if name == "" {
		fmt.Println("[music] recognized: (no match)")
		return
	}

	fmt.Printf("[music] recognized: %s\n", name)
	if name == music.SongOfTimeName {
		songOfTimePlayed = true
		fmt.Println("[music] 時の歌が演奏されました。隠し扉が開きます。")
		OpenDoor()
		go playSongOfTimeAudio()
	}
}

// playSongOfTimeAudio は、プレイヤーが演奏した合図に続けて、確認音
// (music.SongOfTimeConfirmation)を鳴らした後、本家のゼルダのように
// 曲を最初から(music.SongOfTimeOpening→music.SongOfTimeContinuation)
// 通して自動再生する。再生用フックが未登録(ネイティブビルドやJS未初期化時)
// の場合は何もしない。
func playSongOfTimeAudio() {
	audioHooksMu.Lock()
	play, stop, sleep := playNoteHook, stopNoteHook, sleepHook
	audioHooksMu.Unlock()

	if play == nil || stop == nil {
		return
	}
	playNotes(play, stop, sleep, music.SongOfTimeConfirmation)
	sleep(music.SongOfTimePauseDur)
	playNotes(play, stop, sleep, music.SongOfTimeOpening)
	playNotes(play, stop, sleep, music.SongOfTimeContinuation)
}

// playNotes はnotesを順番に、1音ずつ鳴らして止めてを繰り返しながら再生する。
func playNotes(play func(note, velocity int), stop func(note int), sleep func(time.Duration), notes []music.ContinuationNote) {
	for _, n := range notes {
		play(n.MIDINote, 100)
		sleep(n.Duration)
		stop(n.MIDINote)
	}
}
