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
// Mesh/Color等はrenderer.GanonHallCollapseRockObjectで作成済みのため、
// 毎フレームはX/Zを保った垂直移動のTransform更新だけを行う。
type ganonHallCollapseRock struct {
	index     int
	placement renderer.GanonHallCollapseRockPlacement
	y         float64
}

// 崩落演出のパラメータ。「特定の演奏で倒す」処理はまだ無いデバッグ実装のため、
// 見た目が破綻しない程度の値を決め打ちしている。
const (
	ganonHallCollapseRockStartY  = 20.0
	ganonHallCollapseRockFallEnd = -2.0
	ganonHallCollapseFallSpeed   = 14.0
	ganonHallCollapseDuration    = 1.8 // この経過秒数でganon-battle.htmlへ遷移する
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
		ctx.ClearColor(0.15, 0.05, 0.05, 1.0) // 暗く不穏な赤黒い空気

		scene, link, err := renderer.BuildGanonScene(ctx, renderer.GanonBackgroundHall)
		if err != nil {
			fmt.Println("renderer: failed to build ganon scene:", err)
		} else {
			player.Player.SpawnAt(link.SpawnZ)
			renderer.SetLinkTransform(scene, link, player.Player.Transform(link.LocalTransform))

			collapsing := false
			navigated := false
			collapseElapsed := 0.0
			var rocks []*ganonHallCollapseRock

			// Ganonの第一形態を倒した演出の開始処理。デバッグボタン
			// (web/ganon-hall.html)からgoDefeatGanonFirstForm経由で
			// 呼ばれる想定(horseSummonerと同じ軽量コールバックパターン)。
			game.SetGanonHallCollapseTrigger(func() {
				if collapsing {
					return
				}
				collapsing = true
				collapseElapsed = 0.0
				for _, p := range renderer.GanonHallCollapseRockPlacements {
					obj := renderer.GanonHallCollapseRockObject(ctx, p, ganonHallCollapseRockStartY)
					index := len(scene.Objects)
					scene.Objects = append(scene.Objects, obj)
					rocks = append(rocks, &ganonHallCollapseRock{
						index:     index,
						placement: p,
						y:         ganonHallCollapseRockStartY,
					})
				}
			})

			ctx.RunLoop(func(dt float64) {
				if !collapsing {
					player.Player.Update(dt)
					renderer.SetLinkTransform(scene, link, player.Player.Transform(link.LocalTransform))
				} else {
					collapseElapsed += dt
					for _, r := range rocks {
						r.y -= ganonHallCollapseFallSpeed * dt
						if r.y < ganonHallCollapseRockFallEnd {
							r.y = ganonHallCollapseRockFallEnd
						}
						scene.Objects[r.index].Transform = vecmath.Translate(vecmath.NewVec3(r.placement.X, r.y, r.placement.Z))
					}
					if !navigated && collapseElapsed >= ganonHallCollapseDuration {
						navigated = true
						ctx.Navigate("ganon-battle.html")
					}
				}
				scene.Render(ctx)
			})
			fmt.Println("renderer: ganon (hall) scene rendered, game loop started")
		}

		select {}
	}
}
