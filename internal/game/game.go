// Package game は、ゲーム全体のループ・状態管理を担当する。
package game

import (
	"fmt"
	"math"
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

// nearDoorRangeZ は、プレイヤーがこの距離以内に扉に近づいたら「近い」と
// みなす範囲(ワールド単位)。楽譜/オカリナのHUD表示の切り替えに使う。
const nearDoorRangeZ = 4.0

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

// IsNearDoor は、プレイヤー(Link)が扉の近くにいるかどうかを返す。
// 楽譜/オカリナのHUD表示を、扉に近づいた時だけ見せるために使う。
func (g *Game) IsNearDoor() bool {
	return math.Abs(player.Player.Z-renderer.DoorCenterZ) <= nearDoorRangeZ
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

// horseSummoner は、草原フィールドで馬を呼び出す(Sceneに馬Objectを追加する)
// ための関数。cmd/grassland/main.goが起動時に登録する(gameパッケージが
// renderer側の詳細を知る必要がないようコールバックにしている)。
var horseSummoner func()

// SetHorseSummoner は、草原フィールドで馬を呼び出す関数を登録する。
func SetHorseSummoner(f func()) {
	horseSummoner = f
}

// jumpPitchThresholdHz は、この値未満の周波数を「低い音」とみなす閾値。
// 低い音を1回鳴らすだけでジャンプを発生させる(馬に乗っている間に
// 障害物を飛び越える、といった用途に使う)。以前は「低い音を2回」の
// ジェスチャーだったが、オタマトーンは指をスライドさせて音程を変える
// ため2回連続で低い音を出す操作が難しく、ジャンプがほぼ成立しない
// 原因になっていたため、1回の立ち上がりで即座に発火するよう簡略化した。
const jumpPitchThresholdHz = 180.0

// jumpGestureLow は、直前の読み取りが低い音の最中だったかどうかを保持
// する。無音や高い音を挟まずに同じ低い音を鳴らし続けている間は
// 再度発火しない(立ち上がりのみで判定する)ようにするために使う。
var jumpGestureLow bool

// jumpTrigger は、低い音の立ち上がりを検出した際に呼ばれる関数。
// cmd/grassland/main.goが起動時に登録する(horseSummonerと同様の
// コールバックパターン)。馬に乗っていない場合は登録側で無視する想定。
var jumpTrigger func()

// SetJumpTrigger は、ジャンプジェスチャーが成立した際に呼び出す関数を
// 登録する。
func SetJumpTrigger(f func()) {
	jumpTrigger = f
}

// updateJumpGesture は、最新のピッチ(Hz)から低い音の立ち上がり(無音・
// 高い音から低い音に変わった瞬間)を検出し、検出したら即座にjumpTrigger
// を呼ぶ。同じ低い音を鳴らし続けている間は再度発火しない。
func updateJumpGesture(freq float64) {
	isLow := freq > 0 && freq < jumpPitchThresholdHz
	if isLow && !jumpGestureLow && jumpTrigger != nil {
		jumpTrigger()
	}
	jumpGestureLow = isLow
}

// OpenDoor はブリッジ(JavaScript側)から呼ばれ、登録済みのGameインスタンスの
// 扉を開く。インスタンスが未登録の場合は何もしない。
func OpenDoor() {
	if instance != nil {
		instance.OpenDoor()
	}
}

// IsNearDoor は、プレイヤーが扉の近くにいるかどうかを返す。楽譜/オカリナの
// HUD表示を切り替えるために、Update()から毎フレーム参照する。インスタンスが
// 未登録の場合はfalseを返す。
func IsNearDoor() bool {
	if instance == nil {
		return false
	}
	return instance.IsNearDoor()
}

// OnPitchDetected はマイクから検出された最新のピッチ(Hz)をプレイヤーの
// 移動方向判定に渡す。ピッチが検出できなかった場合はfreqに0以下を渡す。
// 実際のプレイヤー移動は、Go側で常時回っているゲームループ
// (Context.RunLoop、cmd/game/main.go参照)がGame.Update()経由で進める。
// あわせて、「低い音を2回」のジャンプジェスチャーの検出も進める
// (updateJumpGesture参照)。
func OnPitchDetected(freq float64) {
	player.Player.OnPitch(freq)
	updateJumpGesture(freq)
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

// titleStartTrigger は、タイトル画面で「ド(C)」が弾かれた際に呼ばれる
// 関数。cmd/title/main.goが起動時に登録する(horseSummoner等と同様の
// コールバックパターン)。タイトル画面以外のページでは未登録のまま。
var titleStartTrigger func()

// SetTitleStartTrigger は、タイトル画面の開始トリガーとして呼び出す関数を
// 登録する。
func SetTitleStartTrigger(f func()) {
	titleStartTrigger = f
}

// OnMIDIEvent はJS-Go Bridge経由で受け取ったMIDIイベントを
// 旋律記録エンジン(internal/music.Recorder)に渡す。あわせて、タイトル
// 画面用に「ド(C、オクターブ不問)のNote On」を即座に検出する
// (旋律認識は無音のタイムアウトを待ってから確定するため、単音への
// 即時反応にはrecorderとは別にここでチェックする)。
func OnMIDIEvent(e midi.Event) {
	recorder.HandleEvent(e)

	if e.IsNoteOn && titleStartTrigger != nil && music.PitchFromMIDINote(e.Note) == music.C {
		titleStartTrigger()
	}
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
