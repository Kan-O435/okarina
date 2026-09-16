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
	player.Player.SpawnAt(link.SpawnZ)
	renderer.SetLinkTransform(scene, link, player.Player.Transform(link.LocalTransform))
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

	player.Player.Update(dt)
	renderer.SetLinkTransform(g.scene, g.link, player.Player.Transform(g.link.LocalTransform))

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

// ganonHallCollapseTrigger は、玉座の間でGanonの第一形態を倒した際の
// 演出(崩落→戦場跡フィールドへページ遷移)を開始する関数。
// cmd/ganon-hall/main.goが起動時に登録する(horseSummonerと同様の
// コールバックパターン)。「特定の演奏で倒す」処理はまだ無いため、
// 現時点ではデバッグボタン(bridge.goのgoDefeatGanonFirstForm)から
// 直接呼ばれる。
var ganonHallCollapseTrigger func()

// SetGanonHallCollapseTrigger は、玉座の間の崩落演出を開始する関数を登録する。
func SetGanonHallCollapseTrigger(f func()) {
	ganonHallCollapseTrigger = f
}

// TriggerGanonHallCollapse はブリッジ(JavaScript側)から呼ばれ、登録済みの
// 崩落演出開始関数を実行する。未登録の場合(玉座の間フィールド以外の
// ページ)は何もしない。
func TriggerGanonHallCollapse() {
	if ganonHallCollapseTrigger != nil {
		ganonHallCollapseTrigger()
	}
}

// ganonHallCollapsePreviewTrigger は、崩落演出を「テスト再生」する(戦場跡
// フィールドへは遷移せず、その場で岩が降って落ち着くところまでを見せて
// 元の状態に戻す)関数。cmd/ganon-hall/main.goが起動時に登録する
// (ganonHallCollapseTriggerと同様のコールバックパターン)。
var ganonHallCollapsePreviewTrigger func()

// SetGanonHallCollapsePreviewTrigger は、崩落演出のテスト再生を開始する
// 関数を登録する。
func SetGanonHallCollapsePreviewTrigger(f func()) {
	ganonHallCollapsePreviewTrigger = f
}

// TriggerGanonHallCollapsePreview はブリッジ(JavaScript側)から呼ばれ、
// 登録済みの崩落演出テスト再生関数を実行する。未登録の場合(玉座の間
// フィールド以外のページ)は何もしない。
func TriggerGanonHallCollapsePreview() {
	if ganonHallCollapsePreviewTrigger != nil {
		ganonHallCollapsePreviewTrigger()
	}
}

// ganonBattleDefeatTrigger は、戦場跡フィールドでGanonの最終形態を倒した際の
// 演出(嵐→雷→爆発→白フェード→エンディングページへの遷移)を開始する関数。
// cmd/ganon-battle/main.goが起動時に登録する(ganonHallCollapseTriggerと
// 同様のコールバックパターン)。onMelodyRecordedが嵐の歌(music.
// GanonBattleMelodyName)を認識した際に、確認音・BGMの再生後にこの
// TriggerGanonBattleDefeat()を呼ぶ。
var ganonBattleDefeatTrigger func()

// SetGanonBattleDefeatTrigger は、戦場跡フィールドのGanon最終形態撃破演出を
// 開始する関数を登録する。
func SetGanonBattleDefeatTrigger(f func()) {
	ganonBattleDefeatTrigger = f
}

// TriggerGanonBattleDefeat はブリッジ(JavaScript側)から呼ばれ、登録済みの
// Ganon最終形態撃破演出開始関数を実行する。未登録の場合(戦場跡フィールド
// 以外のページ)は何もしない。
func TriggerGanonBattleDefeat() {
	if ganonBattleDefeatTrigger != nil {
		ganonBattleDefeatTrigger()
	}
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

// DebugTriggerJump は、オタマトーン無しにキーボードから低い音を鳴らして
// ジャンプジェスチャーが成立したのと同じ効果を発生させるデバッグ用
// エントリーポイント(web/grassland.jsのOキー)。
func DebugTriggerJump() {
	if jumpTrigger != nil {
		jumpTrigger()
	}
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

// melodyHUDVisible は、現在のフィールドで楽譜HUD(および楽譜に添える
// 「〜せよ!」テキスト、web/*.jsのgoMelodyHUDVisible経由)を表示すべき
// かどうかを返す関数。各cmd/*/main.goが、そのフィールドの楽譜HUDと
// 全く同じ表示条件(近づいたら表示・演奏済みなら非表示、など)を渡して
// 起動時に登録する(horseSummoner等と同様のコールバックパターン)。
var melodyHUDVisible func() bool

// SetMelodyHUDVisibleFunc は、現在のフィールドの楽譜HUD表示条件を返す
// 関数を登録する。
func SetMelodyHUDVisibleFunc(f func() bool) {
	melodyHUDVisible = f
}

// MelodyHUDVisible はブリッジ(JavaScript側)から毎フレームポーリングされ、
// 登録済みの楽譜HUD表示条件を評価する。未登録の場合(楽譜HUDが無い
// ページ)はfalseを返す。
func MelodyHUDVisible() bool {
	if melodyHUDVisible == nil {
		return false
	}
	return melodyHUDVisible()
}

// jumpHintVisible は、草原フィールドで「低い音を鳴らしてジャンプ!」の
// ヒントテキスト(web/grassland.jsのgoJumpHintVisible経由)を表示すべき
// かどうかを返す関数。cmd/grassland/main.goが起動時に登録する
// (melodyHUDVisibleと同様のコールバックパターン)。
var jumpHintVisible func() bool

// SetJumpHintVisibleFunc は、ジャンプのヒントテキストの表示条件を返す
// 関数を登録する。
func SetJumpHintVisibleFunc(f func() bool) {
	jumpHintVisible = f
}

// JumpHintVisible はブリッジ(JavaScript側)から毎フレームポーリングされ、
// 登録済みのジャンプヒント表示条件を評価する。未登録の場合(草原
// フィールド以外のページ)はfalseを返す。
func JumpHintVisible() bool {
	if jumpHintVisible == nil {
		return false
	}
	return jumpHintVisible()
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

// endingPetalFountainTrigger は、エンディング画面でオタマトーンの音程が
// 高い音→低い音に変わった瞬間に呼ばれる関数。花びらが地面から爆発的に
// 打ち上がる(下から上へ)演出を開始する。cmd/ending/main.goが起動時に
// 登録する(horseSummoner等と同様のコールバックパターン)。
var endingPetalFountainTrigger func()

// endingPetalShowerTrigger は、エンディング画面でオタマトーンの音程が
// 低い音→高い音に変わった瞬間(および最初の一音)に呼ばれる関数。花びらが
// 上空から爆発的に降り注ぐ(上から下へ)演出を開始する。
var endingPetalShowerTrigger func()

// SetEndingPetalFountainTrigger/SetEndingPetalShowerTrigger は、それぞれの
// 花びら演出を開始する関数を登録する。
func SetEndingPetalFountainTrigger(f func()) {
	endingPetalFountainTrigger = f
}
func SetEndingPetalShowerTrigger(f func()) {
	endingPetalShowerTrigger = f
}

// endingPitchGestureDeadZone は、エンディング画面のピッチジェスチャー判定
// における「音程が変化した」とみなす最小変化量(半音)。プレイヤー移動の
// 判定(internal/player.semitoneDeadZone)と同じ考え方の値を使う。
const endingPitchGestureDeadZone = 0.5

// endingHasPitch/endingLastSemitone は、直前に検出したピッチ(半音値)を
// 保持する、OnEndingPitchDetected専用の状態。
var (
	endingHasPitch     bool
	endingLastSemitone float64
)

// OnEndingPitchDetected はマイクから検出された最新のピッチ(Hz)を受け取り、
// 音程の変化の向きに応じて花びらの演出を開始する。高い音→低い音への変化
// (差がendingPitchGestureDeadZoneを超えて下がった場合)はendingPetal
// FountainTrigger(下から上へ爆発的に)、低い音→高い音への変化(および
// 無音から新しく音が鳴った最初の一音)はendingPetalShowerTrigger
// (上から下へ爆発的に)を呼ぶ。freqが0以下(無音)の場合は基準をリセット
// する(次に音が検出された時点のピッチを新たな基準にする、internal/
// player.State.OnPitchと同じ考え方)。
func OnEndingPitchDetected(freq float64) {
	if freq <= 0 {
		endingHasPitch = false
		return
	}

	semitone := 12 * math.Log2(freq/440)
	if !endingHasPitch {
		// 最初の一音(基準が無い)は、上から降り注ぐ演出をデフォルトにする。
		endingLastSemitone = semitone
		endingHasPitch = true
		if endingPetalShowerTrigger != nil {
			endingPetalShowerTrigger()
		}
		return
	}

	delta := semitone - endingLastSemitone
	switch {
	case delta > endingPitchGestureDeadZone:
		// 低い音→高い音
		if endingPetalShowerTrigger != nil {
			endingPetalShowerTrigger()
		}
	case delta < -endingPitchGestureDeadZone:
		// 高い音→低い音
		if endingPetalFountainTrigger != nil {
			endingPetalFountainTrigger()
		}
	}
	endingLastSemitone = semitone
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

// DebugTriggerTitleStart は、MIDIキーボード/オカリナ無しにキーボードから
// 「ド(C)」を弾いたのと同じ効果を発生させるデバッグ用エントリーポイント
// (web/title.js・web/story.jsのOキー)。
func DebugTriggerTitleStart() {
	if titleStartTrigger != nil {
		go playTitleStartAudio()
	}
}

// OnMIDIEvent はJS-Go Bridge経由で受け取ったMIDIイベントを
// 旋律記録エンジン(internal/music.Recorder)に渡す。あわせて、タイトル
// 画面用に「ド(C、オクターブ不問)のNote On」を即座に検出する
// (旋律認識は無音のタイムアウトを待ってから確定するため、単音への
// 即時反応にはrecorderとは別にここでチェックする)。
func OnMIDIEvent(e midi.Event) {
	recorder.HandleEvent(e)

	if e.IsNoteOn && titleStartTrigger != nil && music.PitchFromMIDINote(e.Note) == music.C {
		go playTitleStartAudio()
	}
}

// playTitleStartAudio は、タイトル画面・ストーリー画面で「ド」が弾かれた
// 際に、次の画面へ遷移する前に短いジングル(music.TitleStartJingle、
// ラ→レ→ミ→ラ)を鳴らす。最後の音を止めた直後に遷移すると余韻(フェード
// アウト)が途中で切れて聞こえるため、music.TitleStartJingleReleaseTailの
// 分だけ待ってから遷移する。再生用フックが未登録(ネイティブビルドや
// JS未初期化時)の場合はジングルを鳴らさずに即座に遷移する。
func playTitleStartAudio() {
	audioHooksMu.Lock()
	play, stop, sleep := playNoteHook, stopNoteHook, sleepHook
	audioHooksMu.Unlock()

	if play != nil && stop != nil {
		playNotes(play, stop, sleep, music.TitleStartJingle)
		sleep(music.TitleStartJingleReleaseTail)
	}
	titleStartTrigger()
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

// TriggerHorseSummon は、馬の歌(music.HorseSongName)を正しく演奏した際と
// 同じ処理(効果音再生+馬を呼び出す)を直接実行する。デバッグボタン
// (web/grassland.html、bridge.goのgoSummonHorse)から、演奏なしで馬に
// 乗った状態を試せるようにするために使う。
func TriggerHorseSummon() {
	horseSongPlayed = true
	go playHorseSongAudio()
	if horseSummoner != nil {
		horseSummoner()
	}
}

// ganonHallMelodyPlayed は、玉座の間の合図(music.GanonHallMelodyName)が
// 正しく演奏されたことを示す。
var ganonHallMelodyPlayed bool

// GanonHallMelodyPlayed は、玉座の間の合図が正しく演奏されたかどうかを
// 返す。
func GanonHallMelodyPlayed() bool {
	return ganonHallMelodyPlayed
}

// ganonBattleMelodyPlayed は、戦場跡の合図(music.GanonBattleMelodyName)が
// 正しく演奏されたことを示す。
var ganonBattleMelodyPlayed bool

// GanonBattleMelodyPlayed は、戦場跡の合図が正しく演奏されたかどうかを
// 返す。
func GanonBattleMelodyPlayed() bool {
	return ganonBattleMelodyPlayed
}

// audioHooksMu は、下のplayNoteHook・stopNoteHook・playConfirmationFanfareHook・
// playHorseJumpSoundHook・playGanonHallCollapseSoundHook・sleepHookへの
// 読み書きを保護する。onMelodyRecordedはgoroutineを起動して非同期に曲の
// 続きを再生するため、bridge.Init()での差し込みやテストでの差し替えと
// 同時に読まれても安全なようにしている。
var audioHooksMu sync.Mutex

// playNoteHook・stopNoteHook は、Goから直接ブラウザの音声再生(Web Audio
// API、web/audio.jsのplayNote/stopNote)を呼び出すためのフック。
// playConfirmationFanfareHookは、確認音(「テレレレレ」、mp3の効果音、
// web/audio.jsのplayConfirmationFanfare)を再生するためのフック。
// playHorseJumpSoundHookは、馬がジャンプした際のいななき効果音
// (web/grassland.jsのplayHorseJumpSound)を再生するためのフック。
// playGanonHallCollapseSoundHookは、玉座の間の崩落演出が始まった際の
// 効果音(web/ganon-hall.jsのplayGanonHallCollapseSound)を再生するための
// フック。playGanonBattleSongHookは、嵐の歌を正しく演奏した後に流す本家の
// BGM(web/ganon-battle.jsのplayGanonBattleSongOfStorms)を再生するための
// フック。playGanonBattleThunderHookは、Ganon最終形態撃破演出中の雷鳴
// (web/ganon-battle.jsで合成)を再生するためのフック。
// playGanonDefeatFanfareHookは、Ganon最終形態を倒した瞬間の撃破ファン
// ファーレ(web/ganon-battle.jsのplayGanonDefeatFanfare)を再生するための
// フック。bridge.Init()がJS側の実装を差し込む。ネイティブビルドやJS
// 未初期化時はnilのまま。
// sleepHookはtime.Sleepの差し替え用(テストで待ち時間を省略する)。
var (
	playNoteHook                   func(note, velocity int)
	stopNoteHook                   func(note int)
	playConfirmationFanfareHook    func()
	playHorseJumpSoundHook         func()
	playGanonHallCollapseSoundHook func()
	playGanonBattleSongHook        func()
	playGanonBattleThunderHook     func()
	playGanonDefeatFanfareHook     func()
	sleepHook                      = time.Sleep
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

// SetPlayConfirmationFanfareFunc は、時の歌・馬の歌の確認音(mp3の効果音)を
// 再生する実装を登録する(bridge.Init()から呼ばれる)。
func SetPlayConfirmationFanfareFunc(f func()) {
	audioHooksMu.Lock()
	playConfirmationFanfareHook = f
	audioHooksMu.Unlock()
}

// SetPlayHorseJumpSoundFunc は、馬がジャンプした際のいななき効果音を
// 再生する実装を登録する(bridge.Init()から呼ばれる)。
func SetPlayHorseJumpSoundFunc(f func()) {
	audioHooksMu.Lock()
	playHorseJumpSoundHook = f
	audioHooksMu.Unlock()
}

// PlayHorseJumpSound は、馬がジャンプした際のいななき効果音(mp3、
// web/assets/audio/horse-jump-neigh.mp3)を再生する。cmd/grassland/
// main.goが、ジャンプジェスチャー成立時(SetJumpTriggerのコールバック内)
// から呼ぶ想定。フックが未登録(ネイティブビルドやJS未初期化時)の場合は
// 何もしない。
func PlayHorseJumpSound() {
	audioHooksMu.Lock()
	play := playHorseJumpSoundHook
	audioHooksMu.Unlock()
	if play != nil {
		play()
	}
}

// SetPlayGanonHallCollapseSoundFunc は、玉座の間の崩落演出が始まった際の
// 効果音を再生する実装を登録する(bridge.Init()から呼ばれる)。
func SetPlayGanonHallCollapseSoundFunc(f func()) {
	audioHooksMu.Lock()
	playGanonHallCollapseSoundHook = f
	audioHooksMu.Unlock()
}

// PlayGanonHallCollapseSound は、玉座の間の崩落演出(岩が降り始める瞬間)
// の効果音(mp3、web/assets/audio/ganon-hall-collapse.mp3)を再生する。
// cmd/ganon-hall/main.goが、崩落演出の開始時に呼ぶ想定。フックが未登録
// (ネイティブビルドやJS未初期化時)の場合は何もしない。
func PlayGanonHallCollapseSound() {
	audioHooksMu.Lock()
	play := playGanonHallCollapseSoundHook
	audioHooksMu.Unlock()
	if play != nil {
		play()
	}
}

// SetPlayGanonBattleSongFunc は、嵐の歌を正しく演奏した後に流す本家の
// BGMを再生する実装を登録する(bridge.Init()から呼ばれる)。
func SetPlayGanonBattleSongFunc(f func()) {
	audioHooksMu.Lock()
	playGanonBattleSongHook = f
	audioHooksMu.Unlock()
}

// SetPlayGanonBattleThunderSoundFunc は、戦場跡フィールドのGanon最終形態
// 撃破演出で雷が落ちた瞬間の雷鳴を再生する実装を登録する(bridge.Init()
// から呼ばれる)。
func SetPlayGanonBattleThunderSoundFunc(f func()) {
	audioHooksMu.Lock()
	playGanonBattleThunderHook = f
	audioHooksMu.Unlock()
}

// PlayGanonBattleThunderSound は、Ganon最終形態撃破演出で雷が落ちた瞬間の
// 雷鳴(mp3等の音声ファイルではなく、web/ganon-battle.js側でWeb Audio
// APIによりその場で合成するノイズ音)を再生する。cmd/ganon-battle/main.goが、
// 雷の閃光が出た瞬間に呼ぶ想定。フックが未登録(ネイティブビルドやJS未
// 初期化時)の場合は何もしない。
func PlayGanonBattleThunderSound() {
	audioHooksMu.Lock()
	play := playGanonBattleThunderHook
	audioHooksMu.Unlock()
	if play != nil {
		play()
	}
}

// SetPlayGanonDefeatFanfareFunc は、Ganon最終形態を倒した際に鳴らす撃破
// ファンファーレを再生する実装を登録する(bridge.Init()から呼ばれる)。
func SetPlayGanonDefeatFanfareFunc(f func()) {
	audioHooksMu.Lock()
	playGanonDefeatFanfareHook = f
	audioHooksMu.Unlock()
}

// PlayGanonDefeatFanfare は、Ganon最終形態を倒した際の撃破ファンファーレ
// (mp3、web/assets/audio/ganon-defeat-fanfare.mp3)を再生する。
// cmd/ganon-battle/main.goが、爆発演出の開始時に呼ぶ想定。ファンファーレが
// 鳴り終わるまでは、白フェード後もrescue.htmlへのページ遷移を待つ
// (cmd/ganon-battle/main.goのganonBattleDefeatFanfareDuration参照)。
// フックが未登録(ネイティブビルドやJS未初期化時)の場合は何もしない。
func PlayGanonDefeatFanfare() {
	audioHooksMu.Lock()
	play := playGanonDefeatFanfareHook
	audioHooksMu.Unlock()
	if play != nil {
		play()
	}
}

// onMelodyRecorded は一連の演奏が確定した際に呼ばれ、登録済みの旋律
// パターンと照合する。扉のメロディ(DOOR_MELODY)または時の歌
// (music.SongOfTimeName)が演奏された場合、doorOpenDelayだけ「ため」て
// から隠し扉を開く。時の歌の場合は、その「ため」の間に確認音
// (「テレレレレ」、mp3の効果音)→曲を最初から通した自動再生
// (SongOfTimeOpening→SongOfTimeContinuation)も行う。馬の歌
// (music.HorseSongName)が演奏された場合は、確認音→続き(HorseSongContinuation)
// の再生に合わせて馬を呼び出す(SetHorseSummonerで登録された関数を呼ぶ)。
// 玉座の間の合図(music.GanonHallMelodyName)が演奏された場合は、確認音
// (「テレレレレ」)を鳴らした後、崩落演出を開始する
// (TriggerGanonHallCollapse、玉座の間フィールド以外では登録済み
// トリガーが無いため何も起きない)。戦場跡の合図(music.
// GanonBattleMelodyName、嵐の歌)が演奏された場合は、確認音→本家のBGM
// (嵐の歌)を鳴らした後、撃破演出を開始する(TriggerGanonBattleDefeat。
// 撃破演出自体はまだ無いため、登録されるまでは何も起きない)。
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
		TriggerHorseSummon()
	}

	if name == music.GanonHallMelodyName {
		ganonHallMelodyPlayed = true
		go playGanonHallMelodyAudio()
	}

	if name == music.GanonBattleMelodyName {
		ganonBattleMelodyPlayed = true
		go playGanonBattleMelodyAudio()
	}
}

// playSongOfTimeAudio は、プレイヤーが演奏した合図に続けて、確認音
// (「テレレレレ」、mp3の効果音)を鳴らした後、本家のゼルダのように曲を
// 最初から(music.SongOfTimeOpening→music.SongOfTimeContinuation)通して
// 自動再生する。再生用フックが未登録(ネイティブビルドやJS未初期化時)の
// 場合は何もしない。
func playSongOfTimeAudio() {
	audioHooksMu.Lock()
	play, stop, sleep, fanfare := playNoteHook, stopNoteHook, sleepHook, playConfirmationFanfareHook
	audioHooksMu.Unlock()

	if play == nil || stop == nil {
		return
	}
	playConfirmationFanfare(fanfare, sleep)
	sleep(music.SongOfTimePauseDur)
	playNotes(play, stop, sleep, music.SongOfTimeOpening)
	playNotes(play, stop, sleep, music.SongOfTimeContinuation)
}

// playHorseSongAudio は、プレイヤーが馬の歌の合図を演奏した後、確認音
// (時の歌と共用)に続けてmusic.HorseSongContinuationを自動再生する。
// 再生用フックが未登録の場合は何もしない。
func playHorseSongAudio() {
	audioHooksMu.Lock()
	play, stop, sleep, fanfare := playNoteHook, stopNoteHook, sleepHook, playConfirmationFanfareHook
	audioHooksMu.Unlock()

	if play == nil || stop == nil {
		return
	}
	playConfirmationFanfare(fanfare, sleep)
	sleep(music.SongOfTimePauseDur)
	playNotes(play, stop, sleep, music.HorseSongContinuation)
}

// playGanonHallMelodyAudio は、プレイヤーが光のプレリュード(玉座の間の
// 合図)を演奏した後、確認音(「テレレレレ」、時の歌・馬の歌と共用)を
// 鳴らしてから崩落演出を開始する(TriggerGanonHallCollapse。演出の中で、
// 岩が降り始める瞬間に別の効果音が鳴る、cmd/ganon-hall/main.go参照)。
// 確認音の再生用フックが未登録(ネイティブビルドやJS未初期化時)の場合は
// 確認音を待たずに崩落演出を開始する。
func playGanonHallMelodyAudio() {
	audioHooksMu.Lock()
	sleep, fanfare := sleepHook, playConfirmationFanfareHook
	audioHooksMu.Unlock()

	playConfirmationFanfare(fanfare, sleep)
	TriggerGanonHallCollapse()
}

// DebugTriggerGanonHallMelody は、演奏無しにキーボードから光のプレリュード
// を正しく演奏したのと同じ効果を発生させるデバッグ用エントリーポイント
// (web/ganon-hall.jsのOキー)。onMelodyRecordedがGanonHallMelodyNameを
// 認識した場合と全く同じ処理(確認音→崩落演出)を行う。
func DebugTriggerGanonHallMelody() {
	ganonHallMelodyPlayed = true
	go playGanonHallMelodyAudio()
}

// playGanonBattleMelodyAudio は、プレイヤーが嵐の歌(戦場跡の合図)を
// 演奏した後、確認音(「テレレレレ」)→本家のBGM(嵐の歌)を鳴らしてから
// 撃破演出を開始する(TriggerGanonBattleDefeat)。BGMの再生用フックが
// 未登録の場合はBGMを待たずに撃破演出を開始する。
func playGanonBattleMelodyAudio() {
	audioHooksMu.Lock()
	sleep, fanfare, song := sleepHook, playConfirmationFanfareHook, playGanonBattleSongHook
	audioHooksMu.Unlock()

	playConfirmationFanfare(fanfare, sleep)
	if song != nil {
		song()
		sleep(music.GanonBattleSongDuration)
	}
	TriggerGanonBattleDefeat()
}

// playConfirmationFanfare は、時の歌・馬の歌を正しく演奏した際の確認音
// (「テレレレレ」、mp3の効果音、web/assets/audio/song-of-time-confirmation.mp3)
// を再生し、その再生時間(music.ConfirmationFanfareDuration)だけ待つ。
// フックが未登録(ネイティブビルドやJS未初期化時)の場合は何もしない。
func playConfirmationFanfare(fanfare func(), sleep func(time.Duration)) {
	if fanfare == nil {
		return
	}
	fanfare()
	sleep(music.ConfirmationFanfareDuration)
}

// playNotes はnotesを順番に、1音ずつ鳴らして止めてを繰り返しながら再生する。
func playNotes(play func(note, velocity int), stop func(note int), sleep func(time.Duration), notes []music.ContinuationNote) {
	for _, n := range notes {
		play(n.MIDINote, 100)
		sleep(n.Duration)
		stop(n.MIDINote)
	}
}
