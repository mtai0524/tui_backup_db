package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"bakdb/config"
)

func newSettingsView(w fyne.Window, s *AppState) fyne.CanvasObject {
	d := s.Defaults

	outDir := widget.NewEntry()
	outDir.SetText(d.OutputDir)
	outDir.SetPlaceHolder("~/backups")
	binPath := widget.NewEntry()
	binPath.SetText(d.BinaryPath)
	binPath.SetPlaceHolder("optional: sqlcmd / mysqldump / pg_dump")

	status := widget.NewLabel("")

	saveBtn := widget.NewButton("💾 Save to .env", func() {
		err := writeEnv(".env", map[string]string{
			"BAKDB_OUTPUT_DIR":  outDir.Text,
			"BAKDB_BINARY_PATH": binPath.Text,
		})
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		s.Defaults = config.Load() // reload so other views see new defaults
		status.SetText("✔ Saved to ./.env")
	})

	form := widget.NewForm(
		widget.NewFormItem("Default output dir", outDir),
		widget.NewFormItem("Binary path", binPath),
	)

	return container.NewVBox(
		widget.NewLabelWithStyle("Settings", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		form,
		saveBtn,
		status,
	)
}
