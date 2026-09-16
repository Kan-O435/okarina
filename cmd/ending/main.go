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

// explosionPetal は、オタマトーンの音を検出するたびに追加で降らせる花びら
// 1枚ぶんの実行時の状態。常時舞っているending.Petalsと違い、これらは
// 音を検出するたびにspawnExplosionPetalで動的に生成するが、renderer.
// EndingExplosionMaxTotal枚に達した後は最も古いものを再利用する
// (リングバッファ、下記spawnExplosionPetal参照)。地面まで落ちた後は
// landed=trueのままその場に留まり、降り積もった花びらとして残る。
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
// (renderer.BuildEndingScene参照)。オタマトーンで音を鳴らすと、花びらが
// 上空から追加で降ってきて降り積もる演出があるため(game.
// OnEndingPitchDetected参照)、bridge.Init()を呼ぶ。
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

				// explosionsは、音を検出するたびに追加していく花びらの実行時
				// 状態。renderer.EndingExplosionMaxTotal枚まではScene.Objects
				// へ新しいObjectを追加していくが、それに達した後はnext
				// ExplosionSlotが指す最も古い花びらのパラメータを上書きして
				// 再利用する(リングバッファ)。マイクのピッチ検出はノイズで
				// 細かく上下しやすく、短時間に連続発火しうるため、上限無しで
				// 増やし続けるとdraw call数・メモリが際限なく増えて重くなる
				// 不具合があった。
				explosions := make([]explosionPetal, 0, renderer.EndingExplosionMaxTotal)
				nextExplosionSlot := 0

				// spawnExplosionPetalは、Link・ゼルダ姫のすぐ周り
				// (renderer.EndingExplosionArea*)にランダムな位置で花びらを
				// 1枚、上空(renderer.EndingExplosionFallStart*)から降らせる。
				spawnExplosionPetal := func() {
					params := explosionPetal{
						x:             randRange(-renderer.EndingExplosionAreaHalfX, renderer.EndingExplosionAreaHalfX),
						z:             randRange(renderer.EndingExplosionAreaMinZ, renderer.EndingExplosionAreaMaxZ),
						y:             randRange(renderer.EndingExplosionFallStartMin, renderer.EndingExplosionFallStartMax),
						velocityY:     -randRange(renderer.EndingExplosionFallSpeedMin, renderer.EndingExplosionFallSpeedMax),
						swayAmplitude: randRange(0.2, 0.6),
						swayFrequency: randRange(0.5, 1.3),
						swayPhase:     randRange(0, 2*math.Pi),
						spinSpeed:     randSignedRange(1.0, 3.0),
					}

					if len(explosions) < renderer.EndingExplosionMaxTotal {
						ending.Scene.Objects = append(ending.Scene.Objects, renderer.Object{
							Mesh:      ending.PetalMesh,
							Texture:   ending.PetalTexture,
							Transform: vecmath.Identity(),
							Color:     vecmath.NewVec3(1, 1, 1),
						})
						params.objectIndex = len(ending.Scene.Objects) - 1
						explosions = append(explosions, params)
						return
					}

					// 上限に達した後は、既存のObjectをそのまま使い回し、
					// パラメータだけ最も古いものへ上書きする(Scene.Objectsは
					// 増やさない)。一周する頃には、その花びらは既に地面に
					// 降り積もっているはずなので、消えて上空から降り直す
					// ように見える。
					params.objectIndex = explosions[nextExplosionSlot].objectIndex
					explosions[nextExplosionSlot] = params
					nextExplosionSlot = (nextExplosionSlot + 1) % renderer.EndingExplosionMaxTotal
				}

				// オタマトーンの音を検出するたびに(音程が大きく変化した
				// 瞬間、および最初の一音)、花びらを上空から降らせる。
				spawnExplosionBurst := func() {
					for n := 0; n < renderer.EndingExplosionPetalCount; n++ {
						spawnExplosionPetal()
					}
				}
				game.SetEndingPetalFountainTrigger(spawnExplosionBurst)
				game.SetEndingPetalShowerTrigger(spawnExplosionBurst)

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

					// 音を検出して降ってきた花びらは、疑似重力(renderer.
					// EndingExplosionGravity)で徐々に加速しながら落下する。
					// 着地済み(landed=true)のものは、その場(Y=0)で静止させた
					// まま何もしない(地面に積もった花びらとして残す)。
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
