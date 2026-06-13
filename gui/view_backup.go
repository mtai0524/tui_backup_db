package gui

import (
	"os/exec"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"bakdb/backup"
)

func newBackupView(w fyne.Window, s *AppState, onEmail func()) fyne.CanvasObject {
	d := s.Defaults

	dbType := widget.NewSelect([]string{"MySQL", "PostgreSQL", "SQL Server"}, nil)
	if d.Type != "" {
		dbType.SetSelected(d.Type)
	} else {
		dbType.SetSelected("MySQL")
	}

	host := widget.NewEntry()
	host.SetText(d.Host)
	host.SetPlaceHolder("localhost")
	port := widget.NewEntry()
	port.SetText(d.Port)
	user := widget.NewEntry()
	user.SetText(d.User)
	pass := widget.NewPasswordEntry()
	pass.SetText(d.Password)
	database := widget.NewEntry()
	database.SetText(d.Database)
	conn := widget.NewEntry()
	conn.SetText(d.ConnString)
	conn.SetPlaceHolder("optional — overrides host/port/user/password")
	outDir := widget.NewEntry()
	outDir.SetText(d.OutputDir)
	outDir.SetPlaceHolder("~ (home) if empty")

	format := widget.NewRadioGroup([]string{".bak", ".sql"}, nil)
	if d.BackupFormat != "" {
		format.SetSelected(d.BackupFormat)
	}
	formatItem := widget.NewFormItem("Format (SQL Server)", format)

	form := widget.NewForm(
		widget.NewFormItem("DB type", dbType),
		widget.NewFormItem("Host", host),
		widget.NewFormItem("Port", port),
		widget.NewFormItem("User", user),
		widget.NewFormItem("Password", pass),
		widget.NewFormItem("Database *", database),
		widget.NewFormItem("Connection string", conn),
		widget.NewFormItem("Output dir", outDir),
		formatItem,
	)

	status := widget.NewLabel("")
	progress := widget.NewProgressBarInfinite()
	progress.Hide()

	sendBtn := widget.NewButton("Send Email", func() { onEmail() })
	sendBtn.Hide()
	openBtn := widget.NewButton("Open folder", nil)
	openBtn.Hide()

	startBtn := widget.NewButton("⬇ Start Backup", nil)
	startBtn.OnTapped = func() {
		if database.Text == "" {
			status.SetText("⚠ Database name is required")
			return
		}
		cfg := backupForm{
			Type: dbType.Selected, Host: host.Text, Port: port.Text,
			User: user.Text, Password: pass.Text, Database: database.Text,
			ConnString: conn.Text, OutputDir: outDir.Text, BackupFormat: format.Selected,
		}.toConfig()

		startBtn.Disable()
		sendBtn.Hide()
		openBtn.Hide()
		progress.Show()
		status.SetText("⏳ Running backup...")

		go func() {
			path, err := backup.ExecuteBackup(cfg)
			fyne.Do(func() {
				progress.Hide()
				startBtn.Enable()
				if err != nil {
					status.SetText("✖ Backup failed: " + err.Error())
					return
				}
				s.LastBackupFile = path
				status.SetText("✔ Backup successful: " + path)
				sendBtn.Show()
				openBtn.Show()
				openBtn.OnTapped = func() { openInFileManager(filepath.Dir(path)) }
			})
		}()
	}

	return container.NewVBox(
		widget.NewLabelWithStyle("Backup", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		form,
		startBtn,
		progress,
		status,
		container.NewHBox(sendBtn, openBtn),
	)
}

// openInFileManager opens dir in the OS file manager.
func openInFileManager(dir string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", dir)
	case "windows":
		cmd = exec.Command("explorer", dir)
	default:
		cmd = exec.Command("xdg-open", dir)
	}
	_ = cmd.Start()
}
