// Package world は、フィールドの状態(昼夜・天候・ギミックの開閉など)を管理する。
// music パッケージでメロディが認識された際に、対応する状態変化をここで適用する。
package world

// DoorState は隠し扉の開閉状態。
type DoorState int

const (
	DoorClosed DoorState = iota
	DoorOpening
	DoorOpened
)

// doorOpenDuration は扉が全開になるまでの秒数。
const doorOpenDuration = 1.2

// Door は隠し扉の状態(開閉フェーズ・開き具合)を保持する。
type Door struct {
	State    DoorState
	Progress float64 // 0(全閉)〜1(全開)
}

// Open は扉を開き始める。すでに開いている/開き始めている場合は何もしない。
func (d *Door) Open() {
	if d.State == DoorClosed {
		d.State = DoorOpening
	}
}

// Update はdeltaTime(秒)だけ扉の開き具合を進める。
// Opening状態でなければ何もしない。
func (d *Door) Update(dt float64) {
	if d.State != DoorOpening {
		return
	}
	d.Progress += dt / doorOpenDuration
	if d.Progress >= 1 {
		d.Progress = 1
		d.State = DoorOpened
	}
}
