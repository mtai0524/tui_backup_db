# TUI Backup Cat Animation — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the plain one-line "backing up" screen with an animated 4-frame ASCII cat plus a "Đang backup ..." message.

**Architecture:** Add an `ui/cat.go` module holding the ASCII frames and a styled render helper, a `catFrame int` counter on the Model advanced off the existing spinner tick, and rewrite `viewBackingUp()` to stack the cat above the message. No new timer — reuse the spinner tick that already drives re-renders.

**Tech Stack:** Go 1.26, Bubble Tea + Bubbles spinner (already in module), Lipgloss. Module path `bakdb`.

---

## Reference: current code facts (verified)

- `viewBackingUp()` in `ui/views.go` renders `m.spinner.View()` + "Backing up %s..." + a help line. `ui/views.go` imports `fmt`, `strings`, `lipgloss`.
- `updateBackingUp(msg)` in `ui/updates.go` switches on `backupFinishedMsg` (→ result state) and otherwise calls `m.spinner.Update(msg)`. `ui/updates.go` imports `bakdb/backup`, `github.com/charmbracelet/bubbles/textinput`, `tea "github.com/charmbracelet/bubbletea"` — it does NOT yet import `github.com/charmbracelet/bubbles/spinner`.
- `spinnerStyle` (colored lipgloss style) exists in `ui/styles.go`.
- `Model` struct is in `ui/model.go`.

This feature touches only the backing-up screen. Do not change other screens, navigation, or the engine.

---

## Task 1: Cat art module + frame logic (TDD)

**Files:**
- Create: `ui/cat.go`
- Test: `ui/cat_test.go`

- [ ] **Step 1: Write the failing test**

Create `ui/cat_test.go`:
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

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./ui/ -run 'TestCatFrameIndex|TestCatView' -v`
Expected: FAIL — `undefined: catFrameIndex`, `undefined: catView`.

- [ ] **Step 3: Implement the cat module**

Create `ui/cat.go`:
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

// catView returns the styled cat frame for the given tick counter.
func catView(frame int) string {
	return spinnerStyle.Render(catFrames[catFrameIndex(frame)])
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./ui/ -run 'TestCatFrameIndex|TestCatView' -v`
Expected: PASS (2 tests).

- [ ] **Step 5: Build**

Run: `go build ./ui/`
Expected: no errors.

- [ ] **Step 6: Commit**

```bash
git add ui/cat.go ui/cat_test.go
git commit -m "feat(tui): add animated ASCII cat frames for backup screen"
```

---

## Task 2: Add catFrame counter to the Model

**Files:**
- Modify: `ui/model.go` (add field to `Model` struct)

- [ ] **Step 1: Add the field**

In `ui/model.go`, inside the `Model` struct, add this field right after the line `	catFrame int`'s intended neighbor. Concretely, add it immediately after the `backupFormat string ...` line (the last field in the struct), so the struct ends:
```go
	databaseName     string // Tên database vừa backup (dùng cho email)
	backupFormat     string // Định dạng backup vừa thực hiện
	catFrame         int    // animation frame counter for the backing-up screen
}
```

- [ ] **Step 2: Build**

Run: `go build ./ui/`
Expected: no errors (the field is unused so far; Go allows unused struct fields).

- [ ] **Step 3: Run tests**

Run: `go test ./ui/`
Expected: existing tests PASS.

- [ ] **Step 4: Commit**

```bash
git add ui/model.go
git commit -m "feat(tui): add catFrame counter to model"
```

---

## Task 3: Advance the frame on spinner tick

**Files:**
- Modify: `ui/updates.go` (`updateBackingUp` + add spinner import)

- [ ] **Step 1: Add the spinner import**

In `ui/updates.go`, change the import block from:
```go
import (
	"bakdb/backup"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)
```
to:
```go
import (
	"bakdb/backup"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)
```

- [ ] **Step 2: Bump catFrame on tick**

In `ui/updates.go`, replace the ENTIRE `updateBackingUp` function with:
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

	if _, ok := msg.(spinner.TickMsg); ok {
		m.catFrame++
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}
```

- [ ] **Step 3: Build**

Run: `go build ./ui/`
Expected: no errors (spinner import now used by `spinner.TickMsg`).

- [ ] **Step 4: Run tests**

Run: `go test ./ui/`
Expected: existing tests PASS.

- [ ] **Step 5: Commit**

```bash
git add ui/updates.go
git commit -m "feat(tui): advance cat frame on each spinner tick"
```

---

## Task 4: Render the cat on the backing-up screen

**Files:**
- Modify: `ui/views.go` (`viewBackingUp`)

- [ ] **Step 1: Replace viewBackingUp**

In `ui/views.go`, replace the ENTIRE `viewBackingUp` function with:
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

- [ ] **Step 2: Build the package and the app**

Run: `go build ./ui/ && go build -o bakdb .`
Expected: no errors; `bakdb` binary produced.

- [ ] **Step 3: Run all tests**

Run: `go test ./...`
Expected: `ui` tests PASS (cat + focus tests), other packages OK.

- [ ] **Step 4: Commit**

```bash
git add ui/views.go
git commit -m "feat(tui): show animated cat on the backup screen"
```

---

## Task 5: Manual verification

The animation needs a real terminal; verify by hand.

- [ ] **Step 1: Run a backup**

Run `./bakdb`, pick a DB type, fill required fields, start the backup.
Expected: while backing up, a 4-frame ASCII cat animates (ears `/\_/\`, changing face/whiskers, legs shifting) above a "Đang backup <DB>..." line with the small dot spinner, and the help line below. When the backup finishes, the result screen replaces it as before.

- [ ] **Step 2: Confirm Windows rendering (if available)**

On Windows PowerShell, confirm the cat shows as ASCII (no broken boxes). Pure ASCII frames should render cleanly.

- [ ] **Step 3: Record result**

If the cat animates and backup still completes to the result screen, the feature is verified.
```
