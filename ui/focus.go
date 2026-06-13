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
