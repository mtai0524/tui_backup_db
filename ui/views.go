package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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

func (m Model) viewBackingUp() string {
	var b strings.Builder
	b.WriteString(catView(m.catFrame))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("%s Đang backup %s...", m.spinner.View(), m.dbType))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Vui lòng đợi, tùy kích thước database."))
	return docStyle.Render(b.String())
}

func (m Model) viewResult() string {
	var b strings.Builder

	if m.err != nil {
		b.WriteString(errorStyle.Render("✖ Backup Failed"))
		b.WriteString("\n\n")
		b.WriteString(m.err.Error())
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("Press 'r' to restart • 'q' to quit"))
	} else {
		b.WriteString(successStyle.Render("✔ Backup Successful!"))
		b.WriteString("\n\n")
		b.WriteString(fmt.Sprintf("File saved to: %s", m.message))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("Press 'e' to send via email • 'r' to restart • 'q' to quit"))
	}

	return docStyle.Render(b.String())
}

func (m Model) renderButton(text string, focused bool) string {
	s := lipgloss.NewStyle().
		Padding(0, 3).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	if focused {
		s = s.BorderForeground(lipgloss.Color("205")).
			Foreground(lipgloss.Color("205")).
			Bold(true)
	}

	return s.Render(text)
}
