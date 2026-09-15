package main

import (
	"fmt"
	"math"
	"runtime"

	"github.com/Kan-O435/okarina/internal/bridge"
	"github.com/Kan-O435/okarina/internal/game"
	"github.com/Kan-O435/okarina/internal/player"
	"github.com/Kan-O435/okarina/internal/renderer"
	"github.com/Kan-O435/okarina/internal/vecmath"
)

// horseSpeedMultiplier は、馬に乗っている間のLinkの移動速度倍率
// (徒歩の何倍で進むか)。
const horseSpeedMultiplier = 2.0

// grasslandFirstObstacleCenterZ/grasslandSongSheetRangeZ は、「馬の歌」の
// 楽譜HUDを表示するかどうかの判定に使う。最初の柵(道に複数並べたうちの
// 1つ目、grasslandObstacleZs[0])からこの距離以内にプレイヤーが近づいたら
// 表示する(神殿フィールドの扉と楽譜HUDの関係と同様)。馬の歌は一度覚えれば
// 以降の柵でも使えるため、2つ目以降の柵の近くでは表示しない。
var grasslandFirstObstacleCenterZ = func() float64 {
	near, far := renderer.GrasslandObstacleBoundsAt(0)
	return (near + far) / 2
}()

const grasslandSongSheetRangeZ = 4.0

// 草原フィールド(扉を抜けた先)用のエントリーポイント。神殿フィールド
// (cmd/game)と同じplayerパッケージ・ブリッジ関数を再利用し、オタマトーンの
// ピッチ入力/矢印キーによるLinkの移動・向き変更をそのまま使えるようにする。
// 隠し扉のような草原フィールド固有のゲームロジックはまだ無い。
func main() {
	fmt.Println("Grassland field initialized")

	if runtime.GOOS == "js" {
		bridge.Init()

		ctx, err := renderer.NewContext("game-canvas")
		if err != nil {
			fmt.Println("renderer: failed to initialize:", err)
		} else {
			width, height := ctx.CanvasSize()
			ctx.Viewport(width, height)
			ctx.EnableDepthTest()
			ctx.EnableBlend()                    // 城の後ろの後光(透過テクスチャ)を正しく合成するため
			ctx.ClearColor(0.6, 0.48, 0.65, 1.0) // 薄紫の不穏な空

			scene, link, projection, err := renderer.BuildGrasslandScene(ctx)
			if err != nil {
				fmt.Println("renderer: failed to build grassland scene:", err)
			} else {
				player.Player.Z = link.SpawnZ
				transitioned := false
				horseIndex := -1
				var horseLocalTransform vecmath.Mat4

				// cameraTransitionElapsed は、馬に乗ってから経過した時間(秒)。
				// 馬に乗った瞬間にカメラを一気に切り替えると視点が急に変わって
				// 分かりにくいため、renderer.GrasslandCameraTransitionDuration
				// かけて、見下ろし気味の固定カメラからマリオのような横視点
				// カメラへ滑らかに補間する(下のRunLoop参照)。
				cameraTransitionElapsed := 0.0

				// 馬の歌が演奏されたら、Linkと同じ場所に馬を呼び出し、以後
				// Linkが馬に乗って移動しているように見せる。すでに呼んで
				// いる場合は何もしない(馬を増やさない)。
				game.SetHorseSummoner(func() {
					if horseIndex >= 0 {
						return
					}
					horse, localTransform, err := renderer.BuildHorseObject(ctx)
					if err != nil {
						fmt.Println("renderer: failed to build horse object:", err)
						return
					}
					horseLocalTransform = localTransform
					horse.Transform = player.Player.Transform(horseLocalTransform)
					scene.Objects = append(scene.Objects, horse)
					horseIndex = len(scene.Objects) - 1
					player.Player.SpeedMultiplier = horseSpeedMultiplier
				})

				// 馬に乗っている間、低い音を2回連続で鳴らすとジャンプする
				// (道の途中の柵を飛び越えるために使う)。馬に乗っていなければ
				// 無視する。
				game.SetJumpTrigger(func() {
					if horseIndex < 0 {
						return
					}
					player.Player.StartJump()
				})

				mountOffset := vecmath.Translate(vecmath.NewVec3(0, renderer.LinkMountHeight, 0))

				horseSongHUD, err := renderer.BuildHorseSongSheetHUD(ctx, width, height)
				if err != nil {
					fmt.Println("renderer: failed to build horse song sheet HUD:", err)
				}

				ctx.RunLoop(func(dt float64) {
					prevZ := player.Player.Z
					deltaZ := player.Player.Update(dt)

					// 道の途中の柵は、ジャンプ中でなければ通り抜けられない。
					// 横切ろうとしていたら、その手前で止める。
					if deltaZ != 0 {
						if blockedZ, blocked := renderer.GrasslandObstacleBlocks(prevZ, player.Player.Z, player.Player.IsJumping()); blocked {
							player.Player.Z = blockedZ
						}
					}

					// ジャンプ中は、前後に動いていなくても(deltaZ==0でも)
					// 見た目のY方向オフセットが変わり続けるため、その間は
					// 毎フレームTransformを更新する。
					if deltaZ != 0 || player.Player.IsJumping() {
						linkTransform := player.Player.Transform(link.LocalTransform)
						if horseIndex >= 0 {
							// 馬に乗っている間は、Linkを持ち上げて背に
							// 座っているように見せる。位置・向きは馬と共通
							// (player.Player)なので、進行方向にあわせて
							// 一緒に向きが変わる。
							linkTransform = mountOffset.Mul(linkTransform)
							scene.Objects[horseIndex].Transform = player.Player.Transform(horseLocalTransform)
						}
						scene.Objects[link.Index].Transform = linkTransform
					}

					// 馬に乗った後は、カメラをマリオのような横視点へ
					// 滑らかに切り替える。乗る前の固定カメラ(eye/target)から、
					// プレイヤーのZ座標に追従する横視点カメラ(eye/target)へ、
					// GrasslandCameraTransitionDuration秒かけて線形補間する。
					// 遷移が終わった後(t=1)も、この式は毎フレームsideのeye/
					// targetを現在のplayer.Player.Zから求め直すため、その
					// まま横視点でプレイヤーを追い続ける。
					if horseIndex >= 0 {
						cameraTransitionElapsed += dt
						t := cameraTransitionElapsed / renderer.GrasslandCameraTransitionDuration
						if t > 1 {
							t = 1
						}
						t = t * t * (3 - 2*t) // smoothstep: 始点・終点で速度0になる滑らかな遷移

						defaultEye, defaultTarget := renderer.GrasslandDefaultCameraEyeTarget()
						sideEye, sideTarget := renderer.GrasslandSideCameraEyeTarget(player.Player.Z)
						eye := defaultEye.Lerp(sideEye, t)
						target := defaultTarget.Lerp(sideTarget, t)
						view := vecmath.LookAt(eye, target, renderer.GrasslandCameraUp())
						scene.ViewProjection = projection.Mul(view)
					}

					scene.Render(ctx)

					// 馬の歌の楽譜は、障害物に近づいた時だけ表示する。
					// 馬の歌を演奏し終えたら(呼び出し済みになったら)消す。
					if horseSongHUD != nil && !game.HorseSongPlayed() &&
						math.Abs(player.Player.Z-grasslandFirstObstacleCenterZ) <= grasslandSongSheetRangeZ {
						horseSongHUD.Render(ctx, width, height)
					}

					// Linkが右奥の木のあたりまで進んだら、次のフィールド
					// (ガノン・玉座の間)へページ遷移する。
					if !transitioned && player.Player.Z <= renderer.GrasslandTreeTriggerZ {
						transitioned = true
						ctx.Navigate("ganon-hall.html")
					}
				})
				fmt.Println("renderer: grassland scene rendered, game loop started")
			}
		}

		select {}
	}
}
