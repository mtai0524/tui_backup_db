package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func newSettingsView(w fyne.Window, s *AppState) fyne.CanvasObject {
	return widget.NewLabel("Settings (todo)")
}
