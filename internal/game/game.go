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
	"github.com/Kan-O435/okarina/internal/world"
)

// melodyIdleTimeout は、この時間MIDI入力がなければ1回の演奏が
// 終わったとみなす無音時間。
const melodyIdleTimeout = 2 * time.Second

// doorMelodyName は、隠し扉を開くメロディに対応するmusic.Patternの名前
// (internal/music/pattern.goのDefaultPatterns参照)。
const doorMelodyName = "DOOR_MELODY"

// doorOpenDelay は、扉のメロディが認識されてから実際に扉が開き始めるまでの
// 「ため」の時間。varにしているのはテストから短く差し替えられるようにするため。
var doorOpenDelay = 7 * time.Second

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

	// autoWalking は、扉が開いた後にプレイヤー入力を無視してLinkを
	// 自動で前進させ続けている(扉をすり抜ける演出中の)状態かどうか。
	autoWalking bool
	// transitioned は、onFieldTransitionをすでに呼んだかどうか
	// (Zがしきい値を超え続けても二重に呼ばないためのガード)。
	transitioned bool
	// onFieldTransition は、Linkが扉を通過し終えたときに1回だけ呼ばれる
	// コールバック(次のフィールドへのページ遷移などに使う)。
	onFieldTransition func()
}

// New はSceneと、Link/扉Objectの配置情報(BuildFieldDemoSceneが返す)から
// Gameを組み立てる。プレイヤーの初期位置をLinkのスポーン地点に合わせる。
// onFieldTransitionは、扉が開いた後にLinkが自動で前進して扉を通過し終えた
// タイミングで1回だけ呼ばれる(nilなら何も起きない)。
func New(scene *renderer.Scene, link renderer.LinkPlacement, doorIndex int, onFieldTransition func()) *Game {
	player.Player.Z = link.SpawnZ
	return &Game{scene: scene, link: link, doorIndex: doorIndex, onFieldTransition: onFieldTransition}
}

// OpenDoor は隠し扉を開き始める。何らかの「特定の動作」(将来的にはMIDIの
// メロディ認識、今は動作確認用のボタン)から呼ばれる想定。
func (g *Game) OpenDoor() {
	g.door.Open()
}

// Update はdeltaTime(秒)だけゲーム状態を進め、扉・プレイヤー(Link)の
// 見た目(Transform)に反映する。プレイヤーは前進/後退に応じて位置だけで
// なく向き(Yaw)も変わるため、毎フレームTransformを一から組み立て直す。
//
// 扉が全開(DoorOpened)になった後は、プレイヤー入力(マイク/デバッグキー)を
// 無視してLinkを自動で前進させ続け、扉をすり抜けた位置
// (renderer.DoorPassThroughZ)まで進んだらonFieldTransitionを1回だけ呼ぶ。
func (g *Game) Update(dt float64) {
	g.door.Update(dt)
	g.scene.Objects[g.doorIndex].Transform = renderer.DoorTransform(g.door.Progress)

	if !g.autoWalking && g.door.State == world.DoorOpened {
		g.autoWalking = true
	}
	if g.autoWalking {
		player.Player.SetDirection(player.Forward)
	}

	deltaZ := player.Player.Update(dt)
	if deltaZ != 0 {
		g.scene.Objects[g.link.Index].Transform = player.Player.Transform(g.link.LocalTransform)
	}

	if g.autoWalking && !g.transitioned && player.Player.Z <= renderer.DoorPassThroughZ {
		g.transitioned = true
		if g.onFieldTransition != nil {
			g.onFieldTransition()
		}
	}
}

// instance は、ブラウザ側(bridge)からの呼び出しを受けるための
// パッケージレベルのシングルトン。
var instance *Game

// SetInstance はブリッジ経由の呼び出し先となるGameインスタンスを登録する。
func SetInstance(g *Game) {
	instance = g
}

// ganonSwitcher は、ガノンフィールドの背景デザイン案(戦場跡/玉座の間)を
// 切り替えるための関数。cmd/ganon/main.goが起動時に登録する
// (gameパッケージはrenderer.GanonBackgroundVariant等を知る必要がないよう、
// 文字列を受け取るだけのコールバックとして持つ)。
var ganonSwitcher func(variant string)

// SetGanonSwitcher は、ガノンフィールドの背景切り替えを行う関数を登録する。
func SetGanonSwitcher(f func(variant string)) {
	ganonSwitcher = f
}

// SwitchGanonBackground はブリッジ(JavaScript側)から呼ばれ、登録済みの
// 切り替え関数を実行する。未登録の場合(ガノンフィールド以外のページ)は
// 何もしない。
func SwitchGanonBackground(variant string) {
	if ganonSwitcher != nil {
		ganonSwitcher(variant)
	}
}

// horseSummoner は、草原フィールドで馬を呼び出す(Sceneに馬Objectを追加する)
// ための関数。cmd/grassland/main.goが起動時に登録する(ganonSwitcherと
// 同様、gameパッケージがrenderer側の詳細を知る必要がないようコールバックに
// している)。
var horseSummoner func()

// SetHorseSummoner は、草原フィールドで馬を呼び出す関数を登録する。
func SetHorseSummoner(f func()) {
	horseSummoner = f
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

// horseSongPlayed は、馬の歌(music.HorseSongName)が正しく演奏された
// ことを示す。
var horseSongPlayed bool

// HorseSongPlayed は、馬の歌が正しく演奏されたかどうかを返す。
func HorseSongPlayed() bool {
	return horseSongPlayed
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

// onMelodyRecorded は一連の演奏が確定した際に呼ばれ、登録済みの旋律
// パターンと照合する。扉のメロディ(DOOR_MELODY)または時の歌
// (music.SongOfTimeName)が演奏された場合、doorOpenDelayだけ「ため」て
// から隠し扉を開く。時の歌の場合は、その「ため」の間に確認音
// (SongOfTimeConfirmation)→曲を最初から通した自動再生
// (SongOfTimeOpening→SongOfTimeContinuation)も行う。馬の歌
// (music.HorseSongName)が演奏された場合は、確認音→続き(HorseSongContinuation)
// の再生に合わせて馬を呼び出す(SetHorseSummonerで登録された関数を呼ぶ)。
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
		go playSongOfTimeAudio()
	}

	if name == doorMelodyName || name == music.SongOfTimeName {
		fmt.Printf("[music] %s recognized, opening door in %s\n", name, doorOpenDelay)
		time.AfterFunc(doorOpenDelay, OpenDoor)
	}

	if name == music.HorseSongName {
		horseSongPlayed = true
		go playHorseSongAudio()
		if horseSummoner != nil {
			horseSummoner()
		}
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

// playHorseSongAudio は、プレイヤーが馬の歌の合図を演奏した後、確認音
// (music.SongOfTimeConfirmationを共用)に続けてmusic.HorseSongContinuationを
// 自動再生する。再生用フックが未登録の場合は何もしない。
func playHorseSongAudio() {
	audioHooksMu.Lock()
	play, stop, sleep := playNoteHook, stopNoteHook, sleepHook
	audioHooksMu.Unlock()

	if play == nil || stop == nil {
		return
	}
	playNotes(play, stop, sleep, music.SongOfTimeConfirmation)
	sleep(music.SongOfTimePauseDur)
	playNotes(play, stop, sleep, music.HorseSongContinuation)
}

// playNotes はnotesを順番に、1音ずつ鳴らして止めてを繰り返しながら再生する。
func playNotes(play func(note, velocity int), stop func(note int), sleep func(time.Duration), notes []music.ContinuationNote) {
	for _, n := range notes {
		play(n.MIDINote, 100)
		sleep(n.Duration)
		stop(n.MIDINote)
	}
}
