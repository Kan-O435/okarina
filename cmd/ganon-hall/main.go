package main

import (
	"fmt"
	"runtime"

	"github.com/Kan-O435/okarina/internal/bridge"
	"github.com/Kan-O435/okarina/internal/game"
	"github.com/Kan-O435/okarina/internal/player"
	"github.com/Kan-O435/okarina/internal/renderer"
	"github.com/Kan-O435/okarina/internal/vecmath"
)

// ganonHallCollapseRock は、崩落演出中に降っている岩1個の実行時状態。
// Mesh/Color等・localTransform(原点中心のスケール・向き)はrenderer.
// GanonHallCollapseRockObject/GanonWipeRockObjectで作成済みのため、
// 毎フレームはX/Zを保った垂直移動のTransform更新だけを行う。
type ganonHallCollapseRock struct {
	index          int
	x, z           float64
	localTransform vecmath.Mat4
	y              float64
}

// 崩落演出のパラメータ。「特定の演奏で倒す」処理はまだ無いデバッグ実装のため、
// 見た目が破綻しない程度の値を決め打ちしている。
const (
	ganonHallCollapseRockStartY  = 20.0
	ganonHallCollapseRockFallEnd = -2.0
	ganonHallCollapseFallSpeed   = 14.0
)

// ganonHallWipeRockZ は、崩落演出のクライマックスで画面を覆う「wipe岩」の
// Z位置。カメラ(renderer.GanonHallCameraZ)の1.5手前に置くことで、
// 近距離ゆえに(そこまで大きなモデルでなくても)画面をほぼ覆って見える
// ようにする。X=0(カメラの正面)。
const ganonHallWipeRockZ = renderer.GanonHallCameraZ - 1.5

// ganonHallWipeRockTriggerY は、wipe岩がこのワールドY以下まで落ちたら
// 「画面をほぼ覆った」とみなし、そのタイミングでganon-battle.htmlへ
// 遷移する。カメラの高さ(renderer.GanonHallCameraY)と同じにしている
// (wipe岩の上端がちょうどカメラの高さに重なる瞬間)。
const ganonHallWipeRockTriggerY = renderer.GanonHallCameraY

// ganonHallSmokePuff は、Ganon撃破演出の最初に立ち上る煙玉1個の実行時状態。
// 見た目(Mesh/Texture/Color)はrenderer.GanonHallSmokeObjectで作成済みの
// Objectを使い回し、毎フレームは配置(renderer.GanonHallSmokePuffPlacement)
// と経過時間の割合から位置・大きさだけを再計算する。
type ganonHallSmokePuff struct {
	index     int
	placement renderer.GanonHallSmokePuffPlacement
}

// ganonHallSmokeDuration は、煙演出の長さ(秒)。この間、岩はまだ降らせず、
// 煙だけを見せる(「岩が降ってくる前に、ガノンから煙が出て倒した感じを
// 軽く演出してほしい」という要望に対応)。
const ganonHallSmokeDuration = 0.6

// ganonHallPhase は、Ganonフィールド(玉座の間)の演出の進行状態。
type ganonHallPhase int

const (
	ganonHallIdle ganonHallPhase = iota
	ganonHallSmoking
	ganonHallCollapsing
)

// ガノンフィールド「玉座の間」案用のエントリーポイント。草原フィールドの
// 木を抜けた先はこちらへ遷移する(cmd/grassland参照)。戦場跡案
// (cmd/ganon-battle)とは別ページ・別Linkに分けており、このページでは
// 玉座の間の背景だけを固定で表示する。
//
// 草原フィールド(cmd/grassland)と同じplayerパッケージ・ブリッジ関数を
// 再利用し、オタマトーンのピッチ入力/矢印キーによるLinkの移動・向き変更を
// そのまま使えるようにする。ボス戦固有のゲームロジックはまだ無い。
func main() {
	fmt.Println("Ganon field (hall) initialized")

	if runtime.GOOS == "js" {
		bridge.Init()

		ctx, err := renderer.NewContext("game-canvas")
		if err != nil {
			fmt.Println("renderer: failed to initialize:", err)
			select {}
		}

		width, height := ctx.CanvasSize()
		ctx.Viewport(width, height)
		ctx.EnableDepthTest()
		ctx.EnableBlend()                     // 煙(透過テクスチャ)を正しく合成するため
		ctx.ClearColor(0.15, 0.05, 0.05, 1.0) // 暗く不穏な赤黒い空気

		scene, link, err := renderer.BuildGanonScene(ctx, renderer.GanonBackgroundHall)
		rockParts, rockPartsErr := renderer.LoadRockDebrisParts(ctx)
		smokeTemplate, smokeErr := renderer.GanonHallSmokeObject(ctx)
		melodySheetHUD, melodySheetErr := renderer.BuildGanonHallMelodySheetHUD(ctx, width, height)
		if err != nil {
			fmt.Println("renderer: failed to build ganon scene:", err)
		} else if rockPartsErr != nil {
			fmt.Println("renderer: failed to load rock debris parts:", rockPartsErr)
		} else if smokeErr != nil {
			fmt.Println("renderer: failed to load smoke effect:", smokeErr)
		} else if melodySheetErr != nil {
			fmt.Println("renderer: failed to build ganon hall melody sheet HUD:", melodySheetErr)
		} else {
			player.Player.SpawnAt(link.SpawnZ)
			renderer.SetLinkTransform(scene, link, player.Player.Transform(link.LocalTransform))

			// 崩落演出の岩(背景の岩+wipe岩)は起動時に1回だけ生成し、開始
			// 位置(ganonHallCollapseRockStartY)に置いた状態でSceneに加えて
			// おく。トリガー時はこれらのyを開始位置に戻すだけでよく、
			// 「テスト再生」で何度でも繰り返せるようにする(毎回作り直すと
			// scene.Objectsが際限なく増えてしまうため)。
			var rocks []*ganonHallCollapseRock
			for i, p := range renderer.GanonHallCollapseRockPlacements {
				obj, localTransform := renderer.GanonHallCollapseRockObject(rockParts, i, p)
				obj.Transform = vecmath.Translate(vecmath.NewVec3(p.X, ganonHallCollapseRockStartY, p.Z)).Mul(localTransform)
				index := len(scene.Objects)
				scene.Objects = append(scene.Objects, obj)
				rocks = append(rocks, &ganonHallCollapseRock{
					index:          index,
					x:              p.X,
					z:              p.Z,
					localTransform: localTransform,
					y:              ganonHallCollapseRockStartY,
				})
			}

			wipeObj, wipeLocalTransform := renderer.GanonWipeRockObject(rockParts)
			wipeObj.Transform = vecmath.Translate(vecmath.NewVec3(0, ganonHallCollapseRockStartY, ganonHallWipeRockZ)).Mul(wipeLocalTransform)
			wipeIndex := len(scene.Objects)
			scene.Objects = append(scene.Objects, wipeObj)
			wipeRock := &ganonHallCollapseRock{
				index:          wipeIndex,
				x:              0,
				z:              ganonHallWipeRockZ,
				localTransform: wipeLocalTransform,
				y:              ganonHallCollapseRockStartY,
			}

			// 煙玉も岩と同様、起動時に1回だけ生成してSceneに加えておき、
			// トリガー時は経過時間をリセットするだけにする(何度でも
			// 「テスト再生」できるようにするため)。
			var smokePuffs []*ganonHallSmokePuff
			for _, p := range renderer.GanonHallSmokePuffPlacements {
				obj := smokeTemplate
				obj.Transform = vecmath.Translate(vecmath.NewVec3(p.X, p.Y, p.Z)).Mul(vecmath.Scale(vecmath.NewVec3(p.StartHalfSize, p.StartHalfSize, 1)))
				index := len(scene.Objects)
				scene.Objects = append(scene.Objects, obj)
				smokePuffs = append(smokePuffs, &ganonHallSmokePuff{index: index, placement: p})
			}

			phase := ganonHallIdle
			smokeElapsed := 0.0
			navigated := false
			preview := false

			// startSmokeは、Ganon撃破演出の最初の段階(煙を軽く出す)を開始する。
			// 煙が出終わったら(ganonHallSmokeDuration経過)、RunLoop側で
			// 崩落演出(岩+wipe岩)へ自動的に引き継ぐ。preview=falseはGanonの
			// 第一形態を倒した本番の演出、preview=trueは「テスト再生」ボタン
			// 用のプレビュー(遷移せず、演出が終わったら元の状態に戻し、
			// 何度でも試せるようにする)。
			startSmoke := func(isPreview bool) {
				if phase != ganonHallIdle {
					return
				}
				phase = ganonHallSmoking
				smokeElapsed = 0
				navigated = false
				preview = isPreview
			}

			// Ganonの第一形態を倒した演出の開始処理。デバッグボタン
			// (web/ganon-hall.html)からgoDefeatGanonFirstForm経由で呼ばれる
			// 想定(horseSummonerと同じ軽量コールバックパターン)。
			game.SetGanonHallCollapseTrigger(func() { startSmoke(false) })

			// 崩落演出の「テスト再生」ボタン(web/ganon-hall.html)から
			// goPreviewGanonHallCollapse経由で呼ばれる。ページ遷移はせず、
			// その場で演出が終わるところまで見せる。
			game.SetGanonHallCollapsePreviewTrigger(func() { startSmoke(true) })

			ctx.RunLoop(func(dt float64) {
				switch phase {
				case ganonHallIdle:
					player.Player.Update(dt)
					renderer.SetLinkTransform(scene, link, player.Player.Transform(link.LocalTransform))

				case ganonHallSmoking:
					smokeElapsed += dt
					t := smokeElapsed / ganonHallSmokeDuration
					if t > 1 {
						t = 1
					}
					for _, s := range smokePuffs {
						p := s.placement
						size := p.StartHalfSize + (p.EndHalfSize-p.StartHalfSize)*t
						y := p.Y + p.RiseHeight*t
						scene.Objects[s.index].Transform = vecmath.Translate(vecmath.NewVec3(p.X, y, p.Z)).Mul(vecmath.Scale(vecmath.NewVec3(size, size, 1)))
					}
					if smokeElapsed >= ganonHallSmokeDuration {
						// 煙が出終わったら、そのまま崩落演出(岩+wipe岩)へ移る。
						// 岩が降り始めるこの瞬間に効果音を鳴らす。
						phase = ganonHallCollapsing
						for _, r := range rocks {
							r.y = ganonHallCollapseRockStartY
						}
						wipeRock.y = ganonHallCollapseRockStartY
						game.PlayGanonHallCollapseSound()
					}

				case ganonHallCollapsing:
					for _, r := range rocks {
						r.y -= ganonHallCollapseFallSpeed * dt
						if r.y < ganonHallCollapseRockFallEnd {
							r.y = ganonHallCollapseRockFallEnd
						}
						scene.Objects[r.index].Transform = vecmath.Translate(vecmath.NewVec3(r.x, r.y, r.z)).Mul(r.localTransform)
					}
					wipeRock.y -= ganonHallCollapseFallSpeed * dt
					if wipeRock.y < ganonHallCollapseRockFallEnd {
						wipeRock.y = ganonHallCollapseRockFallEnd
					}
					scene.Objects[wipeRock.index].Transform = vecmath.Translate(vecmath.NewVec3(wipeRock.x, wipeRock.y, wipeRock.z)).Mul(wipeRock.localTransform)

					if preview {
						if wipeRock.y <= ganonHallCollapseRockFallEnd {
							phase = ganonHallIdle
						}
					} else if !navigated && wipeRock.y <= ganonHallWipeRockTriggerY {
						navigated = true
						ctx.Navigate("ganon-battle.html")
					}
				}
				scene.Render(ctx)

				// 光のプレリュードの楽譜は、崩落演出が始まる前(まだ正しく
				// 演奏できていない間)だけ表示する。
				if !game.GanonHallMelodyPlayed() {
					melodySheetHUD.Render(ctx, width, height)
				}
			})
			fmt.Println("renderer: ganon (hall) scene rendered, game loop started")
		}

		select {}
	}
}
