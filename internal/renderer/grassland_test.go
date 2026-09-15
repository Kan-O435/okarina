package renderer

import "testing"

func TestGrasslandObstacleBlocks_ForwardCrossingWithoutJumpIsBlocked(t *testing.T) {
	prevZ := GrasslandObstacleNearZ + 1 // 障害物の手前
	newZ := GrasslandObstacleFarZ - 1   // 障害物を飛び越えて奥まで進もうとした

	blockedZ, blocked := GrasslandObstacleBlocks(prevZ, newZ, false)
	if !blocked {
		t.Fatal("expected forward crossing without jumping to be blocked")
	}
	if blockedZ != GrasslandObstacleNearZ {
		t.Errorf("blockedZ = %v, want %v (near boundary)", blockedZ, GrasslandObstacleNearZ)
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
	if blockedZ != GrasslandObstacleFarZ {
		t.Errorf("blockedZ = %v, want %v (far boundary)", blockedZ, GrasslandObstacleFarZ)
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
