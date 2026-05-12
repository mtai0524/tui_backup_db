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

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := "\n\n" + m.renderButton("Start Backup", m.focusIndex == len(m.inputs))
	b.WriteString(button)

	b.WriteString(helpStyle.Render("\n\n (tab/shift+tab: move, enter: next/submit, q: quit)"))

	return docStyle.Render(b.String())
}

func (m Model) viewBackingUp() string {
	return docStyle.Render(fmt.Sprintf(
		"%s Backing up %s...\n\n%s",
		m.spinner.View(),
		m.dbType,
		helpStyle.Render("This may take a moment depending on the database size."),
	))
}

func (m Model) viewResult() string {
	var b strings.Builder

	if m.err != nil {
		b.WriteString(errorStyle.Render("✖ Backup Failed"))
		b.WriteString("\n\n")
		b.WriteString(m.err.Error())
	} else {
		b.WriteString(successStyle.Render("✔ Backup Successful!"))
		b.WriteString("\n\n")
		b.WriteString(fmt.Sprintf("File saved to: %s", m.message))
	}

	b.WriteString("\n\n" + helpStyle.Render("Press 'r' to restart or 'q' to quit"))

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
