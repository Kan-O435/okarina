// Package game は、ゲーム全体のループ・状態管理を担当する。
package game

import (
	"fmt"

	"github.com/Kan-O435/okarina/internal/renderer"
	"github.com/Kan-O435/okarina/internal/world"
)

// Run はゲームのエントリーポイント。
// 現時点ではプロジェクトの雛形確認用の最小実装。
func Run() {
	fmt.Println("game.Run() called")
}

// Game は、Sceneと(現時点では)隠し扉の状態をまとめて保持し、毎フレームの
// 更新(Update)をSceneのTransformへ反映する。
type Game struct {
	scene     *renderer.Scene
	doorIndex int
	door      world.Door
}

// New はSceneと、扉Objectのインデックス(BuildFieldDemoSceneが返す)から
// Gameを組み立てる。
func New(scene *renderer.Scene, doorIndex int) *Game {
	return &Game{scene: scene, doorIndex: doorIndex}
}

// OpenDoor は隠し扉を開き始める。何らかの「特定の動作」(将来的にはMIDIの
// メロディ認識、今は動作確認用のボタン)から呼ばれる想定。
func (g *Game) OpenDoor() {
	g.door.Open()
}

// Update はdeltaTime(秒)だけゲーム状態を進め、扉Objectの見た目(Transform)
// に反映する。
func (g *Game) Update(dt float64) {
	g.door.Update(dt)
	g.scene.Objects[g.doorIndex].Transform = renderer.DoorTransform(g.door.Progress)
}

// instance は、ブラウザ側(bridge)からの呼び出しを受けるための
// パッケージレベルのシングルトン。
var instance *Game

// SetInstance はブリッジ経由の呼び出し先となるGameインスタンスを登録する。
func SetInstance(g *Game) {
	instance = g
}

// OpenDoor はブリッジ(JavaScript側)から呼ばれ、登録済みのGameインスタンスの
// 扉を開く。インスタンスが未登録の場合は何もしない。
func OpenDoor() {
	if instance != nil {
		instance.OpenDoor()
	}
}
