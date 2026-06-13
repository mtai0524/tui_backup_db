# TUI: Cat playing Tetris on the backup screen — Design

**Date:** 2026-06-13
**Status:** Approved (pending spec review)

## Goal

Replace the current 4-frame "cat hauling a crate" backup animation with a longer (~8 frame) scene of a cat (left) watching a self-playing Tetris board (right): a piece falls, locks at the bottom, a row fills, the row clears with a sparkle, then the loop repeats. Purely decorative — signals "still working" while the backup runs.

## Context

`bakdb` Bubble Tea TUI in `ui/`. The animation lives entirely in `ui/cat.go`:

```go
var catFrames = []string{ ... }      // the frames, cycled
const catSlowdown = 3                 // ticks per frame advance
func catFrameIndex(frame int) int { return (frame / catSlowdown) % len(catFrames) }
func catView(frame int) string { return spinnerStyle.Render(catFrames[catFrameIndex(frame)]) }
```

`viewBackingUp()` (in `ui/views.go`) calls `catView(m.catFrame)`; `m.catFrame` increments on each spinner tick (in `ui/updates.go`). **This existing machinery is unchanged.** Only the CONTENT of `catFrames` (and possibly `catSlowdown`) changes. The frame-cycling logic, the tests (`catFrameIndex` cycling, `catView` non-empty), the model field, and the tick handler all stay as-is and keep working with any number of frames.

## Decisions

- Layout: cat on the LEFT, Tetris board on the RIGHT, side by side, as one pre-composed ASCII block per frame.
- ~8 frames telling one Tetris beat: piece falls (a few frames) → locks → another piece → a row fills → row clears with ✨ → reset.
- The board is 6 cells wide, 5 rows tall, bordered (`┌─┐ │ │ └─┘`). Blocks drawn with `▓` and `█`; a full row shows as `██████`.
- The cat reuses the thick art (round body, whiskers `== ^ ==`, four feet `u  u`), fixed on the left; only its face shifts slightly (`o o` → `^ ^` when a row clears) for liveliness.
- Each frame is a complete, hand-drawn multi-line string with cat and board already side by side — NO runtime composition. This keeps it simple and guaranteed-aligned, and fits the existing `catFrames []string` mechanism.
- `catSlowdown` may be tuned (e.g. lowered to 2) so the longer loop animates at a pleasant pace; pick a value during implementation and note it.
- Emoji ✨ allowed (already agreed). Block glyphs `▓ █` and box-drawing `┌─┐│└┘` are Unicode — render well in modern terminals (Windows Terminal, Linux/macOS); legacy consoles may show boxes.

## Approach

Only `ui/cat.go`'s `catFrames` literal changes (plus its doc comment, and optionally `catSlowdown`). No other file changes. The 8 frames, in order:

1. T-piece entering at top of board, cat eyes `o o`.
2. T-piece one row lower.
3. T-piece near bottom.
4. T-piece locked at the bottom-left.
5. A second piece (square) entering at top.
6. Square locked next to the T — the bottom row is now full (`██████`), cat eyes widen `O O`.
7. Bottom row flashes full with ✨ beside the board, cat eyes `^ ^` (about to clear).
8. Row cleared — board nearly empty again, leftover block settles, ✨ fades; loop returns to frame 1.

Each frame is a raw string literal. The cat occupies the left columns, the board the right, separated by a few spaces, every line padded so the board's left border stays vertically aligned across all rows of that frame.

Reference frame (illustrative spacing — final art tuned for alignment during implementation):
```
   /\___/\     ┌──────┐
  /  o   o \   │ ▓▓▓  │
 (  == ^ == )  │  ▓   │
  )        (   │      │
 (  )----(  )  │      │
  \ |    | /   │      │
   u    u      └──────┘
```

## Data Flow

No change. `catView(m.catFrame)` returns the current frame string; `viewBackingUp` renders it. When `backupFinishedMsg` arrives, the screen switches to the result view exactly as today.

## Error Handling

Unchanged. Not relevant to a decorative animation.

## Edge Cases

- **Frame count != 4:** `catFrameIndex` uses `% len(catFrames)`, so any count cycles correctly. The existing test `TestCatFrameIndexCycles` is written for the OLD assumption of 4 frames with specific counters → it must be UPDATED to match the new frame count and `catSlowdown` so it still asserts correct cycling. This is the one test change required.
- **Alignment drift:** each frame is hand-aligned; the implementer must visually verify the board's `│` borders line up within each frame (a frame with a ragged border looks broken).
- **Legacy terminals:** block/box glyphs may render as squares on very old consoles. Accepted (same trade-off as the emoji decision).

## Testing

- `TestCatViewNonEmpty` stays valid (every frame non-empty) — unchanged.
- `TestCatFrameIndexCycles` must be rewritten to match the new `len(catFrames)` (8) and the chosen `catSlowdown`. Concretely, if `catSlowdown` stays 3 and there are 8 frames: counters `0,3,6,9,12,15,18,21` map to indices `0,1,2,3,4,5,6,7`, and `24` wraps to `0`. The implementer sets the test cases to whatever the final `catSlowdown`/frame-count produce, and asserts the wrap.
- Manual: run a backup, watch the cat-plus-Tetris loop animate and repeat; confirm borders align and the row-clear beat reads clearly.

## Out of Scope (YAGNI)

- No real Tetris engine (no randomness, no logic) — fixed hand-drawn frames only.
- No runtime composition of cat + board.
- No changes to other screens, navigation, the model, or the tick handler.
```
