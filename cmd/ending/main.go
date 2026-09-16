package main

import (
	"fmt"
	"math"
	"math/rand"
	"runtime"

	"github.com/Kan-O435/okarina/internal/bridge"
	"github.com/Kan-O435/okarina/internal/game"
	"github.com/Kan-O435/okarina/internal/renderer"
	"github.com/Kan-O435/okarina/internal/vecmath"
)

// endingSpinSpeed は、Link・ゼルダ姫の「塊」が1秒間に回転する角度
// (ラジアン)。タイトル画面のオカリナ(titleSpinSpeed)よりゆっくり、
// 静かに回るくらいの速さにしている。
const endingSpinSpeed = 0.35

// randRange は[min, max)の一様乱数を返す(爆発的に舞う花びらの位置・
// 初速等を引くために使う)。
func randRange(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// randSignedRange は、絶対値が[min, max)の範囲でランダムな正または負の
// 値を返す(花びらの回転方向をランダムにするため)。
func randSignedRange(min, max float64) float64 {
	v := randRange(min, max)
	if rand.Intn(2) == 0 {
		return -v
	}
	return v
}

// explosionPetal は、オタマトーンの音程ジェスチャーで追加に舞う花びら
// 1枚ぶんの実行時の状態。常時舞っているending.Petalsと違い、これらは
// ジェスチャーのたびにspawnExplosionPetalで動的に生成する(枚数の上限は
// 設けない)。velocityYにrenderer.EndingExplosionGravityによる疑似重力を
// 適用することで、下から上へ(高い音→低い音)/上から下へ(低い音→高い音)
// のどちらの爆発も同じ運動方程式で表現できる。
type explosionPetal struct {
	objectIndex                             int
	x, z                                    float64
	y, velocityY                            float64
	swayAmplitude, swayFrequency, swayPhase float64
	spinSpeed, angle                        float64
	landed                                  bool
}

// エンディング画面用のエントリーポイント。ガノン最終形態の撃破演出
// (cmd/ganon-battle)から遷移してくる。花畑の上でLinkとゼルダ姫が
// 向かい合って立ち、その2人をまとめて1つの塊とみなしてゆっくり回転させる
// (renderer.BuildEndingScene参照)。オタマトーンの音程を大きく上下させる
// (高い音→低い音、または低い音→高い音)と、花びらが爆発的に追加で舞う
// 演出があるため(game.OnEndingPitchDetected参照)、bridge.Init()を呼ぶ。
func main() {
	fmt.Println("Ending screen initialized")

	if runtime.GOOS == "js" {
		bridge.Init()

		ctx, err := renderer.NewContext("game-canvas")
		if err != nil {
			fmt.Println("renderer: failed to initialize:", err)
		} else {
			width, height := ctx.CanvasSize()
			ctx.Viewport(width, height)
			ctx.EnableDepthTest()
			ctx.EnableBlend()                     // 花びら(透過テクスチャ)を正しく合成するため
			ctx.ClearColor(0.55, 0.75, 0.95, 1.0) // 空のような水色

			ending, err := renderer.BuildEndingScene(ctx)
			if err != nil {
				fmt.Println("renderer: failed to build ending scene:", err)
			} else {
				angle := 0.0

				// 花びら(花吹雪)の実行時の状態(現在のY座標・累積回転角)。
				// 静的なパラメータ(揺れ方・落下速度等)はending.Petals側に
				// 持たせてあるので、ここでは動く値だけを追いかける。
				petalY := make([]float64, len(ending.Petals))
				petalAngle := make([]float64, len(ending.Petals))
				for i, p := range ending.Petals {
					petalY[i] = p.StartY
				}
				elapsed := 0.0

				// explosionsは、音程ジェスチャーのたびに追加していく花びらの
				// 実行時状態。使い回すプールを設けず、生成した分だけ増え続ける
				// (枚数は無制限)。
				var explosions []explosionPetal

				// spawnExplosionPetalは、Link・ゼルダ姫のすぐ周り
				// (renderer.EndingExplosionArea*)にランダムな位置で花びら
				// Objectを1枚新しく生成し、指定した高さ・初速(上向きなら
				// 正、下向きなら負)で運動を開始させる。
				spawnExplosionPetal := func(startY, velocityY float64) {
					ending.Scene.Objects = append(ending.Scene.Objects, renderer.Object{
						Mesh:      ending.PetalMesh,
						Texture:   ending.PetalTexture,
						Transform: vecmath.Identity(),
						Color:     vecmath.NewVec3(1, 1, 1),
					})
					explosions = append(explosions, explosionPetal{
						objectIndex:   len(ending.Scene.Objects) - 1,
						x:             randRange(-renderer.EndingExplosionAreaHalfX, renderer.EndingExplosionAreaHalfX),
						z:             randRange(renderer.EndingExplosionAreaMinZ, renderer.EndingExplosionAreaMaxZ),
						y:             startY,
						velocityY:     velocityY,
						swayAmplitude: randRange(0.2, 0.6),
						swayFrequency: randRange(0.5, 1.3),
						swayPhase:     randRange(0, 2*math.Pi),
						spinSpeed:     randSignedRange(1.0, 3.0),
					})
				}

				// 高い音→低い音: 地面から爆発的に打ち上がる(下から上へ)。
				game.SetEndingPetalFountainTrigger(func() {
					for n := 0; n < renderer.EndingExplosionPetalCount; n++ {
						spawnExplosionPetal(0, randRange(renderer.EndingExplosionRiseSpeedMin, renderer.EndingExplosionRiseSpeedMax))
					}
				})

				// 低い音→高い音(および最初の一音): 上空から爆発的に降り注ぐ
				// (上から下へ)。
				game.SetEndingPetalShowerTrigger(func() {
					for n := 0; n < renderer.EndingExplosionPetalCount; n++ {
						startY := randRange(renderer.EndingExplosionFallStartMin, renderer.EndingExplosionFallStartMax)
						speed := randRange(renderer.EndingExplosionFallSpeedMin, renderer.EndingExplosionFallSpeedMax)
						spawnExplosionPetal(startY, -speed)
					}
				})

				ctx.RunLoop(func(dt float64) {
					angle += endingSpinSpeed * dt
					rotate := vecmath.RotateY(angle)
					for i, local := range ending.PairLocalTransforms {
						ending.Scene.Objects[ending.PairObjectStart+i].Transform = rotate.Mul(local)
					}

					elapsed += dt
					for i, p := range ending.Petals {
						petalY[i] -= p.FallSpeed * dt
						if petalY[i] < 0 {
							// 地面まで落ちたら、また上空から降らせ直す
							// (途切れず舞い続ける花吹雪にするため)。
							petalY[i] = renderer.EndingPetalMaxHeight
						}
						petalAngle[i] += p.SpinSpeed * dt

						x := p.X + math.Sin(elapsed*p.SwayFrequency+p.SwayPhase)*p.SwayAmplitude
						transform := vecmath.Translate(vecmath.NewVec3(x, petalY[i], p.Z)).Mul(vecmath.RotateZ(petalAngle[i]))
						ending.Scene.Objects[p.ObjectIndex].Transform = transform
					}

					// 爆発的に舞う花びらは、疑似重力(renderer.
					// EndingExplosionGravity)で放物運動させる。着地済み
					// (landed=true)のものは、その場(Y=0)で静止させたまま
					// 何もしない(地面に積もった花びらとして残す)。
					for i := range explosions {
						p := &explosions[i]
						if p.landed {
							continue
						}
						p.velocityY -= renderer.EndingExplosionGravity * dt
						p.y += p.velocityY * dt
						p.angle += p.spinSpeed * dt

						if p.y <= 0 {
							p.y = 0
							p.landed = true
						}

						x := p.x + math.Sin(elapsed*p.swayFrequency+p.swayPhase)*p.swayAmplitude
						transform := vecmath.Translate(vecmath.NewVec3(x, p.y, p.z)).Mul(vecmath.RotateZ(p.angle))
						ending.Scene.Objects[p.objectIndex].Transform = transform
					}

					ending.Scene.Render(ctx)
				})
				fmt.Println("renderer: ending scene rendered, game loop started")
			}
		}

		select {}
	}
}
