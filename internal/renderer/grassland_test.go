package renderer

import "testing"

func TestGrasslandObstacleBlocks_ForwardCrossingWithoutJumpIsBlocked(t *testing.T) {
	near, far := GrasslandObstacleBoundsAt(0)
	prevZ := near + 1 // 障害物の手前
	newZ := far - 1   // 障害物を飛び越えて奥まで進もうとした

	blockedZ, blocked := GrasslandObstacleBlocks(prevZ, newZ, false)
	if !blocked {
		t.Fatal("expected forward crossing without jumping to be blocked")
	}
	want := near + grasslandObstacleClampMargin
	if blockedZ != want {
		t.Errorf("blockedZ = %v, want %v (near boundary + margin)", blockedZ, want)
	}
}

func TestGrasslandObstacleBlocks_ForwardCrossingWhileJumpingIsAllowed(t *testing.T) {
	near, far := GrasslandObstacleBoundsAt(0)
	prevZ := near + 1
	newZ := far - 1

	_, blocked := GrasslandObstacleBlocks(prevZ, newZ, true)
	if blocked {
		t.Fatal("expected forward crossing while jumping to be allowed")
	}
}

func TestGrasslandObstacleBlocks_BackwardCrossingWithoutJumpIsBlocked(t *testing.T) {
	near, far := GrasslandObstacleBoundsAt(0)
	prevZ := far - 1 // 障害物の奥
	newZ := near + 1 // 障害物を通り抜けて手前まで戻ろうとした

	blockedZ, blocked := GrasslandObstacleBlocks(prevZ, newZ, false)
	if !blocked {
		t.Fatal("expected backward crossing without jumping to be blocked")
	}
	want := far - grasslandObstacleClampMargin
	if blockedZ != want {
		t.Errorf("blockedZ = %v, want %v (far boundary - margin)", blockedZ, want)
	}
}

func TestGrasslandObstacleBlocks_NoOverlapIsNeverBlocked(t *testing.T) {
	near, _ := GrasslandObstacleBoundsAt(0)
	prevZ := near + 5
	newZ := near + 4 // まだ障害物の範囲に入っていない

	_, blocked := GrasslandObstacleBlocks(prevZ, newZ, false)
	if blocked {
		t.Fatal("expected movement that never reaches the obstacle to be allowed")
	}
}

// TestGrasslandObstacleBlocks_MovingAwayAfterBeingBlockedIsNotStuck は、
// 一度手前の境界+余白まで押し戻された後、そこからさらに離れる方向へ
// 進もうとした場合に、再び(誤って)ブロックされてスタックしないことを
// 検証する回帰テスト。境界ちょうどに止めていた頃は、次のフレームの判定が
// 「まだ範囲内」と誤認識し、離れる方向へも進めなくなるバグがあった。
func TestGrasslandObstacleBlocks_MovingAwayAfterBeingBlockedIsNotStuck(t *testing.T) {
	near, _ := GrasslandObstacleBoundsAt(0)
	// 前進して手前でブロックされた直後の位置(境界+余白)から出発する。
	stoppedAt := near + grasslandObstacleClampMargin
	newZ := stoppedAt + 1 // さらに手前(スポーン側)へ後退しようとする

	_, blocked := GrasslandObstacleBlocks(stoppedAt, newZ, false)
	if blocked {
		t.Fatal("expected moving away from the obstacle after being blocked to not be blocked again")
	}
}

// TestGrasslandObstacleBlocks_SecondObstacleAlsoBlocks は、複数配置した
// 柵のうち、1つ目を通過した後の2つ目でも同様にブロックされることを
// 検証する(単一の障害物から複数へ一般化した際の回帰テスト)。
func TestGrasslandObstacleBlocks_SecondObstacleAlsoBlocks(t *testing.T) {
	if GrasslandObstacleCount() < 2 {
		t.Fatal("expected at least two obstacles to be configured")
	}
	near, far := GrasslandObstacleBoundsAt(1)
	prevZ := near + 1
	newZ := far - 1

	blockedZ, blocked := GrasslandObstacleBlocks(prevZ, newZ, false)
	if !blocked {
		t.Fatal("expected forward crossing of the second obstacle without jumping to be blocked")
	}
	want := near + grasslandObstacleClampMargin
	if blockedZ != want {
		t.Errorf("blockedZ = %v, want %v (near boundary + margin)", blockedZ, want)
	}
}

// TestGrasslandObstacleBlocks_BetweenObstaclesIsNeverBlocked は、2つの
// 障害物の間の区間(どちらの範囲にも入っていない)では移動がブロック
// されないことを検証する。
func TestGrasslandObstacleBlocks_BetweenObstaclesIsNeverBlocked(t *testing.T) {
	if GrasslandObstacleCount() < 2 {
		t.Fatal("expected at least two obstacles to be configured")
	}
	_, far0 := GrasslandObstacleBoundsAt(0)
	near1, _ := GrasslandObstacleBoundsAt(1)
	mid := (far0 + near1) / 2

	_, blocked := GrasslandObstacleBlocks(mid+0.1, mid-0.1, false)
	if blocked {
		t.Fatal("expected movement strictly between two obstacles to be allowed")
	}
}
