package ui

import "testing"

func TestCatFrameIndexCycles(t *testing.T) {
	cases := map[int]int{0: 0, 3: 1, 6: 2, 9: 3, 12: 0, 15: 1}
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
