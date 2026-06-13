package ui

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

// catSlowdown controls how many spinner ticks pass before the cat advances one
// frame. Higher = slower cat.
const catSlowdown = 3

// catFrameIndex maps a raw tick counter to the displayed frame index, slowed by
// catSlowdown and wrapped over the frame list.
func catFrameIndex(frame int) int {
	return (frame / catSlowdown) % len(catFrames)
}

// catView returns the styled cat frame for the given tick counter.
func catView(frame int) string {
	return spinnerStyle.Render(catFrames[catFrameIndex(frame)])
}
