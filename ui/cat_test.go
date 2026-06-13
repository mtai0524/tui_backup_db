package ui

import "testing"

func TestCatFrameIndexCycles(t *testing.T) {
	// 8 frames, catSlowdown = 3: counter/3 mod 8 selects the frame.
	cases := map[int]int{0: 0, 3: 1, 6: 2, 9: 3, 12: 4, 15: 5, 18: 6, 21: 7, 24: 0}
	for counter, want := range cases {
		if got := catFrameIndex(counter); got != want {
			t.Fatalf("counter %d: got index %d want %d", counter, got, want)
		}
	}
}

func TestCatViewNonEmpty(t *testing.T) {
	for _, c := range []int{0, 1, 5, 100} {
		if catView(c) == "" {
			t.Fatalf("catView(%d) was empty", c)
		}
	}
}
