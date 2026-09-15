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

// ガノンフィールド「戦場跡」案用のエントリーポイント。玉座の間案
// (cmd/ganon-hall)とは別ページ・別Linkに分けており、このページでは
// 戦場跡の背景だけを固定で表示する。
//
// 草原フィールド(cmd/grassland)と同じplayerパッケージ・ブリッジ関数を
// 再利用し、オタマトーンのピッチ入力/矢印キーによるLinkの移動・向き変更を
// そのまま使えるようにする。ボス戦固有のゲームロジックはまだ無い。
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
		ctx.EnableBlend()                     // 楽譜HUD(透過テクスチャ)を正しく合成するため
		ctx.ClearColor(0.15, 0.05, 0.05, 1.0) // 暗く不穏な赤黒い空気

		scene, link, err := renderer.BuildGanonScene(ctx, renderer.GanonBackgroundBattle)
		rockParts, rockPartsErr := renderer.LoadRockDebrisParts(ctx)
		melodySheetHUD, melodySheetErr := renderer.BuildGanonBattleMelodySheetHUD(ctx, width, height)
		if err != nil {
			fmt.Println("renderer: failed to build ganon scene:", err)
		} else if rockPartsErr != nil {
			fmt.Println("renderer: failed to load rock debris parts:", rockPartsErr)
		} else if melodySheetErr != nil {
			fmt.Println("renderer: failed to build ganon battle melody sheet HUD:", melodySheetErr)
		} else {
			player.Player.SpawnAt(link.SpawnZ)
			renderer.SetLinkTransform(scene, link, player.Player.Transform(link.LocalTransform))

			wipeObj, wipeLocalTransform := renderer.GanonWipeRockObject(rockParts)
			wipeY := ganonBattleWipeRockStartY
			wipeObj.Transform = vecmath.Translate(vecmath.NewVec3(0, wipeY, ganonBattleWipeRockZ)).Mul(wipeLocalTransform)
			wipeIndex := len(scene.Objects)
			scene.Objects = append(scene.Objects, wipeObj)
			wipeFalling := true

			ctx.RunLoop(func(dt float64) {
				player.Player.Update(dt)
				renderer.SetLinkTransform(scene, link, player.Player.Transform(link.LocalTransform))
				if wipeFalling {
					wipeY -= ganonBattleWipeRockSpeed * dt
					if wipeY <= ganonBattleWipeRockFallEnd {
						wipeY = ganonBattleWipeRockFallEnd
						wipeFalling = false
					}
					scene.Objects[wipeIndex].Transform = vecmath.Translate(vecmath.NewVec3(0, wipeY, ganonBattleWipeRockZ)).Mul(wipeLocalTransform)
				}
				scene.Render(ctx)

				// 嵐の歌の楽譜は、撃破演出が始まる前(まだ正しく演奏できて
				// いない間)だけ表示する。
				if !game.GanonBattleMelodyPlayed() {
					melodySheetHUD.Render(ctx, width, height)
				}
			})
			fmt.Println("renderer: ganon (battle) scene rendered, game loop started")
		}

		select {}
	}
}
