package music

import (
	"testing"
	"time"

	"github.com/Kan-O435/okarina/internal/midi"
)

func TestRecorderRecordsMelodyAfterIdleTimeout(t *testing.T) {
	done := make(chan Melody, 1)
	r := NewRecorder(30*time.Millisecond, func(m Melody) {
		done <- m
	})

	notes := []int{60, 62, 64, 67} // C D E G
	for _, n := range notes {
		r.HandleEvent(midi.Event{Note: n, Velocity: 100, IsNoteOn: true})
		r.HandleEvent(midi.Event{Note: n, Velocity: 0, IsNoteOn: false})
	}

	select {
	case melody := <-done:
		want := []Pitch{C, D, E, G}
		got := melody.Pitches()
		if len(got) != len(want) {
			t.Fatalf("got %v pitches, want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("pitch[%d] = %s, want %s", i, got[i], want[i])
			}
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("melody was not recorded within timeout")
	}
}

func TestRecorderIgnoresNoteOffWithoutMatchingNoteOn(t *testing.T) {
	r := NewRecorder(10*time.Millisecond, func(m Melody) {
		t.Errorf("onDone should not be called, got %v", m.Pitches())
	})

	r.HandleEvent(midi.Event{Note: 60, Velocity: 0, IsNoteOn: false})

	time.Sleep(50 * time.Millisecond)
}
