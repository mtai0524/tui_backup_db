# TUI: Collapsible "Advanced options" — Design

**Date:** 2026-06-13
**Status:** Approved (pending spec review)

## Goal

Make the TUI backup form easier to use by hiding the optional fields behind a collapsible **"Advanced options"** row. By default the form shows only the 5 required connection fields; the 4 optional fields appear when the user expands the advanced section. This reduces visual clutter for new users without removing access for power users.

## Context

`bakdb` reverted to the Bubble Tea TUI (`main.go` launches `ui.InitialModel()`). The engine (`backup/`, `email/`, `config/`) is unchanged. Only the **details-entry screen** (`stateEnterDetails`) changes. All other screens (select DB, backing up, result, email modal) stay exactly as they are.

Current form (`ui/views.go` `viewEnterDetails`, `ui/model.go`) renders all 9 `textinput.Model` fields in a flat vertical list, navigated by `focusIndex` over inputs 0..8 plus a button. Index map today:

- 0 Host, 1 Port, 2 Username, 3 Password, 4 Database (REQUIRED)
- 5 Connection String, 6 Binary Path, 7 Output Directory, 8 Backup Format (OPTIONAL)

The Backup Format field (8) is already shown only for SQL Server (`displayCount` is 9 for SQL Server, else 8).

## Decisions

- Required fields shown always: Host, Port, Username, Password, Database.
- Optional fields (Connection String, Binary Path, Output Directory, Backup Format) live under a collapsible row.
- The collapse toggle is a focusable row rendered as `▸ Advanced options` (collapsed) / `▾ Advanced options` (expanded). Pressing **Enter** on it toggles expansion.
- Backup Format (index 8) only appears in the expanded advanced section **when** `dbType == "SQL Server"` (preserve existing behavior).
- **Auto-expand on startup** if `.env` pre-filled any optional field (5,6,7, or 8). Otherwise start collapsed.
- Everything else (colors, other screens) unchanged.

## Approach

Keep the existing 9 `textinput.Model` slice and index map unchanged — the engine wiring and `applyDefaults` already depend on it. Add UI state for the collapse and route navigation through a computed list of "focusable stops" rather than a flat integer range.

### New model state (`ui/model.go`)

Add one field to `Model`:
```go
advancedExpanded bool // whether the optional-fields section is shown
```

Set it in `InitialModel()` after `applyDefaults`: expand if any of inputs 5,6,7,8 has a non-empty value.

### Focus order as a computed sequence (`ui/model.go` or `ui/updates.go`)

Replace the implicit `focusIndex 0..displayCount` walk with a helper that returns the ordered list of focus stops given current state. A stop is either an input index, the advanced-toggle, or the button. Define stop identifiers:

```go
const (
    stopAdvancedToggle = -1 // the "▸/▾ Advanced options" row
    stopButton         = -2 // the "Start Backup" button
)

// focusStops returns the ordered focusable stops for the current form state.
func (m Model) focusStops() []int {
    stops := []int{0, 1, 2, 3, 4} // required fields
    stops = append(stops, stopAdvancedToggle)
    if m.advancedExpanded {
        stops = append(stops, 5, 6, 7)
        if m.dbType == "SQL Server" {
            stops = append(stops, 8)
        }
    }
    stops = append(stops, stopButton)
    return stops
}
```

`focusIndex` becomes an index INTO `focusStops()` (0..len-1), not a raw input index. A small accessor `currentStop()` returns `m.focusStops()[m.focusIndex]`.

### Navigation (`ui/updates.go` `updateEnterDetails`)

- **tab / down / enter**: move focus to next stop (wrap or clamp consistent with current behavior). When the current stop is `stopAdvancedToggle` and the key is **enter**, toggle `advancedExpanded` instead of moving, then clamp `focusIndex` so it stays valid (the stop list length changes). When the current stop is `stopButton` and key is enter, submit (start backup) — same as today.
- **shift+tab / up**: move to previous stop.
- Only the input whose index == current stop gets `Focus()`; all others `Blur()`. The toggle and button are not text inputs, so when they're focused, no textinput is focused.
- When toggling collapse, if focus was on a now-hidden input, it can't be (toggle is only reachable from the toggle row), so set `focusIndex` to the toggle row's position after toggling.

### Rendering (`ui/views.go` `viewEnterDetails`)

Render in this order, each line labeled:

1. Title (unchanged).
2. The 5 required inputs, each on its own line.
3. The advanced toggle row: `▸ Advanced options` or `▾ Advanced options`. When it is the focused stop, render with `focusedStyle` (highlighted); otherwise normal.
4. If expanded: inputs 5, 6, 7, and (SQL Server only) 8, indented two spaces to show grouping.
5. The Start Backup button (`renderButton`, focused when current stop == stopButton).
6. Help line (unchanged text, still valid).

A helper renders the toggle row so focus styling is consistent:
```go
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

## Data Flow

No change to backup execution. On submit, the engine still reads inputs 0..8 by index; hidden optional inputs simply keep their current (possibly empty) values, exactly as before. Collapsing does not clear values — it only hides them from view and from the focus walk.

## Error Handling

Unchanged. Submit validation and the result screen are untouched.

## Edge Cases

- **Collapse while focused inside advanced:** not possible — the toggle is the only stop that collapses, and it sits above the advanced inputs. After toggling, focus stays on the toggle row.
- **Switch DB type to/from SQL Server while expanded:** `focusStops()` recomputes, so Backup Format appears/disappears correctly. If `focusIndex` would exceed the new length, clamp to last stop. (DB type is chosen on the previous screen, so this is rare, but the clamp keeps it safe.)
- **`.env` sets only Backup Format on a non-SQL-Server type:** auto-expand still triggers (a value exists), but field 8 won't render until SQL Server is selected. Acceptable — the value is preserved and the section is open.

## Testing

The view is rendered text; logic worth testing is the focus-stop computation (pure):

- `focusStops()` returns `[0,1,2,3,4, stopAdvancedToggle, stopButton]` when collapsed.
- Expanded, non-SQL-Server: `[0,1,2,3,4, toggle, 5,6,7, button]`.
- Expanded, SQL Server: `[0,1,2,3,4, toggle, 5,6,7,8, button]`.
- Auto-expand helper: returns true iff any of inputs 5–8 has a non-empty value.

These are unit tests on small pure helpers in `ui/` (new `ui/focus_test.go`). Manual check: run the TUI, confirm collapsed-by-default, Enter expands/collapses, tab skips hidden fields, backup still works.

## Out of Scope (YAGNI)

- No restyling of colors, no changes to other screens, no new keybindings beyond Enter-on-toggle, no field reordering.
```
