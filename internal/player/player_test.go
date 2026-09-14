package player

import (
	"math"
	"testing"
)

func TestOnPitch_RisingFrequencyMovesForward(t *testing.T) {
	s := &State{}
	s.OnPitch(300) // baseline
	s.OnPitch(400) // 大きく上昇

	if s.Direction != Forward {
		t.Fatalf("expected Forward, got %v", s.Direction)
	}

	deltaZ := s.Update(0.1)
	if deltaZ >= 0 {
		t.Fatalf("expected forward movement to decrease Z, got deltaZ=%v", deltaZ)
	}
}

func TestOnPitch_FallingFrequencyMovesBackward(t *testing.T) {
	s := &State{}
	s.OnPitch(400)
	s.OnPitch(300) // 大きく下降

	if s.Direction != Backward {
		t.Fatalf("expected Backward, got %v", s.Direction)
	}

	deltaZ := s.Update(0.1)
	if deltaZ <= 0 {
		t.Fatalf("expected backward movement to increase Z, got deltaZ=%v", deltaZ)
	}
}

func TestOnPitch_SmallChangeIsIdle(t *testing.T) {
	s := &State{}
	s.OnPitch(440)
	s.OnPitch(441) // ほぼ同じ音程(半音未満の変化)

	if s.Direction != Idle {
		t.Fatalf("expected Idle for a tiny change, got %v", s.Direction)
	}
	if deltaZ := s.Update(0.1); deltaZ != 0 {
		t.Fatalf("expected no movement while Idle, got deltaZ=%v", deltaZ)
	}
}

func TestOnPitch_SilenceResetsBaseline(t *testing.T) {
	s := &State{}
	s.OnPitch(300)
	s.OnPitch(400) // Forwardになるはず
	s.OnPitch(-1)  // 無音(検出できず)

	if s.Direction != Idle {
		t.Fatalf("expected Idle right after silence, got %v", s.Direction)
	}

	// 無音明けの最初の音は新しい基準になるだけで、直前の音との比較はしない。
	s.OnPitch(1000)
	if s.Direction != Idle {
		t.Fatalf("expected Idle on the first pitch after silence, got %v", s.Direction)
	}
}

func TestUpdate_FacesMovementDirection(t *testing.T) {
	s := &State{}
	s.OnPitch(300)
	s.OnPitch(400) // Forward
	s.Update(0.1)
	if s.Yaw != 0 {
		t.Fatalf("expected Yaw=0 while moving forward, got %v", s.Yaw)
	}

	s.OnPitch(300) // Backward
	s.Update(0.1)
	if s.Yaw != math.Pi {
		t.Fatalf("expected Yaw=Pi while moving backward, got %v", s.Yaw)
	}

	// Idleになっても、直前に向いていた方向を保つ(不自然に正面へ戻さない)。
	s.OnPitch(300) // 変化なし → Idle
	s.Update(0.1)
	if s.Direction != Idle {
		t.Fatalf("expected Idle, got %v", s.Direction)
	}
	if s.Yaw != math.Pi {
		t.Fatalf("expected Yaw to stay at Pi while Idle, got %v", s.Yaw)
	}
}
