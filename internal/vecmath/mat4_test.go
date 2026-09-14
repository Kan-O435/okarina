package vecmath

import (
	"math"
	"testing"
)

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func TestIdentityMulVec4(t *testing.T) {
	m := Identity()
	v := [4]float64{1, 2, 3, 1}

	got := m.MulVec4(v)
	if got != v {
		t.Errorf("Identity().MulVec4(v) = %v, want %v", got, v)
	}
}

func TestTranslate(t *testing.T) {
	m := Translate(NewVec3(10, 20, 30))
	origin := [4]float64{0, 0, 0, 1}

	got := m.MulVec4(origin)
	want := [4]float64{10, 20, 30, 1}
	if got != want {
		t.Errorf("Translate().MulVec4(origin) = %v, want %v", got, want)
	}
}

func TestScale(t *testing.T) {
	m := Scale(NewVec3(2, 3, 4))
	v := [4]float64{1, 1, 1, 1}

	got := m.MulVec4(v)
	want := [4]float64{2, 3, 4, 1}
	if got != want {
		t.Errorf("Scale().MulVec4(v) = %v, want %v", got, want)
	}
}

func TestMulComposesTranslateThenScale(t *testing.T) {
	// M = Scale(2,2,2) * Translate(1,0,0) は「まずTranslate、次にScale」を意味する。
	m := Scale(NewVec3(2, 2, 2)).Mul(Translate(NewVec3(1, 0, 0)))
	origin := [4]float64{0, 0, 0, 1}

	got := m.MulVec4(origin)
	want := [4]float64{2, 0, 0, 1} // (0,0,0) -> Translate -> (1,0,0) -> Scale -> (2,0,0)
	if got != want {
		t.Errorf("composed Mul().MulVec4(origin) = %v, want %v", got, want)
	}
}

func TestLookAtEyeMapsToOrigin(t *testing.T) {
	eye := NewVec3(0, 0, 5)
	target := NewVec3(0, 0, 0)
	up := NewVec3(0, 1, 0)

	view := LookAt(eye, target, up)
	got := view.MulVec4([4]float64{eye.X, eye.Y, eye.Z, 1})

	// ビュー空間ではカメラ自身は原点に来る。
	if !approxEqual(got[0], 0) || !approxEqual(got[1], 0) || !approxEqual(got[2], 0) {
		t.Errorf("LookAt: eye in view space = %v, want approx (0,0,0,*)", got)
	}
}

func TestPerspectiveProjectsNearFarZ(t *testing.T) {
	proj := Perspective(math.Pi/2, 1.0, 1.0, 100.0)

	// z=-near(視点から見て奥)の点は、透視除算後にNDC z = -1 になるべき。
	nearPoint := [4]float64{0, 0, -1, 1}
	got := proj.MulVec4(nearPoint)
	ndcZ := got[2] / got[3]
	if !approxEqual(ndcZ, -1) {
		t.Errorf("Perspective: near plane NDC z = %v, want -1", ndcZ)
	}

	farPoint := [4]float64{0, 0, -100, 1}
	got = proj.MulVec4(farPoint)
	ndcZ = got[2] / got[3]
	if !approxEqual(ndcZ, 1) {
		t.Errorf("Perspective: far plane NDC z = %v, want 1", ndcZ)
	}
}
