package gui

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func newHistoryView(w fyne.Window, s *AppState, onEmail func()) fyne.CanvasObject {
	dir := s.Defaults.OutputDir
	resolved := resolveDir(dir)

	files, err := listBackups(dir)
	header := widget.NewLabelWithStyle("History — "+resolved, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	if err != nil || len(files) == 0 {
		return container.NewVBox(header, widget.NewLabel("No backup files found in this directory."))
	}

	rows := container.NewVBox()
	for _, fi := range files {
		fi := fi
		full := filepath.Join(resolved, fi.Name())
		label := widget.NewLabel(fmt.Sprintf("%s — %s", fi.Name(), humanSize(fi.Size())))
		sendBtn := widget.NewButton("Send", func() {
			s.LastBackupFile = full
			onEmail()
		})
		openBtn := widget.NewButton("Open", func() { openInFileManager(resolved) })
		rows.Add(container.NewBorder(nil, nil, label, container.NewHBox(sendBtn, openBtn)))
	}

	return container.NewVBox(header, container.NewVScroll(rows))
}

func humanSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
