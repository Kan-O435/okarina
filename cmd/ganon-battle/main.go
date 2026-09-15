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

// ganon-hall.htmlの崩落演出は、カメラのすぐ正面にwipe岩(renderer.
// GanonWipeRockObject)を落として画面をほぼ覆った瞬間にganon-battle.htmlへ
// 遷移する(cmd/ganon-hall/main.go参照)。このページ側では、読み込んだ
// 直後から同じ岩がすでに画面を覆っている状態で始め、それが下へ通り過ぎて
// いくことで場面が見えてくるようにし、遷移の継ぎ目を岩の動きに紛れさせる
// (岩がganon-hall側からそのまま続けて落ちてきたように見せるのが狙い)。
//
// ganonBattleWipeRockZは、戦場跡案のカメラ(renderer.GanonBattleCameraZ)の
// 1.5手前。ganonBattleWipeRockStartYは、読み込んだ瞬間からカメラの高さ
// (renderer.GanonBattleCameraY)を覆っているように、カメラの高さより
// 少し下(-1.0)から始める(岩の見た目の高さ3.0の範囲にカメラの高さが
// 収まるようにするため)。ganonHallCollapseFallSpeed(cmd/ganon-hall/main.go)
// と同じ速度で、十分下(ganonBattleWipeRockFallEnd)まで落ちたら止める。
const (
	ganonBattleWipeRockZ       = renderer.GanonBattleCameraZ - 1.5
	ganonBattleWipeRockStartY  = renderer.GanonBattleCameraY - 1.0
	ganonBattleWipeRockFallEnd = -10.0
	ganonBattleWipeRockSpeed   = 14.0
)

// ganonBattleDefeatPhase は、Ganon最終形態の撃破演出(嵐→雷→爆発→白
// フェード→エンディングページへの遷移)の進行状態。ganonHallPhase
// (cmd/ganon-hall/main.go)と同じ考え方で、idleの間は通常の操作を続け、
// トリガーが呼ばれたら1段階ずつ進める。
type ganonBattleDefeatPhase int

const (
	ganonBattleDefeatIdle ganonBattleDefeatPhase = iota
	ganonBattleDefeatStorming
	ganonBattleDefeatLightning
	ganonBattleDefeatExploding
	ganonBattleDefeatFading
)

// 各段階の長さ(秒)。トドメの演出なので、テンポよく畳みかけるのではなく、
// じっくり見せる長さにしている。
const (
	ganonBattleDefeatStormingDuration  = 2.5
	ganonBattleDefeatLightningDuration = 0.5
	ganonBattleDefeatExplodingDuration = 1.2
	ganonBattleDefeatFadingDuration    = 1.2
)

// ganonBattleStormSkyColor は、嵐が強まっていく間にClearColorを近づけて
// いく先の色(暗い暴風雨の空)。フィールド通常時の色(0.15, 0.05, 0.05)から
// ここへ線形補間する。
var ganonBattleStormSkyColor = vecmath.NewVec3(0.05, 0.05, 0.09)

// ganonBattleDefeatSmokePuff/ganonBattleDefeatDebrisは、爆発段階で使う
// 煙玉・岩片1個ぶんの実行時状態。ganonHallSmokePuff/ganonHallCollapseRock
// (cmd/ganon-hall/main.go)と同じく、見た目(Mesh/Texture/Color、または
// Mesh/Color+localTransform)は起動時に1回だけ作ってscene.Objectsに積み、
// 毎フレームはindexで対象のTransformだけを上書きする。
type ganonBattleDefeatSmokePuff struct {
	index     int
	placement renderer.GanonHallSmokePuffPlacement
}

type ganonBattleDefeatDebris struct {
	index          int
	placement      renderer.GanonHallCollapseRockPlacement
	localTransform vecmath.Mat4
}

// ガノンフィールド「戦場跡」案用のエントリーポイント。玉座の間案
// (cmd/ganon-hall)とは別ページ・別Linkに分けており、このページでは
// 戦場跡の背景だけを固定で表示する。
//
// 草原フィールド(cmd/grassland)と同じplayerパッケージ・ブリッジ関数を
// 再利用し、オタマトーンのピッチ入力/矢印キーによるLinkの移動・向き変更を
// そのまま使えるようにする。
//
// 起動直後のwipe岩演出(ganon-hallからの遷移演出)とは別に、Ganon最終形態を
// 倒した際の撃破演出(嵐→雷→爆発→白フェード→ending.htmlへの遷移)を
// ganonBattleDefeatPhaseのステートマシンとして持つ。「特定の演奏で倒す」
// メロディ判定はまだ無いため、現時点ではデバッグボタンから
// game.SetGanonBattleDefeatTrigger経由で呼ばれる想定。
func main() {
	fmt.Println("Ganon field (battle) initialized")

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
		ctx.EnableBlend()                     // 嵐雲・煙・白フェードの半透明合成のため
		ctx.ClearColor(0.15, 0.05, 0.05, 1.0) // 暗く不穏な赤黒い空気

		scene, link, err := renderer.BuildGanonScene(ctx, renderer.GanonBackgroundBattle)
		rockParts, rockPartsErr := renderer.LoadRockDebrisParts(ctx)
		smokeTemplate, smokeErr := renderer.GanonHallSmokeObject(ctx)
		whiteFade, whiteFadeErr := renderer.BuildWhiteFadeOverlay(ctx, width, height)
		if err != nil {
			fmt.Println("renderer: failed to build ganon scene:", err)
		} else if rockPartsErr != nil {
			fmt.Println("renderer: failed to load rock debris parts:", rockPartsErr)
		} else if smokeErr != nil {
			fmt.Println("renderer: failed to load smoke effect:", smokeErr)
		} else if whiteFadeErr != nil {
			fmt.Println("renderer: failed to build white fade overlay:", whiteFadeErr)
		} else {
			player.Player.SpawnAt(link.SpawnZ)
			renderer.SetLinkTransform(scene, link, player.Player.Transform(link.LocalTransform))

			// wipe岩の「画面を覆った状態から降りていく」導入演出。撃破演出
			// (ganonBattleDefeatPhase)とは独立して動き、ページ読み込み後
			// 1秒程度で終わる。
			wipeObj, wipeLocalTransform := renderer.GanonWipeRockObject(rockParts)
			wipeY := ganonBattleWipeRockStartY
			wipeObj.Transform = vecmath.Translate(vecmath.NewVec3(0, wipeY, ganonBattleWipeRockZ)).Mul(wipeLocalTransform)
			wipeIndex := len(scene.Objects)
			scene.Objects = append(scene.Objects, wipeObj)
			wipeFalling := true

			phase := ganonBattleDefeatIdle
			phaseElapsed := 0.0
			overlayAlpha := 0.0
			navigated := false

			// 嵐雲は、Ganon最終形態撃破演出が始まって初めてSceneに追加する
			// (トリガー前は普通の空を見せておきたいため)。追加時の元の
			// Transform(位置・スケール・向きが焼き込まれたもの)を控えて
			// おき、毎フレームその上から追加でスケール(0→1)をかけることで
			// 「湧き上がってくる」ように見せる。
			var stormCloudIndices []int
			var stormCloudBaseTransforms []vecmath.Mat4

			// 雷は撃破演出の雷フェーズに入って初めて生成・配置する
			// (idle中はまだ落ちていないため)。
			lightningIndex := -1

			var smokePuffs []*ganonBattleDefeatSmokePuff
			var debris []*ganonBattleDefeatDebris

			// startDefeatは、Ganon最終形態撃破演出を開始する。idle以外の
			// 間に呼ばれても無視する(演出中の多重トリガーを防ぐ、
			// cmd/ganon-hall/main.goのstartSmokeと同様のガード)。
			startDefeat := func() {
				if phase != ganonBattleDefeatIdle {
					return
				}
				phase = ganonBattleDefeatStorming
				phaseElapsed = 0
				navigated = false
			}

			// デバッグボタン(web/ganon-battle.html)からgoDefeatGanonFinalForm
			// 経由で呼ばれる想定。メロディ判定が実装された際は、ここに登録
			// する関数を差し替えずとも、game.TriggerGanonBattleDefeat()を
			// 呼ぶ側を増やすだけで配線できる。
			game.SetGanonBattleDefeatTrigger(startDefeat)

			ctx.RunLoop(func(dt float64) {
				switch phase {
				case ganonBattleDefeatIdle:
					player.Player.Update(dt)
					renderer.SetLinkTransform(scene, link, player.Player.Transform(link.LocalTransform))

				case ganonBattleDefeatStorming:
					if stormCloudIndices == nil {
						clouds, cloudsErr := renderer.GanonBattleStormCloudObjects(ctx)
						if cloudsErr != nil {
							fmt.Println("renderer: failed to build storm cloud objects:", cloudsErr)
							// 嵐雲が作れなくても演出自体は止めない(次の
							// フェーズへ進める)。
							stormCloudIndices = []int{}
						} else {
							for _, obj := range clouds {
								index := len(scene.Objects)
								scene.Objects = append(scene.Objects, obj)
								stormCloudIndices = append(stormCloudIndices, index)
								stormCloudBaseTransforms = append(stormCloudBaseTransforms, obj.Transform)
							}
						}
					}

					phaseElapsed += dt
					t := phaseElapsed / ganonBattleDefeatStormingDuration
					if t > 1 {
						t = 1
					}
					eased := t * t * (3 - 2*t) // smoothstep: 湧き上がりが滑らかになるように

					for i, index := range stormCloudIndices {
						scene.Objects[index].Transform = stormCloudBaseTransforms[i].Mul(vecmath.Scale(vecmath.NewVec3(eased, eased, eased)))
					}

					r := 0.15 + (ganonBattleStormSkyColor.X-0.15)*eased
					g := 0.05 + (ganonBattleStormSkyColor.Y-0.05)*eased
					b := 0.05 + (ganonBattleStormSkyColor.Z-0.05)*eased
					ctx.ClearColor(r, g, b, 1.0)

					if phaseElapsed >= ganonBattleDefeatStormingDuration {
						phase = ganonBattleDefeatLightning
						phaseElapsed = 0
					}

				case ganonBattleDefeatLightning:
					if lightningIndex < 0 {
						lightningObj, lightningLocalTransform, lightningErr := renderer.GanonBattleLightningObject(ctx)
						if lightningErr != nil {
							fmt.Println("renderer: failed to build lightning object:", lightningErr)
						} else {
							lightningObj.Transform = vecmath.Translate(vecmath.NewVec3(0, renderer.GanonBattleLightningStrikeY, renderer.GanonBattleBossZ)).Mul(lightningLocalTransform)
							lightningIndex = len(scene.Objects)
							scene.Objects = append(scene.Objects, lightningObj)
						}
						game.PlayGanonBattleThunderSound()
					}

					phaseElapsed += dt
					t := phaseElapsed / ganonBattleDefeatLightningDuration
					if t > 1 {
						t = 1
					}
					// 閃光: 序盤(30%地点)でピークに達し、フェーズ終端で0に
					// 戻る、素早い明滅。
					const flashPeak = 0.3
					if t <= flashPeak {
						overlayAlpha = t / flashPeak
					} else {
						overlayAlpha = 1 - (t-flashPeak)/(1-flashPeak)
					}

					if phaseElapsed >= ganonBattleDefeatLightningDuration {
						phase = ganonBattleDefeatExploding
						phaseElapsed = 0
						overlayAlpha = 0
					}

				case ganonBattleDefeatExploding:
					if smokePuffs == nil && debris == nil {
						for _, p := range renderer.GanonBattleExplosionSmokePlacements {
							obj := smokeTemplate
							obj.Transform = vecmath.Translate(vecmath.NewVec3(p.X, p.Y, p.Z)).Mul(vecmath.Scale(vecmath.NewVec3(p.StartHalfSize, p.StartHalfSize, 1)))
							index := len(scene.Objects)
							scene.Objects = append(scene.Objects, obj)
							smokePuffs = append(smokePuffs, &ganonBattleDefeatSmokePuff{index: index, placement: p})
						}
						for i, p := range renderer.GanonBattleExplosionDebrisPlacements {
							obj, localTransform := renderer.GanonHallCollapseRockObject(rockParts, i, p)
							obj.Transform = vecmath.Translate(vecmath.NewVec3(0, renderer.GanonBattleLightningStrikeY, renderer.GanonBattleBossZ)).Mul(localTransform)
							index := len(scene.Objects)
							scene.Objects = append(scene.Objects, obj)
							debris = append(debris, &ganonBattleDefeatDebris{index: index, placement: p, localTransform: localTransform})
						}
					}

					phaseElapsed += dt
					t := phaseElapsed / ganonBattleDefeatExplodingDuration
					if t > 1 {
						t = 1
					}

					for _, s := range smokePuffs {
						p := s.placement
						size := p.StartHalfSize + (p.EndHalfSize-p.StartHalfSize)*t
						y := p.Y + p.RiseHeight*t
						scene.Objects[s.index].Transform = vecmath.Translate(vecmath.NewVec3(p.X, y, p.Z)).Mul(vecmath.Scale(vecmath.NewVec3(size, size, 1)))
					}

					for _, d := range debris {
						p := d.placement
						x := p.X * t
						z := renderer.GanonBattleBossZ + (p.Z-renderer.GanonBattleBossZ)*t
						y := renderer.GanonBattleLightningStrikeY + 4*3.0*t*(1-t) // 放物線: 立ち上ってから落ちる
						scene.Objects[d.index].Transform = vecmath.Translate(vecmath.NewVec3(x, y, z)).Mul(d.localTransform)
					}

					if phaseElapsed >= ganonBattleDefeatExplodingDuration {
						phase = ganonBattleDefeatFading
						phaseElapsed = 0
					}

				case ganonBattleDefeatFading:
					phaseElapsed += dt
					t := phaseElapsed / ganonBattleDefeatFadingDuration
					if t > 1 {
						t = 1
					}
					overlayAlpha = t

					if !navigated && phaseElapsed >= ganonBattleDefeatFadingDuration {
						navigated = true
						ctx.Navigate("ending.html")
					}
				}

				if wipeFalling {
					wipeY -= ganonBattleWipeRockSpeed * dt
					if wipeY <= ganonBattleWipeRockFallEnd {
						wipeY = ganonBattleWipeRockFallEnd
						wipeFalling = false
					}
					scene.Objects[wipeIndex].Transform = vecmath.Translate(vecmath.NewVec3(0, wipeY, ganonBattleWipeRockZ)).Mul(wipeLocalTransform)
				}

				scene.Render(ctx)

				if overlayAlpha > 0 {
					whiteFade.Render(ctx, width, height, overlayAlpha)
				}
			})
			fmt.Println("renderer: ganon (battle) scene rendered, game loop started")
		}

		select {}
	}
}
