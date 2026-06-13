# TUI Cat-Playing-Tetris Animation — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the 4-frame cat animation on the backup screen with an 8-frame scene of a cat (left) watching a self-playing Tetris board (right).

**Architecture:** Only the `catFrames` literal (and its doc comment) in `ui/cat.go` changes — 4 frames become 8 hand-drawn frames, each a complete cat+board ASCII block. The frame-cycling logic, the model field, the tick handler, and `catView`/`catFrameIndex` are unchanged. The one test that hard-codes the old frame count (`TestCatFrameIndexCycles`) is updated for 8 frames.

**Tech Stack:** Go 1.26, Bubble Tea, Lipgloss. Module path `bakdb`.

---

## Reference: current code facts (verified)

`ui/cat.go` currently holds `var catFrames []string` (4 frames), `const catSlowdown = 3`, `func catFrameIndex(frame int) int { return (frame / catSlowdown) % len(catFrames) }`, and `func catView(frame int) string { return spinnerStyle.Render(catFrames[catFrameIndex(frame)]) }`.

`ui/cat_test.go` currently has:
- `TestCatFrameIndexCycles` with `cases := map[int]int{0:0, 3:1, 6:2, 9:3, 12:0, 15:1}` — written for 4 frames; **this will break** when there are 8 frames and must be updated.
- `TestCatViewNonEmpty` — checks `catView(c) != ""` for several counters; stays valid.

We keep `catSlowdown = 3`. Touch ONLY `ui/cat.go` and `ui/cat_test.go`.

---

## Task 1: Update the cycle test for 8 frames (TDD red)

Update the test first so it expresses the new 8-frame contract and fails against the current 4-frame data.

**Files:**
- Modify: `ui/cat_test.go` (`TestCatFrameIndexCycles`)

- [ ] **Step 1: Replace TestCatFrameIndexCycles**

In `ui/cat_test.go`, replace the ENTIRE `TestCatFrameIndexCycles` function with:
```go
func TestCatFrameIndexCycles(t *testing.T) {
	// 8 frames, catSlowdown = 3: counter/3 mod 8 selects the frame.
	cases := map[int]int{0: 0, 3: 1, 6: 2, 9: 3, 12: 4, 15: 5, 18: 6, 21: 7, 24: 0}
	for counter, want := range cases {
		if got := catFrameIndex(counter); got != want {
			t.Fatalf("counter %d: got index %d want %d", counter, got, want)
		}
	}
}
```
Leave `TestCatViewNonEmpty` unchanged.

- [ ] **Step 2: Run the test to verify it FAILS**

Run: `go test ./ui/ -run TestCatFrameIndexCycles -v`
Expected: FAIL — with only 4 frames, `catFrameIndex(12)` is `(12/3)%4 = 0`, but the test now wants `4`. The failure proves the test is bound to the new 8-frame contract.

(No commit yet — the data change in Task 2 makes it pass.)

---

## Task 2: Replace catFrames with 8 cat-plus-Tetris frames

**Files:**
- Modify: `ui/cat.go` (the `catFrames` literal and its doc comment)

- [ ] **Step 1: Replace the doc comment and catFrames literal**

In `ui/cat.go`, replace the comment block above `catFrames` AND the entire `var catFrames = []string{ ... }` literal with EXACTLY the following. (Keep `catSlowdown`, `catFrameIndex`, `catView` and the `package ui` line untouched.)

```go
// catFrames are the animation frames for the backing-up screen, cycled in order.
// A cat (left) watches a self-playing Tetris board (right): a piece falls, locks,
// a row fills, then clears with a sparkle, and the loop repeats.
// Uses block/box-drawing glyphs and an emoji; renders best in modern terminals
// (Windows Terminal, most Linux/macOS terminals). Legacy consoles may show boxes.
var catFrames = []string{
	// 1: T-piece enters at the top, cat watching
	`   /\___/\     ┌──────┐
  /  o   o \   │ ▓▓▓  │
 (  == ^ == )  │  ▓   │
  )        (   │      │
 (  )----(  )  │      │
  \ |    | /   │      │
   u      u    └──────┘`,
	// 2: piece falls one row
	`   /\___/\     ┌──────┐
  /  o   o \   │      │
 (  == ^ == )  │ ▓▓▓  │
  )        (   │  ▓   │
 (  )----(  )  │      │
  \ |    | /   │      │
   u      u    └──────┘`,
	// 3: piece near the bottom
	`   /\___/\     ┌──────┐
  /  o   o \   │      │
 (  == ^ == )  │      │
  )        (   │      │
 (  )----(  )  │ ▓▓▓  │
  \ |    | /   │  ▓   │
   u      u    └──────┘`,
	// 4: piece locks at the bottom-left
	`   /\___/\     ┌──────┐
  /  o   o \   │      │
 (  == ^ == )  │      │
  )        (   │      │
 (  )----(  )  │ ▓▓▓  │
  \ |    | /   │██▓███│
   u      u    └──────┘`,
	// 5: a square piece enters at the top
	`   /\___/\     ┌──────┐
  /  o   o \   │   ▓▓ │
 (  == ^ == )  │   ▓▓ │
  )        (   │      │
 (  )----(  )  │ ▓▓▓  │
  \ |    | /   │██▓███│
   u      u    └──────┘`,
	// 6: square locks; bottom row is now full, cat eyes widen
	`   /\___/\     ┌──────┐
  /  O   O \   │      │
 (  == ^ == )  │      │
  )        (   │   ▓▓ │
 (  )----(  )  │ ▓▓▓▓▓│
  \ |    | /   │██████│
   u      u    └──────┘`,
	// 7: full row flashes, sparkle, cat happy (about to clear)
	`   /\___/\     ┌──────┐  ✨
  /  ^   ^ \   │      │
 (  == ^ == )  │      │
  )        (   │   ▓▓ │
 (  )----(  )  │ ▓▓▓▓▓│
  \ |    | /   │██████│  ✨
   u      u    └──────┘`,
	// 8: row cleared, leftovers settle, sparkle fading; loop resets
	`   /\___/\     ┌──────┐
  /  ^   ^ \   │      │
 (  == ^ == )  │      │
  )        (   │      │
 (  )----(  )  │      │
  \ |    | /   │   ▓▓ │
   u      u    └──────┘  ✨`,
}
```

- [ ] **Step 2: Build the package**

Run: `go build ./ui/`
Expected: no errors.

- [ ] **Step 3: Run the cycle test — now PASSES**

Run: `go test ./ui/ -run TestCatFrameIndexCycles -v`
Expected: PASS (there are now 8 frames, so `catFrameIndex(12) = (12/3)%8 = 4`, matching the test).

- [ ] **Step 4: Run all ui tests**

Run: `go test ./ui/`
Expected: PASS (TestCatFrameIndexCycles, TestCatViewNonEmpty, and the focus tests).

- [ ] **Step 5: Check formatting**

Run: `gofmt -l ui/cat.go`
Expected: no output (file is gofmt-clean). If it prints the filename, run `gofmt -w ui/cat.go` and re-run the tests.

- [ ] **Step 6: Build the whole app**

Run: `go build -o bakdb .`
Expected: `bakdb` binary produced, no errors.

- [ ] **Step 7: Commit**

```bash
git add ui/cat.go ui/cat_test.go
git commit -m "feat(tui): cat watching a self-playing Tetris board on backup screen"
```

---

## Task 3: Visual alignment check + manual verification

The frames are hand-aligned; verify the board borders line up and the loop reads well.

- [ ] **Step 1: Eyeball each frame's border alignment**

Open `ui/cat.go`. For each of the 8 frames, confirm the right-side board borders (`┌──────┐`, the `│ ... │` rows, and `└──────┘`) start at the same column within that frame. A frame whose `│` borders are ragged will look broken. Fix spacing if any frame drifts (the cat columns on the left must be the same width on every line of a frame).

NOTE: Box-drawing and block glyphs are width-1 in a monospace terminal; the emoji `✨` is width-2, so it is placed only to the RIGHT of the board (after `┌──────┐` / `└──────┘`) where trailing width does not push anything out of alignment. Do not insert `✨` between the cat and the board.

- [ ] **Step 2: Run the app and watch the animation**

Run: `./bakdb`, pick a DB type, fill required fields, start a backup.
Expected: during backup, the cat sits on the left and the Tetris board on the right shows a piece falling, locking, a row filling to `██████`, then clearing with a ✨, then looping. The cat's eyes widen (`O O`) as the row fills and turn happy (`^ ^`) on clear.

- [ ] **Step 3: Record result**

If the animation loops cleanly with aligned borders, the feature is verified. Note any frame that looked misaligned for a follow-up tweak.
```
