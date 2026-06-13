package ui

import (
	"bakdb/backup"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// ── Select DB ─────────────────────────────────────────────────────────────────

func (m Model) updateSelectDB(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.dbType = i.title
				m.state = stateEnterDetails
				// Chỉ áp dụng port mặc định khi user chưa nạp port từ .env
				if m.inputs[1].Value() == "" {
					switch m.dbType {
					case "MySQL":
						m.inputs[1].SetValue("3306")
					case "PostgreSQL":
						m.inputs[1].SetValue("5432")
					case "SQL Server":
						m.inputs[1].SetValue("1433")
					}
				}
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// ── Enter Details ─────────────────────────────────────────────────────────────

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

func (m *Model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

// ── Backing Up ────────────────────────────────────────────────────────────────

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

// ── Result ────────────────────────────────────────────────────────────────────

func (m Model) updateResult(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			return InitialModel(), nil
		case "e":
			if m.err == nil && m.message != "" {
				m.emailModal = NewEmailModal(m.message, m.defaults, m.databaseName, m.backupFormat)
				m.state = stateEmail
				return m, textinput.Blink
			}
		}
	}
	return m, nil
}

// ── Email Modal ───────────────────────────────────────────────────────────────

func (m Model) updateEmail(msg tea.Msg) (Model, tea.Cmd) {
	if sent, ok := msg.(EmailSentMsg); ok {
		if sent.Err == nil {
			m.state = stateResult
			m.emailModal.Active = false
		}
		return m, nil
	}

	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
		m.state = stateResult
		m.emailModal.Active = false
		return m, nil
	}

	var cmd tea.Cmd
	m.emailModal, cmd = m.emailModal.Update(msg)
	return m, cmd
}

// ── Backup Command ────────────────────────────────────────────────────────────

type backupFinishedMsg struct {
	path         string
	err          error
	databaseName string
	backupFormat string
}

func (m Model) startBackupCmd() tea.Cmd {
	return func() tea.Msg {
		dbName := m.inputs[4].Value()
		backupFmt := m.inputs[8].Value()
		cfg := backup.Config{
			Host:         m.inputs[0].Value(),
			Port:         m.inputs[1].Value(),
			User:         m.inputs[2].Value(),
			Password:     m.inputs[3].Value(),
			Database:     dbName,
			Type:         m.dbType,
			ConnString:   m.inputs[5].Value(),
			BinaryPath:   m.inputs[6].Value(),
			OutputDir:    m.inputs[7].Value(),
			BackupFormat: backupFmt,
		}
		path, err := backup.ExecuteBackup(cfg)
		return backupFinishedMsg{path: path, err: err, databaseName: dbName, backupFormat: backupFmt}
	}
}
