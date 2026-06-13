# TUI Collapsible Advanced Options — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Hide the 4 optional fields in the TUI backup form behind a collapsible "Advanced options" row, showing only the 5 required fields by default.

**Architecture:** Keep the existing 9-element `inputs` slice and its index map unchanged (engine wiring depends on it). Add a `advancedExpanded bool` to the Model, replace the flat `focusIndex 0..maxFocusIdx` walk with a computed `focusStops()` list of focusable stops (required inputs → advanced toggle → optional inputs if expanded → button), and re-render the details screen to draw the toggle row and conditionally the optional fields.

**Tech Stack:** Go 1.26, Bubble Tea (`github.com/charmbracelet/bubbletea`), Lipgloss, Bubbles textinput — all already in the module. Module path `bakdb`.

---

## Reference: current code facts (already verified)

- `Model` struct is in `ui/model.go`. Fields include `state`, `dbType string`, `inputs []textinput.Model` (len 9), `focusIndex int`.
- Index map: 0 Host, 1 Port, 2 Username, 3 Password, 4 Database (required); 5 Connection String, 6 Binary Path, 7 Output Directory, 8 Backup Format (optional). Field 8 is SQL-Server-only.
- `applyDefaults(inputs, defaults)` in `ui/model.go` sets inputs 0..8 from `config.Defaults`.
- `updateEnterDetails` in `ui/updates.go` currently computes `maxFocusIdx` (7, or 8 for SQL Server), submits when `focusIndex == maxFocusIdx+1` (the button), and loops `for i := range inputs` setting Focus/Blur + PromptStyle/TextStyle.
- `viewEnterDetails` in `ui/views.go` renders title, then `for i := 0; i < displayCount` inputs, then the button via `renderButton`, then help text. `displayCount` is 8, or 9 for SQL Server.
- `renderButton(text string, focused bool) string` and styles `focusedStyle`, `noStyle` exist in `ui/views.go` / `ui/styles.go`.

We are changing ONLY the details-entry screen. Do not touch select-DB, backing-up, result, or email screens.

---

## Task 1: focusStops + auto-expand helpers (pure logic, TDD)

Add the pure navigation helpers and a unit test. These encode the new focus order and the auto-expand rule.

**Files:**
- Create: `ui/focus.go`
- Test: `ui/focus_test.go`
- Modify: `ui/model.go` (add `advancedExpanded bool` field to `Model`)

- [ ] **Step 1: Add the model field**

In `ui/model.go`, inside the `Model` struct, add this field (place it right after `focusIndex   int`):
```go
	advancedExpanded bool // whether the optional-fields section is shown
```

- [ ] **Step 2: Write the failing test**

Create `ui/focus_test.go`:
```go
package ui

import (
	"reflect"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
)

func TestFocusStopsCollapsed(t *testing.T) {
	m := Model{dbType: "MySQL", advancedExpanded: false}
	got := m.focusStops()
	want := []int{0, 1, 2, 3, 4, stopAdvancedToggle, stopButton}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("collapsed: got %v want %v", got, want)
	}
}

func TestFocusStopsExpandedNonSQLServer(t *testing.T) {
	m := Model{dbType: "MySQL", advancedExpanded: true}
	got := m.focusStops()
	want := []int{0, 1, 2, 3, 4, stopAdvancedToggle, 5, 6, 7, stopButton}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expanded non-mssql: got %v want %v", got, want)
	}
}

func TestFocusStopsExpandedSQLServer(t *testing.T) {
	m := Model{dbType: "SQL Server", advancedExpanded: true}
	got := m.focusStops()
	want := []int{0, 1, 2, 3, 4, stopAdvancedToggle, 5, 6, 7, 8, stopButton}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expanded mssql: got %v want %v", got, want)
	}
}

func TestShouldAutoExpand(t *testing.T) {
	mk := func(vals map[int]string) []textinput.Model {
		ins := make([]textinput.Model, 9)
		for i := range ins {
			ins[i] = textinput.New()
		}
		for i, v := range vals {
			ins[i].SetValue(v)
		}
		return ins
	}
	if shouldAutoExpand(mk(map[int]string{0: "localhost", 4: "db"})) {
		t.Fatal("only required fields set → should NOT auto-expand")
	}
	if !shouldAutoExpand(mk(map[int]string{7: "~/backups"})) {
		t.Fatal("optional field 7 set → should auto-expand")
	}
	if !shouldAutoExpand(mk(map[int]string{8: ".bak"})) {
		t.Fatal("optional field 8 set → should auto-expand")
	}
}
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `go test ./ui/ -run 'TestFocusStops|TestShouldAutoExpand' -v`
Expected: FAIL — `undefined: stopAdvancedToggle`, `undefined: focusStops`, `undefined: shouldAutoExpand`.

- [ ] **Step 4: Implement the helpers**

Create `ui/focus.go`:
```go
package ui

import "github.com/charmbracelet/bubbles/textinput"

// Special focus-stop identifiers (negative so they never collide with input indices).
const (
	stopAdvancedToggle = -1 // the "▸/▾ Advanced options" row
	stopButton         = -2 // the "Start Backup" button
)

// focusStops returns the ordered focusable stops for the current form state.
// Required inputs (0-4) are always present; the advanced toggle always follows;
// optional inputs (5-7, plus 8 for SQL Server) appear only when expanded; the
// Start Backup button is last.
func (m Model) focusStops() []int {
	stops := []int{0, 1, 2, 3, 4, stopAdvancedToggle}
	if m.advancedExpanded {
		stops = append(stops, 5, 6, 7)
		if m.dbType == "SQL Server" {
			stops = append(stops, 8)
		}
	}
	stops = append(stops, stopButton)
	return stops
}

// shouldAutoExpand reports whether any optional field (indices 5-8) already has
// a value, in which case the advanced section should start expanded.
func shouldAutoExpand(inputs []textinput.Model) bool {
	for _, i := range []int{5, 6, 7, 8} {
		if inputs[i].Value() != "" {
			return true
		}
	}
	return false
}
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./ui/ -run 'TestFocusStops|TestShouldAutoExpand' -v`
Expected: PASS (4 tests).

- [ ] **Step 6: Build the package**

Run: `go build ./ui/`
Expected: no errors. (The new `advancedExpanded` field is unused so far — that's fine, Go allows unused struct fields.)

- [ ] **Step 7: Commit**

```bash
git add ui/focus.go ui/focus_test.go ui/model.go
git commit -m "feat(tui): add focusStops and auto-expand helpers"
```

---

## Task 2: Set auto-expand in InitialModel

Wire `advancedExpanded` so the section starts open when `.env` pre-filled an optional field.

**Files:**
- Modify: `ui/model.go` (in `InitialModel`, after `applyDefaults(inputs, defaults)`)

- [ ] **Step 1: Set the field after defaults are applied**

In `ui/model.go`, find this block in `InitialModel`:
```go
	applyDefaults(inputs, defaults)

	s := spinner.New()
```
Change it to:
```go
	applyDefaults(inputs, defaults)

	autoExpand := shouldAutoExpand(inputs)

	s := spinner.New()
```
Then, in the `m := Model{...}` literal in the same function, add the field. The literal currently is:
```go
	m := Model{
		state:    stateSelectDB,
		list:     l,
		inputs:   inputs,
		spinner:  s,
		defaults: defaults,
	}
```
Change it to:
```go
	m := Model{
		state:            stateSelectDB,
		list:             l,
		inputs:           inputs,
		spinner:          s,
		defaults:         defaults,
		advancedExpanded: autoExpand,
	}
```

- [ ] **Step 2: Build and run existing tests**

Run: `go build ./ui/ && go test ./ui/`
Expected: build OK; the 4 focus tests still PASS.

- [ ] **Step 3: Commit**

```bash
git add ui/model.go
git commit -m "feat(tui): auto-expand advanced section when .env sets optional fields"
```

---

## Task 3: Rewrite navigation in updateEnterDetails

Replace the flat `focusIndex` walk with a `focusStops()`-based walk. `focusIndex` now indexes into `focusStops()`. Enter on the toggle expands/collapses; Enter on the button submits.

**Files:**
- Modify: `ui/updates.go` (`updateEnterDetails`)

- [ ] **Step 1: Replace the body of updateEnterDetails**

In `ui/updates.go`, replace the ENTIRE `updateEnterDetails` function (currently lines ~43-92) with:
```go
func (m Model) updateEnterDetails(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()
			stops := m.focusStops()
			cur := stops[m.focusIndex]

			// Enter on the advanced toggle expands/collapses instead of moving.
			if s == "enter" && cur == stopAdvancedToggle {
				m.advancedExpanded = !m.advancedExpanded
				// Keep focus on the toggle row after toggling. Its position is
				// fixed at index 5 (after the 5 required inputs).
				m.focusIndex = 5
				return m, m.refocusInputs()
			}

			// Enter on the button submits.
			if s == "enter" && cur == stopButton {
				m.state = stateBackingUp
				return m, tea.Batch(m.spinner.Tick, m.startBackupCmd())
			}

			// Otherwise move between stops.
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}
			last := len(stops) - 1
			if m.focusIndex > last {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = last
			}

			return m, m.refocusInputs()
		}
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

// refocusInputs focuses the textinput at the current stop (if the current stop
// is an input index) and blurs all others, updating prompt/text styles. The
// advanced-toggle and button stops are not inputs, so when they are current no
// input is focused.
func (m *Model) refocusInputs() tea.Cmd {
	stops := m.focusStops()
	current := stops[m.focusIndex]
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		if i == current {
			cmds[i] = m.inputs[i].Focus()
			m.inputs[i].PromptStyle = focusedStyle
			m.inputs[i].TextStyle = focusedStyle
			continue
		}
		m.inputs[i].Blur()
		m.inputs[i].PromptStyle = noStyle
		m.inputs[i].TextStyle = noStyle
	}
	return tea.Batch(cmds...)
}
```

NOTE: `m.focusIndex = 5` is the toggle's fixed position because `focusStops()` always lists the 5 required inputs (indices 0-4 in the stops slice, positions 0-4) followed by the toggle (position 5). Collapsing/expanding never changes the toggle's position.

- [ ] **Step 2: Build**

Run: `go build ./ui/`
Expected: no errors. (`updateInputs` and `startBackupCmd` already exist and are unchanged.)

- [ ] **Step 3: Run tests**

Run: `go test ./ui/`
Expected: the 4 focus tests still PASS (this task changes navigation, not the pure helpers).

- [ ] **Step 4: Commit**

```bash
git add ui/updates.go
git commit -m "feat(tui): navigate via focusStops with collapsible advanced toggle"
```

---

## Task 4: Render the toggle row and conditional optional fields

Update `viewEnterDetails` to draw the 5 required inputs, the toggle row, the optional inputs only when expanded, then the button — with focus styling driven by the current stop.

**Files:**
- Modify: `ui/views.go` (`viewEnterDetails`, and add `renderAdvancedToggle`)

- [ ] **Step 1: Replace viewEnterDetails and add the toggle renderer**

In `ui/views.go`, replace the ENTIRE `viewEnterDetails` function (currently lines ~10-35) with:
```go
func (m Model) viewEnterDetails() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf(" Backup %s Database ", m.dbType)))
	b.WriteString("\n\n")

	stops := m.focusStops()
	current := stops[m.focusIndex]

	// Required fields (0-4), always shown.
	for i := 0; i <= 4; i++ {
		b.WriteString(m.inputs[i].View())
		b.WriteRune('\n')
	}

	// Advanced toggle row.
	b.WriteString("\n")
	b.WriteString(m.renderAdvancedToggle(current == stopAdvancedToggle))
	b.WriteString("\n")

	// Optional fields, indented, only when expanded.
	if m.advancedExpanded {
		optional := []int{5, 6, 7}
		if m.dbType == "SQL Server" {
			optional = append(optional, 8)
		}
		for _, i := range optional {
			b.WriteString("  ")
			b.WriteString(m.inputs[i].View())
			b.WriteRune('\n')
		}
	}

	b.WriteString("\n")
	b.WriteString(m.renderButton("Start Backup", current == stopButton))

	b.WriteString(helpStyle.Render("\n\n (tab/shift+tab: move • enter: next/expand/submit • ctrl+c: quit)"))

	return docStyle.Render(b.String())
}

// renderAdvancedToggle draws the collapsible advanced-options row, marked ▸ when
// collapsed and ▾ when expanded, highlighted when it is the focused stop.
func (m Model) renderAdvancedToggle(focused bool) string {
	marker := "▸"
	if m.advancedExpanded {
		marker = "▾"
	}
	label := marker + " Advanced options"
	if focused {
		return focusedStyle.Render(label)
	}
	return label
}
```

- [ ] **Step 2: Build**

Run: `go build ./ui/`
Expected: no errors. (`titleStyle`, `helpStyle`, `docStyle`, `focusedStyle`, `renderButton` all already exist.)

- [ ] **Step 3: Run all package tests**

Run: `go test ./ui/`
Expected: 4 focus tests PASS.

- [ ] **Step 4: Build the whole app**

Run: `go build -o bakdb .`
Expected: produces `bakdb` binary, no errors.

- [ ] **Step 5: Commit**

```bash
git add ui/views.go
git commit -m "feat(tui): render collapsible advanced options section"
```

---

## Task 5: Manual verification

The TUI needs a terminal; verify by hand.

- [ ] **Step 1: Run the app (no .env, or .env without optional fields)**

Run: `./bakdb`
Pick a DB type with Enter. On the details screen, expected:
- The 5 required fields (Host, Port, Username, Password, Database) are shown.
- A `▸ Advanced options` row appears below them (collapsed).
- The optional fields are NOT visible.
- `Start Backup` button below.

- [ ] **Step 2: Expand/collapse**

Tab down to the `▸ Advanced options` row (it highlights) and press Enter.
Expected: it becomes `▾ Advanced options` and the optional fields (Connection String, Binary Path, Output Directory; plus Backup Format if SQL Server) appear indented. Press Enter again → collapses, fields hidden.

- [ ] **Step 3: Navigation skips hidden fields**

With advanced collapsed, tab from Database.
Expected: focus goes Database → Advanced toggle → Start Backup (does NOT stop on hidden optional fields).

- [ ] **Step 4: Backup still works**

Fill required fields, go to Start Backup, press Enter.
Expected: backing-up screen, then result (success or a clear error). The engine still reads all 9 inputs by index, so hidden-but-filled optional values are still used.

- [ ] **Step 5: Auto-expand from .env (optional check)**

Create a `.env` with `BAKDB_OUTPUT_DIR=~/backups` and run `./bakdb`.
Expected: the advanced section starts expanded (▾) showing the pre-filled Output Directory.

- [ ] **Step 6: Record result**

If all steps pass, the feature is verified.
```
