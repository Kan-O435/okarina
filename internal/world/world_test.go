package world

import "testing"

func TestDoorOpenTransitionsThroughStates(t *testing.T) {
	var d Door
	if d.State != DoorClosed {
		t.Fatalf("initial state = %v, want DoorClosed", d.State)
	}

	d.Update(0.5) // Closedのまま: Updateは何もしない
	if d.State != DoorClosed || d.Progress != 0 {
		t.Fatalf("Update before Open() changed state: state=%v progress=%v", d.State, d.Progress)
	}

	d.Open()
	if d.State != DoorOpening {
		t.Fatalf("state after Open() = %v, want DoorOpening", d.State)
	}

	d.Update(0.6) // doorOpenDuration(1.2s)の半分
	if d.State != DoorOpening {
		t.Fatalf("state after partial update = %v, want DoorOpening", d.State)
	}
	if d.Progress <= 0 || d.Progress >= 1 {
		t.Fatalf("progress after partial update = %v, want in (0,1)", d.Progress)
	}

	d.Update(10) // 十分大きいdtで全開まで進める
	if d.State != DoorOpened {
		t.Fatalf("state after long update = %v, want DoorOpened", d.State)
	}
	if d.Progress != 1 {
		t.Fatalf("progress after long update = %v, want 1", d.Progress)
	}

	d.Update(1) // Opened後はこれ以上変化しない
	if d.State != DoorOpened || d.Progress != 1 {
		t.Fatalf("state changed after already Opened: state=%v progress=%v", d.State, d.Progress)
	}
}

func TestDoorOpenIsIdempotentWhileOpening(t *testing.T) {
	var d Door
	d.Open()
	d.Update(0.5)
	progress := d.Progress

	d.Open() // すでにOpening中なので何もしないはず
	if d.Progress != progress || d.State != DoorOpening {
		t.Fatalf("calling Open() while Opening changed state unexpectedly: state=%v progress=%v", d.State, d.Progress)
	}
}
