# TUI: Animated cat on the backup screen — Design

**Date:** 2026-06-13
**Status:** Approved (pending spec review)

## Goal

Make the "backing up" screen fun by showing an animated ASCII cat (4 frames cycling) plus a "Đang backup ..." message, instead of just a one-line spinner. No real progress percentage (the engine runs the dump tool in one shot and does not report progress), so this is a looping animation that signals "still working".

## Context

`bakdb` Bubble Tea TUI in `ui/`. The backing-up screen is `viewBackingUp()` in `ui/views.go`:

```go
func (m Model) viewBackingUp() string {
	return docStyle.Render(fmt.Sprintf(
		"%s Backing up %s...\n\n%s",
		m.spinner.View(),
		m.dbType,
		helpStyle.Render("This may take a moment depending on the database size."),
	))
}
```

The spinner already ticks: `updateBackingUp` (in `ui/updates.go`) calls `m.spinner.Update(msg)` on each `spinner.TickMsg`, which keeps re-rendering. We reuse that existing tick to advance the cat frame — no new timer needed. `spinnerStyle` (a colored lipgloss style) already exists in `ui/styles.go`.

## Decisions

- ASCII-art cat, 4 frames, cycling continuously while backing up.
- Pure ASCII (no emoji) so it renders in any terminal including Windows PowerShell.
- Frame advances off the existing spinner tick — every few ticks, advance one frame, so it animates at a pleasant pace (not too fast).
- Cat is colored with `spinnerStyle`.
- The result screen (success/fail) is unchanged.
- Keep the descriptive help line, but the whole screen is recomposed around the cat.

## Approach

### Frame counter on the Model

Add one int field to `Model` (`ui/model.go`):
```go
	catFrame int // animation frame counter for the backing-up screen
```

Advance it in `updateBackingUp` (`ui/updates.go`) whenever a spinner tick arrives. To slow the cat relative to the spinner's fast tick, advance the cat frame on every tick but pick the displayed frame as `(catFrame / catSlowdown) % len(catFrames)`. This keeps the cat smooth without a second timer.

### Cat art module

Create `ui/cat.go` holding the frames and a render helper:

```go
package ui

// catFrames are the animation frames for the backing-up screen, cycled in order.
// Pure ASCII so they render on every terminal (including Windows PowerShell).
var catFrames = []string{
	`   /\_/\
  ( o.o )
   > ^ <    ~
  /     \ `,
	`   /\_/\
  ( -.- )
   > ^ <   ~~
  /|   |\ `,
	`   /\_/\
  ( o.o )
   > ^ <  ~~~
  /     \ `,
	`   /\_/\
  ( ^.^ )
   > ^ < ~
  _/   \_`,
}

// catSlowdown controls how many spinner ticks pass before the cat advances one
// frame. Higher = slower cat.
const catSlowdown = 3

// catFrameIndex maps a raw tick counter to the displayed frame index, slowed by
// catSlowdown and wrapped over the frame list.
func catFrameIndex(frame int) int {
	return (frame / catSlowdown) % len(catFrames)
}

// catView returns the cat frame for the given counter, styled.
func catView(frame int) string {
	return spinnerStyle.Render(catFrames[catFrameIndex(frame)])
}
```

### Re-render the backing-up screen

Rewrite `viewBackingUp()` (`ui/views.go`) to stack the cat, the message, and the help line:

```go
func (m Model) viewBackingUp() string {
	var b strings.Builder
	b.WriteString(catView(m.catFrame))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("%s Đang backup %s...", m.spinner.View(), m.dbType))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Vui lòng đợi, tùy kích thước database."))
	return docStyle.Render(b.String())
}
```

The small dot-spinner stays on the message line (it gives the moving "dots" feel next to the text); the cat is the main visual above it. `fmt` and `strings` are already imported in `views.go`.

### Advance the frame on tick

In `ui/updates.go`, `updateBackingUp` currently is:
```go
func (m Model) updateBackingUp(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case backupFinishedMsg:
		m.state = stateResult
		m.databaseName = msg.databaseName
		m.backupFormat = msg.backupFormat
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.message = msg.path
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}
```

Add a frame bump on spinner tick. Change the tail to:
```go
	if _, ok := msg.(spinner.TickMsg); ok {
		m.catFrame++
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}
```

This requires importing `github.com/charmbracelet/bubbles/spinner` in `ui/updates.go` (the spinner type is already used via `m.spinner`, but the `spinner.TickMsg` type name must be imported; verify whether the import already exists and add it if not).

## Data Flow

No change to backup execution. The cat is purely visual. When `backupFinishedMsg` arrives, state switches to `stateResult` exactly as today; the cat disappears with the screen.

## Error Handling

Unchanged. Errors still surface on the result screen.

## Edge Cases

- **Very fast backup:** the cat may only show a frame or two before the result screen replaces it. Acceptable — no special handling.
- **Counter overflow:** `catFrame` is an int incremented per tick; at typical tick rates it would take years to overflow. The modulo keeps the index valid regardless. No reset needed.
- **Windows rendering:** pure ASCII avoids emoji/box-drawing issues; the `~` whiskers and `/\` ears are basic ASCII.

## Testing

The animation is visual, but the frame-selection logic is pure and worth a unit test:

- `catView(frame)` returns a non-empty string for several counters.
- The displayed frame index cycles: for `catSlowdown = 3` and 4 frames, frame counter 0,3,6,9 select art indices 0,1,2,3, and 12 wraps back to 0. Test via a small exported-within-package helper `catFrameIndex(frame int) int` returning `(frame/catSlowdown)%len(catFrames)`, used by `catView`.

Add `ui/cat_test.go`:
```go
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
```

(`catView` calls `catFrameIndex` internally so the two stay consistent.)

## Out of Scope (YAGNI)

- No real progress percentage, no progress bar, no sound, no configurable animations, no other screens changed.
```
