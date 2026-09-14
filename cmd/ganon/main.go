package main

import (
	"fmt"
	"runtime"

	"github.com/Kan-O435/okarina/internal/bridge"
	"github.com/Kan-O435/okarina/internal/game"
	"github.com/Kan-O435/okarina/internal/player"
	"github.com/Kan-O435/okarina/internal/renderer"
)

// ガノンフィールド(草原の右奥の木を抜けた先)用のエントリーポイント。
// 草原フィールド(cmd/grassland)と同じplayerパッケージ・ブリッジ関数を
// 再利用し、オタマトーンのピッチ入力/矢印キーによるLinkの移動・向き変更を
// そのまま使えるようにする。ボス戦固有のゲームロジックはまだ無い。
//
// 背景は2つのデザイン案(戦場跡/玉座の間)を比較中のため、ページ下部の
// ボタンから実行時に切り替えられるようにしている(game.SetGanonSwitcher
// 経由でJSからの呼び出しを受ける)。
func main() {
	fmt.Println("Ganon field initialized")

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

		var scene *renderer.Scene
		var link renderer.LinkPlacement

		// loadVariant は、指定した背景デザイン案でシーンを丸ごと作り直し、
		// 現在表示中のシーン・Linkの位置をそれに差し替える。
		loadVariant := func(variant renderer.GanonBackgroundVariant) {
			newScene, newLink, err := renderer.BuildGanonScene(ctx, variant)
			if err != nil {
				fmt.Println("renderer: failed to build ganon scene:", err)
				return
			}
			scene = newScene
			link = newLink
			player.Player.Z = link.SpawnZ
			fmt.Println("renderer: ganon scene rebuilt, objects:", len(scene.Objects))
		}

		loadVariant(renderer.GanonBackgroundBattle)

		game.SetGanonSwitcher(func(variant string) {
			// JS(ボタンのクリックイベント)から直接呼ばれるjs.Funcコールバック
			// の中でLoadGLBMesh(テクスチャデコードの完了をチャネル受信で
			// 待つ)を直接実行すると、そのままの呼び出しスタックでブロック
			// してしまう可能性があるため、新しいゴルーチンで実行する。
			go func() {
				switch variant {
				case "hall":
					fmt.Println("ganon: switching to hall (旧案) background")
					loadVariant(renderer.GanonBackgroundHall)
				default:
					fmt.Println("ganon: switching to battle (現行案) background")
					loadVariant(renderer.GanonBackgroundBattle)
				}
			}()
		})

		ctx.RunLoop(func(dt float64) {
			if scene == nil {
				return
			}
			deltaZ := player.Player.Update(dt)
			if deltaZ != 0 {
				scene.Objects[link.Index].Transform = player.Player.Transform(link.LocalTransform)
			}
			scene.Render(ctx)
		})
		fmt.Println("renderer: ganon scene rendered, game loop started")

		select {}
	}
}
