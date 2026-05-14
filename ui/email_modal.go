package ui

import (
	"bakdb/config"
	"bakdb/email"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	modalBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 3).
			Width(60)

	modalTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("63")).
			MarginBottom(1)

	modalLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	modalSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("42")).
				Bold(true)

	modalErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	modalHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)

	focusedInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("63"))

	blurredInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245"))
)

// ── Field index ───────────────────────────────────────────────────────────────

const (
	fieldFrom = iota
	fieldAppPass
	fieldTo
	fieldSubject
	fieldCount
)

// ── Messages ──────────────────────────────────────────────────────────────────

// EmailSentMsg được gửi về model cha khi gửi mail xong (thành công hoặc lỗi)
type EmailSentMsg struct {
	Err error
}

// emailSendResultMsg dùng nội bộ trong modal
type emailSendResultMsg struct {
	err error
}

// ── Model ─────────────────────────────────────────────────────────────────────

// EmailModal là một sub-model Bubble Tea hiển thị dưới dạng modal.
// Model cha cần:
//  1. Lưu *EmailModal trong state
//  2. Gọi modal.Update(msg) khi modal.Active == true
//  3. Render modal.View() chồng lên giao diện chính
type EmailModal struct {
	Active         bool
	backupFile     string // đường dẫn file backup cần đính kèm
	databaseName   string
	backupFormat   string
	inputs         [fieldCount]textinput.Model
	focusedIdx     int
	sending        bool
	statusMsg      string
	statusIsErr    bool
}

// NewEmailModal khởi tạo modal với đường dẫn file backup, database name, backup format và defaults từ .env.
func NewEmailModal(backupFile string, d config.Defaults, databaseName, backupFormat string) EmailModal {
	m := EmailModal{
		backupFile:   backupFile,
		databaseName: databaseName,
		backupFormat: backupFormat,
		Active:       true,
	}

	placeholders := [fieldCount]string{
		"you@gmail.com",
		"xxxx xxxx xxxx xxxx  (Gmail App Password)",
		"recipient@example.com  (nhiều địa chỉ cách nhau bởi dấu phẩy)",
		"Database Backup  (tuỳ chọn)",
	}
	prefills := [fieldCount]string{
		d.EmailFrom,
		d.EmailAppPassword,
		d.EmailTo,
		d.EmailSubject,
	}

	for i := range m.inputs {
		t := textinput.New()
		t.Placeholder = placeholders[i]
		t.CharLimit = 256
		t.Width = 50
		if i == fieldAppPass {
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}
		if prefills[i] != "" {
			t.SetValue(prefills[i])
		}
		if i == fieldFrom {
			t.Focus()
			t.PromptStyle = focusedInputStyle
			t.TextStyle = focusedInputStyle
		}
		m.inputs[i] = t
	}

	return m
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m EmailModal) Init() tea.Cmd {
	return textinput.Blink
}

// ── Update ────────────────────────────────────────────────────────────────────

func (m EmailModal) Update(msg tea.Msg) (EmailModal, tea.Cmd) {
	if !m.Active {
		return m, nil
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.sending {
			break // chặn input khi đang gửi
		}

		switch msg.String() {
		case "esc":
			m.Active = false
			return m, nil

		case "tab", "down":
			m.nextField()

		case "shift+tab", "up":
			m.prevField()

		case "enter":
			if m.focusedIdx < fieldCount-1 {
				m.nextField()
			} else {
				// Focus ở field cuối → submit
				return m, m.sendEmail()
			}

		case "ctrl+s":
			return m, m.sendEmail()
		}

	case emailSendResultMsg:
		m.sending = false
		if msg.err != nil {
			m.statusMsg = "✖ " + msg.err.Error()
			m.statusIsErr = true
		} else {
			m.statusMsg = "✔ Gửi email thành công!"
			m.statusIsErr = false
		}
		// Thông báo lên model cha
		return m, func() tea.Msg { return EmailSentMsg{Err: msg.err} }
	}

	// Cập nhật tất cả input fields
	for i := range m.inputs {
		var cmd tea.Cmd
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m EmailModal) View() string {
	if !m.Active {
		return ""
	}

	labels := [fieldCount]string{
		"Gmail của bạn   ",
		"App Password    ",
		"Gửi đến         ",
		"Tiêu đề         ",
	}

	var sb strings.Builder

	sb.WriteString(modalTitleStyle.Render("✉  Gửi backup qua Email"))
	sb.WriteString("\n\n")

	sb.WriteString(modalHintStyle.Render(
		fmt.Sprintf("File: %s", m.backupFile),
	))
	sb.WriteString("\n\n")

	for i, input := range m.inputs {
		label := modalLabelStyle.Render(labels[i])
		sb.WriteString(label + "\n")
		sb.WriteString(input.View())
		sb.WriteString("\n\n")
	}

	// Status
	if m.sending {
		sb.WriteString(modalHintStyle.Render("⟳ Đang gửi..."))
	} else if m.statusMsg != "" {
		if m.statusIsErr {
			sb.WriteString(modalErrorStyle.Render(m.statusMsg))
		} else {
			sb.WriteString(modalSuccessStyle.Render(m.statusMsg))
		}
	}

	sb.WriteString("\n\n")
	sb.WriteString(modalHintStyle.Render("Enter/Tab: chuyển field  •  Ctrl+S: gửi  •  Esc: đóng"))

	return modalBoxStyle.Render(sb.String())
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (m *EmailModal) nextField() {
	m.inputs[m.focusedIdx].Blur()
	m.focusedIdx = (m.focusedIdx + 1) % fieldCount
	m.inputs[m.focusedIdx].Focus()
}

func (m *EmailModal) prevField() {
	m.inputs[m.focusedIdx].Blur()
	m.focusedIdx = (m.focusedIdx - 1 + fieldCount) % fieldCount
	m.inputs[m.focusedIdx].Focus()
}

func (m *EmailModal) sendEmail() tea.Cmd {
	from := strings.TrimSpace(m.inputs[fieldFrom].Value())
	appPass := strings.TrimSpace(m.inputs[fieldAppPass].Value())
	toRaw := strings.TrimSpace(m.inputs[fieldTo].Value())
	subject := strings.TrimSpace(m.inputs[fieldSubject].Value())

	if from == "" || appPass == "" || toRaw == "" {
		m.statusMsg = "✖ Vui lòng điền đầy đủ Gmail, App Password và địa chỉ nhận"
		m.statusIsErr = true
		return nil
	}

	// Parse danh sách địa chỉ nhận (cách nhau bởi dấu phẩy hoặc dấu cách)
	var toList []string
	for _, addr := range strings.FieldsFunc(toRaw, func(r rune) bool {
		return r == ',' || r == ';'
	}) {
		addr = strings.TrimSpace(addr)
		if addr != "" {
			toList = append(toList, addr)
		}
	}

	if len(toList) == 0 {
		m.statusMsg = "✖ Địa chỉ email nhận không hợp lệ"
		m.statusIsErr = true
		return nil
	}

	m.sending = true
	m.statusMsg = ""

	// Lấy thông tin file để truyền vào email
	backupFile := m.backupFile
	fileInfo, _ := os.Stat(backupFile)
	fileSize := int64(0)
	if fileInfo != nil {
		fileSize = fileInfo.Size()
	}

	cfg := email.Config{
		FromAddress:    from,
		AppPassword:    appPass,
		ToAddresses:    toList,
		Subject:        subject,
		BackupFileName: filepath.Base(backupFile),
		BackupSize:     fileSize,
		DatabaseName:   m.databaseName,
		BackupFormat:   m.backupFormat,
	}

	return func() tea.Msg {
		err := email.Send(cfg, backupFile)
		return emailSendResultMsg{err: err}
	}
}
