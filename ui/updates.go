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

			if s == "enter" && m.focusIndex == len(m.inputs) {
				m.state = stateBackingUp
				return m, tea.Batch(m.spinner.Tick, m.startBackupCmd())
			}

			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	cmd := m.updateInputs(msg)
	return m, cmd
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
				m.emailModal = NewEmailModal(m.message, m.defaults)
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
	path string
	err  error
}

func (m Model) startBackupCmd() tea.Cmd {
	return func() tea.Msg {
		cfg := backup.Config{
			Host:       m.inputs[0].Value(),
			Port:       m.inputs[1].Value(),
			User:       m.inputs[2].Value(),
			Password:   m.inputs[3].Value(),
			Database:   m.inputs[4].Value(),
			Type:       m.dbType,
			ConnString: m.inputs[5].Value(),
			BinaryPath: m.inputs[6].Value(),
			OutputDir:  m.inputs[7].Value(), // index 7: output directory
		}
		path, err := backup.ExecuteBackup(cfg)
		return backupFinishedMsg{path: path, err: err}
	}
}
