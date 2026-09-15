package renderer

import "testing"

func TestGrasslandObstacleBlocks_ForwardCrossingWithoutJumpIsBlocked(t *testing.T) {
	prevZ := GrasslandObstacleNearZ + 1 // 障害物の手前
	newZ := GrasslandObstacleFarZ - 1   // 障害物を飛び越えて奥まで進もうとした

	blockedZ, blocked := GrasslandObstacleBlocks(prevZ, newZ, false)
	if !blocked {
		t.Fatal("expected forward crossing without jumping to be blocked")
	}
	want := GrasslandObstacleNearZ + grasslandObstacleClampMargin
	if blockedZ != want {
		t.Errorf("blockedZ = %v, want %v (near boundary + margin)", blockedZ, want)
	}
}

func TestGrasslandObstacleBlocks_ForwardCrossingWhileJumpingIsAllowed(t *testing.T) {
	prevZ := GrasslandObstacleNearZ + 1
	newZ := GrasslandObstacleFarZ - 1

	_, blocked := GrasslandObstacleBlocks(prevZ, newZ, true)
	if blocked {
		t.Fatal("expected forward crossing while jumping to be allowed")
	}
}

func TestGrasslandObstacleBlocks_BackwardCrossingWithoutJumpIsBlocked(t *testing.T) {
	prevZ := GrasslandObstacleFarZ - 1 // 障害物の奥
	newZ := GrasslandObstacleNearZ + 1 // 障害物を通り抜けて手前まで戻ろうとした

	blockedZ, blocked := GrasslandObstacleBlocks(prevZ, newZ, false)
	if !blocked {
		t.Fatal("expected backward crossing without jumping to be blocked")
	}
	want := GrasslandObstacleFarZ - grasslandObstacleClampMargin
	if blockedZ != want {
		t.Errorf("blockedZ = %v, want %v (far boundary - margin)", blockedZ, want)
	}
}

func TestGrasslandObstacleBlocks_NoOverlapIsNeverBlocked(t *testing.T) {
	prevZ := GrasslandObstacleNearZ + 5
	newZ := GrasslandObstacleNearZ + 4 // まだ障害物の範囲に入っていない

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
	// 前進して手前でブロックされた直後の位置(境界+余白)から出発する。
	stoppedAt := GrasslandObstacleNearZ + grasslandObstacleClampMargin
	newZ := stoppedAt + 1 // さらに手前(スポーン側)へ後退しようとする

	_, blocked := GrasslandObstacleBlocks(stoppedAt, newZ, false)
	if blocked {
		t.Fatal("expected moving away from the obstacle after being blocked to not be blocked again")
	}
}
