package vecmath

import (
	"math"
	"testing"
)

func vec3ApproxEqual(a, b Vec3) bool {
	const epsilon = 1e-9
	return math.Abs(a.X-b.X) < epsilon && math.Abs(a.Y-b.Y) < epsilon && math.Abs(a.Z-b.Z) < epsilon
}

func TestVec3AddSub(t *testing.T) {
	a := NewVec3(1, 2, 3)
	b := NewVec3(4, 5, 6)

	got := a.Add(b)
	want := NewVec3(5, 7, 9)
	if got != want {
		t.Errorf("Add() = %v, want %v", got, want)
	}

	gotSub := b.Sub(a)
	wantSub := NewVec3(3, 3, 3)
	if gotSub != wantSub {
		t.Errorf("Sub() = %v, want %v", gotSub, wantSub)
	}
}

func TestVec3Dot(t *testing.T) {
	a := NewVec3(1, 2, 3)
	b := NewVec3(4, -5, 6)

	got := a.Dot(b)
	want := 1*4 + 2*-5 + 3*6.0
	if got != want {
		t.Errorf("Dot() = %v, want %v", got, want)
	}
}

func TestVec3Cross(t *testing.T) {
	x := NewVec3(1, 0, 0)
	y := NewVec3(0, 1, 0)

	got := x.Cross(y)
	want := NewVec3(0, 0, 1)
	if got != want {
		t.Errorf("Cross(x, y) = %v, want %v", got, want)
	}
}

func TestVec3Normalize(t *testing.T) {
	v := NewVec3(3, 0, 4)
	got := v.Normalize()
	want := NewVec3(0.6, 0, 0.8)
	if !vec3ApproxEqual(got, want) {
		t.Errorf("Normalize() = %v, want %v", got, want)
	}

	zero := NewVec3(0, 0, 0)
	if got := zero.Normalize(); got != zero {
		t.Errorf("Normalize() of zero vector = %v, want %v", got, zero)
	}
}

func TestVec3Lerp(t *testing.T) {
	a := NewVec3(0, 0, 0)
	b := NewVec3(10, 20, -10)

	if got := a.Lerp(b, 0); got != a {
		t.Errorf("Lerp(t=0) = %v, want %v", got, a)
	}
	if got := a.Lerp(b, 1); got != b {
		t.Errorf("Lerp(t=1) = %v, want %v", got, b)
	}

	want := NewVec3(5, 10, -5)
	if got := a.Lerp(b, 0.5); !vec3ApproxEqual(got, want) {
		t.Errorf("Lerp(t=0.5) = %v, want %v", got, want)
	}
}
