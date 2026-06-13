package ui

import (
	"bakdb/config"

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
	stateEmail
)

type Model struct {
	state            state
	dbType           string
	list             list.Model
	inputs           []textinput.Model
	focusIndex       int
	advancedExpanded bool // whether the optional-fields section is shown
	spinner          spinner.Model
	err              error
	message          string
	quitting         bool
	emailModal       EmailModal
	defaults         config.Defaults
	databaseName     string // Tên database vừa backup (dùng cho email)
	backupFormat     string // Định dạng backup vừa thực hiện
	catFrame         int    // animation frame counter for the backing-up screen
}

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

func InitialModel() Model {
	defaults := config.Load()

	items := []list.Item{
		item{title: "MySQL", desc: "Backup a MySQL/MariaDB database using mysqldump"},
		item{title: "PostgreSQL", desc: "Backup a PostgreSQL database using pg_dump"},
		item{title: "SQL Server", desc: "Backup a SQL Server database using sqlcmd"},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select Database Type"

	// 9 input fields: index 0-7 như cũ, index 8 là Backup Format (SQL Server only)
	inputs := make([]textinput.Model, 9)
	var t textinput.Model
	for i := range inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 256

		switch i {
		case 0:
			t.Placeholder = "Host (e.g. localhost)"
			t.CharLimit = 128
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "Port (e.g. 3306)"
			t.CharLimit = 8
		case 2:
			t.Placeholder = "Username"
			t.CharLimit = 128
		case 3:
			t.Placeholder = "Password"
			t.CharLimit = 128
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		case 4:
			t.Placeholder = "Database Name"
			t.CharLimit = 128
		case 5:
			t.Placeholder = "Connection String (Optional, overrides fields above)"
		case 6:
			t.Placeholder = "Tool Binary Path (Optional, e.g. C:\\bin\\mysqldump.exe)"
		case 7:
			t.Placeholder = "Output Directory (Optional, default: current directory)"
		case 8:
			t.Placeholder = "Backup Format for SQL Server: .bak or .sql (Optional, default: auto)"
		}

		inputs[i] = t
	}

	applyDefaults(inputs, defaults)

	autoExpand := shouldAutoExpand(inputs)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	m := Model{
		state:            stateSelectDB,
		list:             l,
		inputs:           inputs,
		spinner:          s,
		defaults:         defaults,
		advancedExpanded: autoExpand,
	}

	// Nếu .env chỉ định BAKDB_TYPE hợp lệ, nhảy thẳng vào màn nhập chi tiết
	// với port mặc định phù hợp (nếu user không tự set port).
	if defaults.Type != "" {
		m.dbType = defaults.Type
		m.state = stateEnterDetails
		if defaults.Port == "" {
			switch defaults.Type {
			case "MySQL":
				m.inputs[1].SetValue("3306")
			case "PostgreSQL":
				m.inputs[1].SetValue("5432")
			case "SQL Server":
				m.inputs[1].SetValue("1433")
			}
		}
	}

	return m
}

func applyDefaults(inputs []textinput.Model, d config.Defaults) {
	set := func(idx int, v string) {
		if v != "" {
			inputs[idx].SetValue(v)
		}
	}
	set(0, d.Host)
	set(1, d.Port)
	set(2, d.User)
	set(3, d.Password)
	set(4, d.Database)
	set(5, d.ConnString)
	set(6, d.BinaryPath)
	set(7, d.OutputDir)
	set(8, d.BackupFormat)
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "q":
			// Chỉ quit ở màn hình không có text input để tránh chặn gõ chữ "q"
			if m.state == stateSelectDB || m.state == stateResult {
				m.quitting = true
				return m, tea.Quit
			}
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
	case stateEmail:
		return m.updateEmail(msg)
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
	case stateEmail:
		return m.emailModal.View()
	}

	return "Unknown state"
}
