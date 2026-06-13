package ui

// catFrames are the animation frames for the backing-up screen, cycled in order.
// A cat digging up and hauling a data crate (📦), themed for "backing up data".
// Uses a couple of emoji; renders best in modern terminals (Windows Terminal,
// most Linux/macOS terminals). Legacy consoles may show the emoji as a box.
var catFrames = []string{
	`     /\_/\    ✨
    ( o.o )  __
    / >📦< \_/
    |  | |  |
    u  |_|  u`,
	`     /\_/\
    ( -.- )  __
    / >📦< \_/
    |  | |  |
    u__|_|__u`,
	`     /\_/\    ✨
    ( o.o )
    / >  < \__📦
    |  | |   |
    u  |_|   u`,
	`     /\_/\   ✨✨
    ( ^.^ )      📦
    / >  < \    /
    |  | |  \__/
    u__|_|__u`,
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
