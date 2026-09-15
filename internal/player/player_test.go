package player

import (
	"math"
	"testing"

	"github.com/Kan-O435/okarina/internal/vecmath"
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
	if s.Yaw != math.Pi {
		t.Fatalf("expected Yaw=Pi while moving forward, got %v", s.Yaw)
	}

	s.OnPitch(300) // Backward
	s.Update(0.1)
	if s.Yaw != 0 {
		t.Fatalf("expected Yaw=0 while moving backward, got %v", s.Yaw)
	}

	// Idleになっても、直前に向いていた方向を保つ(不自然に正面へ戻さない)。
	s.OnPitch(300) // 変化なし → Idle
	s.Update(0.1)
	if s.Direction != Idle {
		t.Fatalf("expected Idle, got %v", s.Direction)
	}
	if s.Yaw != 0 {
		t.Fatalf("expected Yaw to stay at 0 while Idle, got %v", s.Yaw)
	}
}

func TestUpdate_SpeedMultiplier(t *testing.T) {
	base := &State{}
	base.SetDirection(Forward)
	baseDeltaZ := base.Update(1.0)

	boosted := &State{SpeedMultiplier: 2.0}
	boosted.SetDirection(Forward)
	boostedDeltaZ := boosted.Update(1.0)

	if boostedDeltaZ != baseDeltaZ*2 {
		t.Fatalf("expected boosted deltaZ to be 2x base (base=%v, boosted=%v)", baseDeltaZ, boostedDeltaZ)
	}
}

func TestUpdate_ZeroSpeedMultiplierActsAsOne(t *testing.T) {
	zeroValue := &State{}
	explicitOne := &State{SpeedMultiplier: 1.0}

	zeroValue.SetDirection(Forward)
	explicitOne.SetDirection(Forward)

	if got, want := zeroValue.Update(1.0), explicitOne.Update(1.0); got != want {
		t.Fatalf("expected zero-value SpeedMultiplier to behave like 1.0 (got=%v, want=%v)", got, want)
	}
}

func TestStartJump_RisesThenLands(t *testing.T) {
	s := &State{}
	s.StartJump()

	if !s.IsJumping() {
		t.Fatal("expected IsJumping() to be true right after StartJump()")
	}

	// 弧の途中(半分)ではY方向オフセットが正(空中にいる)はず。
	s.Update(jumpDuration / 2)
	mid := s.Transform(vecmath.Identity())
	if mid[13] <= 0 {
		t.Fatalf("expected a positive Y offset mid-jump, got %v", mid[13])
	}
	if !s.IsJumping() {
		t.Fatal("expected still jumping halfway through the arc")
	}

	// 弧の残り半分を進めると着地する。
	s.Update(jumpDuration/2 + 0.01)
	if s.IsJumping() {
		t.Fatal("expected IsJumping() to be false after the jump duration elapses")
	}
	landed := s.Transform(vecmath.Identity())
	if landed[13] != 0 {
		t.Fatalf("expected Y offset to be 0 after landing, got %v", landed[13])
	}
}

func TestUpdate_ContinuesLastDirectionWhileJumpingEvenIfIdle(t *testing.T) {
	s := &State{}
	s.SetDirection(Forward)
	s.Update(0.1) // 前進中の向きをlastMoveDirectionに記録させる

	s.StartJump()
	s.SetDirection(Idle) // ジャンプのジェスチャー自体がDirectionをIdleにする想定
	before := s.Z

	deltaZ := s.Update(0.1)

	if deltaZ == 0 {
		t.Fatal("expected the player to keep moving forward while jumping even if Direction is Idle")
	}
	if s.Z >= before {
		t.Fatalf("expected Z to decrease (moving forward) while jumping, got before=%v after=%v", before, s.Z)
	}
}

func TestUpdate_StaysStillWhileJumpingIfNeverMoved(t *testing.T) {
	s := &State{}
	s.StartJump()

	deltaZ := s.Update(0.1)

	if deltaZ != 0 {
		t.Fatalf("expected no movement while jumping with no prior direction, got deltaZ=%v", deltaZ)
	}
}

func TestStartJump_IgnoredWhileAlreadyJumping(t *testing.T) {
	s := &State{}
	s.StartJump()
	s.Update(jumpDuration / 2)
	elapsedBefore := s.jumpElapsed

	s.StartJump() // 既にジャンプ中なので無視されるはず
	if s.jumpElapsed != elapsedBefore {
		t.Fatalf("expected StartJump() to be a no-op while already jumping, elapsed changed from %v to %v", elapsedBefore, s.jumpElapsed)
	}
}
