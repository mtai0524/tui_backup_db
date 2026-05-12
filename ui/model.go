package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type state int

const (
	stateSelectDB state = iota
	stateEnterDetails
	stateBackingUp
	stateResult
)

type Model struct {
	state      state
	dbType     string
	list       list.Model
	inputs     []textinput.Model
	focusIndex int
	spinner    spinner.Model
	err        error
	message    string
	quitting   bool
}

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

func InitialModel() Model {
	// DB Selection List
	items := []list.Item{
		item{title: "MySQL", desc: "Backup a MySQL/MariaDB database using mysqldump"},
		item{title: "PostgreSQL", desc: "Backup a PostgreSQL database using pg_dump"},
		item{title: "SQL Server", desc: "Backup a SQL Server database using sqlcmd"},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select Database Type"

	// Credentials Inputs
	inputs := make([]textinput.Model, 7)
	var t textinput.Model
	for i := range inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 128

		switch i {
		case 0:
			t.Placeholder = "Host (e.g. localhost)"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "Port (e.g. 3306)"
		case 2:
			t.Placeholder = "Username"
		case 3:
			t.Placeholder = "Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		case 4:
			t.Placeholder = "Database Name"
		case 5:
			t.Placeholder = "Connection String (Optional, overrides fields)"
			t.CharLimit = 256
		case 6:
			t.Placeholder = "Tool Binary Path (Optional, e.g. C:\\bin\\mysqldump.exe)"
		}

		inputs[i] = t
	}

	// Spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	return Model{
		state:   stateSelectDB,
		list:    l,
		inputs:  inputs,
		spinner: s,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	switch m.state {
	case stateSelectDB:
		return m.updateSelectDB(msg)
	case stateEnterDetails:
		return m.updateEnterDetails(msg)
	case stateBackingUp:
		return m.updateBackingUp(msg)
	case stateResult:
		return m.updateResult(msg)
	}

	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	switch m.state {
	case stateSelectDB:
		return docStyle.Render(m.list.View())
	case stateEnterDetails:
		return m.viewEnterDetails()
	case stateBackingUp:
		return m.viewBackingUp()
	case stateResult:
		return m.viewResult()
	}

	return "Unknown state"
}
